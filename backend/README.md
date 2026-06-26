# Linkr — Backend

Go API server for Linkr. See the [root README](../README.md) for full-stack setup instructions.

## Running

```bash
# from repo root
task run

# or from this directory
go run ./cmd/api/
```

Requires `DATABASE_URL` and `JWT_SECRET` (32+ chars) set in `.env`. Copy `.env.example` to get started. Migrations run automatically on startup.

## Testing

```bash
# from repo root (recommended — includes race detector)
task test

# or from this directory
go test -race ./...
```

## Building

```bash
task build
# produces ./bin/linkr
```

## Swagger

```bash
task swag   # regenerates docs/ from godoc annotations
```

Docs served at `http://localhost:8080/swagger/index.html` while the server is running.

## Package layout

```
cmd/api/          Entry point, dependency wiring, graceful shutdown
internal/
  auth/           JWT sign / verify
  clicks/         Async pipeline: buffered channel → batch CopyFrom → Postgres
  config/         Typed env config with defaults
  domain/         Core types: Link, ClickEvent, LinkStats, OverviewStats
  http/
    handler/      Thin HTTP translators (bind → usecase → JSON)
    middleware/   RequestID, logger, recovery, Prometheus metrics, JWT, rate limit
  repository/     pgx queries (link, click, user repos)
  service/        Tiered cache: L1 expirable LRU + L2 Redis + gobreaker circuit breaker
  shortcode/      Secure random base62 code generation
  ua/             User-agent and referrer parsing (no external DB)
  usecase/        Business logic: create, list, stats, auth
migrations/       SQL migrations managed by golang-migrate
```
