# Linkr — URL Shortener with Analytics

A full-stack URL shortener built with Go (Gin) and Next.js. Create short links, track clicks, and view per-link analytics broken down by day, device, browser, and referrer.

---

## Prerequisites

| Tool | Version |
|---|---|
| Docker + Compose | any recent |
| Go | 1.26+ *(local dev only)* |
| Node.js | 20+ *(local dev only)* |
| [Task](https://taskfile.dev) | any *(optional convenience runner)* |

---

## Quick Start — Docker (recommended)

Runs the full stack (Postgres, Redis, Go backend, Next.js frontend) with a single command.

```bash
# 1. Create your env file
cp .env.example .env
#    Edit .env and set JWT_SECRET to a random string of 32+ characters

# 2. Build images and start all services
docker compose up --build
```

- Dashboard: `http://localhost:3000`
- API / Swagger: `http://localhost:8080/swagger/index.html`

To stop: `docker compose down`. To also wipe the database volume: `docker compose down -v`.

---

## Quick Start — Local Dev

### 1. Start Postgres + Redis

```bash
docker compose up -d postgres redis
```

### 2. Configure the backend

```bash
cp backend/.env.example backend/.env
# Edit backend/.env — set JWT_SECRET (32+ chars) and optionally REDIS_URL
```

### 3. Run the backend

Migrations run automatically on startup.

```bash
task run          # or: cd backend && go run ./cmd/api/
```

API available at `http://localhost:8080`. Swagger UI at `http://localhost:8080/swagger/index.html`.

### 4. Configure the frontend

```bash
cp frontend/.env.example frontend/.env.local
```

### 5. Run the frontend

```bash
task web          # or: cd frontend && npm install && npm run dev
```

Dashboard at `http://localhost:3000`.

---

## Running Tests

```bash
# via Task (runs with race detector)
task test

# or manually
cd backend && go test -race ./...
```

Tests cover: short-code generation, URL validation, JWT auth, async click pipeline (flush on batch size, flush on ticker, drop on full buffer, drain on shutdown, error resilience, concurrent enqueue under race detector), link domain logic, and user-agent parsing.

---

## Project Structure

```
.
├── backend/                  Go API server
│   ├── cmd/api/              Entry point (main.go)
│   ├── internal/
│   │   ├── clicks/           Async click pipeline
│   │   ├── config/           Environment config
│   │   ├── domain/           Core types (Link, ClickEvent, Stats)
│   │   ├── http/             Gin router, handlers, middleware
│   │   ├── repository/       PostgreSQL queries (pgx)
│   │   ├── service/          Tiered cache (LRU + Redis + circuit breaker)
│   │   ├── shortcode/        Short-code generation
│   │   └── usecase/          Business logic
│   └── migrations/           SQL migrations (golang-migrate)
├── frontend/                 Next.js 15 app
│   ├── app/                  App Router pages
│   └── components/           UI components
├── docker-compose.yml        Full-stack: Postgres + Redis + backend + frontend
├── Taskfile.yml              Task runner shortcuts
└── DECISIONS.md              Architecture decisions and trade-offs
```

---

## Environment Variables

### Backend (`backend/.env`)

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `JWT_SECRET` | *(required, 32+ chars)* | HMAC secret for JWT signing |
| `PORT` | `8080` | HTTP listen port |
| `REDIS_URL` | *(empty — disables Redis)* | Redis connection URL |
| `CACHE_SIZE` | `10000` | L1 LRU max entries |
| `L1_CACHE_TTL_SEC` | `30` | L1 TTL in seconds |
| `REDIS_CACHE_TTL_SEC` | `300` | Redis TTL in seconds |
| `CLICK_BUFFER_SIZE` | `10000` | Click pipeline channel capacity |
| `CLICK_BATCH_SIZE` | `500` | Events flushed per DB write |
| `CLICK_FLUSH_INTERVAL_MS` | `200` | Max ms between flushes |
| `CLICK_WORKERS` | `4` | Concurrent flush workers |
| `DB_MAX_CONNS` | `25` | pgxpool max connections |
| `DB_MIN_CONNS` | `5` | pgxpool min connections |
| `RATE_LIMIT_RPS` | `100` | Per-IP requests per second |
| `RATE_LIMIT_BURST` | `200` | Per-IP burst allowance |
| `CACHE_CONTROL_MAX_AGE_SEC` | `60` | `Cache-Control` max-age on redirects |

### Frontend (`frontend/.env.local`)

| Variable | Default | Description |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend base URL |

---

## API Overview

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/auth/register` | — | Create account |
| `POST` | `/api/auth/login` | — | Get JWT |
| `GET` | `/api/auth/me` | JWT | Current user |
| `POST` | `/api/links` | JWT | Create short link |
| `GET` | `/api/links` | JWT | List links (cursor pagination) |
| `PATCH` | `/api/links/:id` | JWT | Toggle active / set expiry |
| `DELETE` | `/api/links/:id` | JWT | Soft-delete link |
| `GET` | `/api/links/:code/stats` | JWT | Click analytics |
| `GET` | `/api/analytics/overview` | JWT | Aggregate stats across all links |
| `GET` | `/:code` | — | Redirect (hot path) |
| `GET` | `/metrics` | — | Prometheus metrics |
| `GET` | `/healthz` | — | Liveness probe |
| `GET` | `/readyz` | — | Readiness probe (pings DB) |

Full interactive docs: `http://localhost:8080/swagger/index.html`
