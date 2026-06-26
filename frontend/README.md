# Linkr — Frontend

Next.js 16 dashboard for Linkr. See the [root README](../README.md) for full-stack setup instructions.

## Running

```bash
# install dependencies (first time only)
npm install

# start dev server
npm run dev         # http://localhost:3000

# or via Task from repo root
task web
```

## Other commands

```bash
npm run build       # production build
npm run start       # serve the production build
npm run lint        # ESLint
```

## Environment variables

Copy `.env.example` to `.env.local` and adjust if needed:

```bash
cp .env.example .env.local
```

| Variable | Default | Description |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend base URL (used client-side for redirects) |
| `API_URL` | `http://localhost:8080` | Backend base URL (used server-side in Route Handlers) |
| `JWT_COOKIE_NAME` | `linkr_token` | Name of the HTTP-only JWT cookie |

## Page structure

```
app/
  page.tsx                        Root — redirects to /dashboard or /login
  (auth)/
    login/page.tsx                Login form
    register/page.tsx             Register form
  (protected)/
    layout.tsx                    Auth guard + sidebar shell
    dashboard/page.tsx            Link list with infinite scroll
    analytics/page.tsx            Aggregate stats across all links
    links/[code]/page.tsx         Per-link stats (charts, breakdowns)
```

## Key components

| Component | Purpose |
|---|---|
| `LinkTable` | Infinite-scroll link list, create dialog, search, copy/navigate actions |
| `CreateLinkForm` | Controlled form with Zod validation, expiry picker, alias field |
| `StatsContent` | Per-link analytics: daily area chart, cumulative growth, day-of-week bar chart, device/browser donuts, referrer list |
| `DashboardStats` | Aggregate overview cards and charts (analytics page) |
| `ClicksChart` | Reusable Recharts area chart for daily click data |
| `DonutChart` | Reusable Recharts pie chart for categorical breakdowns |
| `Sidebar` | Bitly-style left nav with route links and logout |
