# DECISIONS.md

---

## 14. The 3–4 most significant design decisions

**Tiered cache on the redirect path.** The redirect endpoint is the only path that sees real scale. I put an in-process expirable LRU (30s TTL) in front of Redis, backed by Postgres. The LRU is expirable rather than plain LRU because a deleted link would otherwise serve indefinitely from warm replicas. Every Redis call is wrapped in a `gobreaker` circuit breaker — five consecutive failures opens it and requests fall through to Postgres directly, no goroutine pile-up or timeout cascade. The accepted cost is up to 30s of staleness on a delete or update. For a URL shortener that's fine.

**Gin and raw SQL with pgx — no ORM.** I chose Gin over stdlib or lighter routers because the project needed middleware chaining (request ID, logger, recovery, Prometheus, JWT, rate limiting), struct binding with `go-playground/validator`, and Swagger integration. Gin handles all of that with a mature API. Chi or stdlib would have worked but meant writing more glue code around things Gin already provides.

For the database layer I chose raw SQL with pgx directly rather than an ORM (GORM, Ent). Three things drove that: `CopyFrom` for the click pipeline bulk-inserts (a Postgres wire-protocol feature most ORMs don't expose cleanly), keyset pagination using `WHERE (created_at, id) < ($1, $2)` which would be awkward in GORM, and the redirect hot path where I wanted zero magic and minimal allocations. pgx v5's `pgx.CollectOneRow` with `pgx.RowToStructByName` handles row scanning cleanly without code generation. The schema is small enough that writing SQL by hand is not a burden.

**Async click pipeline with deliberate drop.** `Enqueue` is a non-blocking channel send. Workers accumulate events and batch-flush to Postgres via `CopyFrom`. If the buffer fills, new events are dropped — the redirect returns before any write I/O happens. A blocked redirect is worse than a missed click. On `SIGTERM` the pipeline drains cleanly. On `SIGKILL`, whatever is buffered is lost — acceptable, since clicks are analytics, not financial data.

**Redis rate limiter with per-instance fallback.** I shared one Redis connection pool between the cache and the rate limiter. When Redis is healthy, limiting is global across replicas. When Redis is down, each replica falls back to its own token bucket, so the effective limit becomes `rps × replica_count`. Permissive rate limiting during a Redis outage beats rejecting all traffic.



---

## 15. Heavy traffic and concurrent load

The redirect path is stateless and fully cacheable. L1 LRU serves warm links from memory with no network I/O. A `singleflight` group on Postgres reads collapses N concurrent cold misses for the same code into one query. The async pipeline means a viral spike never back-pressures the redirect. `Cache-Control: public, max-age=60` is already set — adding a CDN requires zero code changes and removes most origin load at the edge.

Redis is already in the stack — not a future scaling step. Per-instance L1 cache works fine for a single node but breaks down the moment you run multiple replicas: each instance warms its own cache independently and a Postgres cold miss on one replica isn't visible to the others. Redis acts as the shared L2 that all replicas read from and write to, so the cache is coherent across the fleet. I included it from the start because horizontal scaling was a design goal, not an afterthought.

**What breaks first:** Postgres. All writes — link creation and click batch flushes — go to one node. That's the genuine first bottleneck under sustained write-heavy load. Second is the click channel saturating and dropping events, which is by design but at some point you'd want to observe the drop rate in metrics. Third, rate limiting degrades gracefully but becomes permissive during a Redis outage.

**How to scale further:** CDN first — the headers are already set, it's purely an infrastructure change. Then a read replica to offload stats queries from the write primary. Long term, replace the in-process click buffer with Redis Streams for crash durability and independent consumer scaling.

---

## 16. Async click recording under load and on crash

`Pipeline.Enqueue` does a non-blocking channel send and returns immediately. The redirect handler is done before any I/O starts. Workers flush when the batch hits 500 events or every 200ms, whichever comes first, using pgx `CopyFrom` (Postgres COPY wire protocol — significantly faster than row-by-row inserts at volume).

On `SIGTERM`: `Stop()` closes the channel. Workers' `for range` drains remaining events and flushes partial batches before the `WaitGroup` releases. No loss.

On `SIGKILL` or OOM: whatever is in the channel plus any mid-flush batch is lost. The loss is bounded by `bufferSize × flushInterval` — typically a few seconds of traffic. This is acceptable because click counts aren't reconciled against anything financial. If zero-loss were a hard requirement, the right answer is Redis Streams with consumer offset checkpointing, not a bigger in-process buffer.

---

## 17. What I'd do differently with another week

**Durable click pipeline.** The in-process buffered channel is the most architecturally fragile piece. Redis Streams would give crash durability and let flush workers scale independently of API servers — a meaningful improvement for modest added complexity.

**Redis pub/sub invalidation.** A delete evicts from Redis immediately but L1 on other replicas coasts on its 30s TTL. A pub/sub broadcast would make invalidation instant across all replicas.

**E2E tests.** The create → redirect → stats flow has no Playwright coverage. The async flush delay — a click lands in the DB up to 200ms after the redirect — is exactly the kind of timing gap that unit tests can't catch.

**Real-time-ish stats updates.** Click counts on the dashboard only refresh when the user manually refreshes or returns to the tab. For a link that's actively getting traffic, that's a poor experience. A short-interval poll on a dedicated `GET /api/stats` endpoint (or SSE) would give live-ish feedback without requiring a full page reload. The async pipeline already has up to 200ms of flush delay on top of that, so truly real-time isn't achievable without architectural changes, but "updates every few seconds" is realistic with minimal extra work.

---

## 18. AI tooling — where I used it and where I pushed back

I used Claude Code throughout: scaffolding components, the sidebar layout refactor, TanStack Query integration, and reviewing the backend architecture (tiered cache, circuit breaker, rate limiting).

**Override — plain LRU to expirable LRU.** The initial cache scaffolding used a plain `golang-lru`. I switched it to `hashicorp/golang-lru/v2/expirable`. A plain LRU only evicts on capacity — a deleted link would stay in L1 indefinitely on any replica that had it warm, because nothing would push it out. The expirable variant adds TTL-based eviction, so a deletion is guaranteed to be visible across all replicas within 30 seconds without any invalidation broadcast. That's the only reason for the swap; the rest of the API is identical.

**Override — polling the paginated list.** Claude suggested `refetchInterval: 30_000` on the `useInfiniteQuery` backing the dashboard. I removed it. Polling an infinite-scroll list fires one request per loaded page every 30 seconds per active user, and produces inconsistent aggregate counts when not all pages are loaded. I switched to `staleTime: 30_000` with `refetchOnWindowFocus` — data stays fresh for 30 seconds, refetches on tab return, no background polling.
