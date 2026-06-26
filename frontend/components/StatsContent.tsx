"use client"

import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts"
import type { LinkStats } from "@/lib/types"
import ClicksChart from "@/components/ClicksChart"
import BreakdownList from "@/components/BreakdownList"
import DonutChart from "@/components/DonutChart"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { RefreshCw } from "lucide-react"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"
const DOW_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]

const DEVICE_ICONS: Record<string, string> = {
  desktop: "🖥️",
  mobile: "📱",
  tablet: "📟",
}

function fmtNum(n: number): string {
  return n.toLocaleString(undefined, { maximumFractionDigits: 1 })
}

interface Props {
  code: string
  initialStats?: LinkStats
}

export default function StatsContent({ code, initialStats }: Props) {
  const {
    data: stats,
    isLoading,
    isError,
    error,
    refetch,
    isRefetching,
  } = useQuery<LinkStats>({
    queryKey: ["stats", code],
    queryFn: async () => {
      const res = await fetch(`/api/links-proxy/${code}/stats`)
      if (res.status === 401) {
        window.location.replace("/api/auth/logout")
        throw new Error("Session expired")
      }
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw Object.assign(new Error(body.error ?? "Failed to fetch stats"), {
          status: res.status,
        })
      }
      return res.json()
    },
    initialData: initialStats,
    staleTime: 30_000,
  })

  const daily = stats?.daily ?? []

  const avgClicksPerDay = useMemo(() => {
    if (daily.length === 0) return 0
    const total = daily.reduce((s, d) => s + d.count, 0)
    return total / daily.length
  }, [daily])

  const peakDay = useMemo(() => {
    if (daily.length === 0) return null
    return daily.reduce((best, d) => (d.count > best.count ? d : best))
  }, [daily])

  const dowData = useMemo(() => {
    const counts = Array(7).fill(0) as number[]
    daily.forEach((d) => {
      const dow = new Date(d.day + "T00:00:00").getDay()
      counts[dow] += d.count
    })
    return DOW_LABELS.map((name, i) => ({ name, clicks: counts[i] }))
  }, [daily])

  const cumulativeData = useMemo(() => {
    let sum = 0
    return daily.map((d) => {
      sum += d.count
      return { day: d.day.slice(5), total: sum }
    })
  }, [daily])

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-24 bg-muted rounded-xl animate-pulse" />
          ))}
        </div>
        <div className="h-56 bg-muted rounded-xl animate-pulse" />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="h-52 bg-muted rounded-xl animate-pulse" />
          <div className="h-52 bg-muted rounded-xl animate-pulse" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="h-64 bg-muted rounded-xl animate-pulse" />
          <div className="h-64 bg-muted rounded-xl animate-pulse" />
        </div>
        <div className="h-48 bg-muted rounded-xl animate-pulse" />
      </div>
    )
  }

  if (isError) {
    const status = (error as { status?: number }).status
    if (status === 404) return <p className="text-muted-foreground">Link not found.</p>
    if (status === 403) return <p className="text-muted-foreground">You don&apos;t own this link.</p>
    return <p className="text-destructive">Failed to load stats. Is the API running?</p>
  }

  const deviceItems = (stats?.devices ?? []).map((d) => ({ name: d.device, count: d.count }))
  const browserItems = (stats?.browsers ?? []).map((b) => ({ name: b.browser, count: b.count }))
  const refererItems = (stats?.referers ?? []).map((r) => ({ name: r.domain, count: r.count }))

  return (
    <div className="space-y-6">
      {/* Four summary stat cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground mb-1">Short link</p>
            <a
              href={`${API_BASE}/${code}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary font-mono hover:underline break-all"
            >
              /{code}
            </a>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground mb-1">Total clicks</p>
            <p className="text-3xl font-bold">{stats?.total_clicks.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground mb-1">Avg / day</p>
            <p className="text-3xl font-bold">{fmtNum(avgClicksPerDay)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground mb-1">Peak day</p>
            {peakDay ? (
              <>
                <p className="text-2xl font-bold">{peakDay.count.toLocaleString()}</p>
                <p className="text-xs text-muted-foreground mt-0.5">{peakDay.day}</p>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">No data yet</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Daily clicks area chart */}
      <Card>
        <CardContent className="pt-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-medium text-muted-foreground">Clicks over time</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => refetch()}
              disabled={isRefetching}
              aria-label="Refresh stats"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${isRefetching ? "animate-spin" : ""}`} />
            </Button>
          </div>
          <ClicksChart daily={stats?.daily ?? []} />
        </CardContent>
      </Card>

      {/* Cumulative growth + Day-of-week pattern */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm font-medium text-muted-foreground mb-4">Cumulative growth</p>
            {cumulativeData.length === 0 ? (
              <p className="text-sm text-muted-foreground py-10 text-center">No data yet</p>
            ) : (
              <ResponsiveContainer width="100%" height={185}>
                <AreaChart data={cumulativeData}>
                  <defs>
                    <linearGradient id="gradCumulative" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#e11d48" stopOpacity={0.2} />
                      <stop offset="95%" stopColor="#e11d48" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--muted))" />
                  <XAxis
                    dataKey="day"
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={40}
                  />
                  <Tooltip
                    contentStyle={{ fontSize: 12 }}
                    formatter={(v) => [(v as number).toLocaleString(), "Total clicks"]}
                  />
                  <Area
                    type="monotone"
                    dataKey="total"
                    stroke="#e11d48"
                    strokeWidth={2}
                    fill="url(#gradCumulative)"
                    dot={false}
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-5">
            <p className="text-sm font-medium text-muted-foreground mb-4">Clicks by day of week</p>
            {daily.length === 0 ? (
              <p className="text-sm text-muted-foreground py-10 text-center">No data yet</p>
            ) : (
              <ResponsiveContainer width="100%" height={185}>
                <BarChart data={dowData} barCategoryGap="30%">
                  <CartesianGrid
                    strokeDasharray="3 3"
                    stroke="hsl(var(--muted))"
                    vertical={false}
                  />
                  <XAxis
                    dataKey="name"
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    tick={{ fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={36}
                    allowDecimals={false}
                  />
                  <Tooltip
                    contentStyle={{ fontSize: 12 }}
                    formatter={(v) => [(v as number).toLocaleString(), "Clicks"]}
                  />
                  <Bar dataKey="clicks" fill="#e11d48" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Device + Browser donut charts */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <DonutChart
          title="Device breakdown"
          items={deviceItems}
          formatName={(name) => `${DEVICE_ICONS[name] ?? ""} ${name}`.trim()}
        />
        <DonutChart title="Browser breakdown" items={browserItems} />
      </div>

      {/* Top referrers ranked list */}
      <BreakdownList
        title="Top referrers"
        items={refererItems}
        formatName={(name) => (name === "direct" ? "Direct / None" : name)}
      />

      {/* Privacy notice */}
      <p className="text-xs text-muted-foreground text-center">
        Analytics are aggregated and anonymised. Raw visitor signals (IP addresses, user agents) are
        never stored.
      </p>
    </div>
  )
}
