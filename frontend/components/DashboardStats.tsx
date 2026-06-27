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
  Cell,
} from "recharts"
import type { OverviewStats } from "@/lib/types"
import BreakdownList from "@/components/BreakdownList"
import DonutChart from "@/components/DonutChart"
import { Card, CardContent } from "@/components/ui/card"

const DOW_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
const TOP_LINK_COLORS = [1, 0.86, 0.72, 0.58, 0.44, 0.34, 0.26, 0.2, 0.15, 0.1]

function fmtNum(n: number): string {
  return n.toLocaleString(undefined, { maximumFractionDigits: 1 })
}

interface Props {
  initialStats?: OverviewStats
}

export default function DashboardStats({ initialStats }: Props) {
  const {
    data: stats,
    isLoading,
    isError,
  } = useQuery<OverviewStats>({
    queryKey: ["analytics-overview"],
    queryFn: async () => {
      const res = await fetch("/api/analytics-overview")
      if (res.status === 401) {
        window.location.replace("/api/auth/logout")
        throw new Error("Session expired")
      }
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw Object.assign(new Error(body.error ?? "Failed to fetch overview"), {
          status: res.status,
        })
      }
      return res.json()
    },
    initialData: initialStats,
    staleTime: 60_000,
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

  const cumulativeData = useMemo(
    () =>
      daily.reduce<{ day: string; total: number }[]>((acc, d) => {
        const prev = acc[acc.length - 1]?.total ?? 0
        return [...acc, { day: d.day.slice(5), total: prev + d.count }]
      }, []),
    [daily]
  )

  const topLinksData = useMemo(
    () =>
      (stats?.top_links ?? []).map((l) => ({
        name: `/${l.short_code}`,
        clicks: l.total_clicks,
      })),
    [stats?.top_links]
  )

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-24 bg-muted rounded-xl animate-pulse" />
          ))}
        </div>
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-56 bg-muted rounded-xl animate-pulse" />
        ))}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-56 bg-muted rounded-xl animate-pulse" />
          ))}
        </div>
        <div className="h-48 bg-muted rounded-xl animate-pulse" />
      </div>
    )
  }

  if (isError) {
    return <p className="text-destructive">Failed to load analytics. Is the API running?</p>
  }

  const deviceItems = (stats?.devices ?? []).map((d) => ({ name: d.device, count: d.count }))
  const browserItems = (stats?.browsers ?? []).map((b) => ({ name: b.browser, count: b.count }))
  const refererItems = (stats?.referers ?? []).map((r) => ({ name: r.domain, count: r.count }))
  const hasClicks = (stats?.total_clicks ?? 0) > 0

  return (
    <div className="space-y-6">
      {/* Summary stat cards */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Links</p>
            <p className="text-3xl font-bold">{stats?.total_links.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Active</p>
            <p className="text-3xl font-bold">{stats?.active_links.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Total clicks</p>
            <p className="text-3xl font-bold">{stats?.total_clicks.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Avg / day</p>
            <p className="text-3xl font-bold">{fmtNum(avgClicksPerDay)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Peak day</p>
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
      {hasClicks && (
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm font-medium text-muted-foreground mb-4">Clicks over time (all links)</p>
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={daily.map((d) => ({ day: d.day.slice(5), clicks: d.count }))}>
                <defs>
                  <linearGradient id="gradDaily" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#e11d48" stopOpacity={0.2} />
                    <stop offset="95%" stopColor="#e11d48" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--muted))" />
                <XAxis dataKey="day" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} interval="preserveStartEnd" />
                <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} width={40} allowDecimals={false} />
                <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v) => [(v as number).toLocaleString(), "Clicks"]} />
                <Area type="monotone" dataKey="clicks" stroke="#e11d48" strokeWidth={2} fill="url(#gradDaily)" dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}

      {/* Cumulative growth + Day-of-week pattern */}
      {hasClicks && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Card>
            <CardContent className="pt-5">
              <p className="text-sm font-medium text-muted-foreground mb-4">Cumulative growth</p>
              <ResponsiveContainer width="100%" height={185}>
                <AreaChart data={cumulativeData}>
                  <defs>
                    <linearGradient id="gradCumulativeOverview" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#e11d48" stopOpacity={0.2} />
                      <stop offset="95%" stopColor="#e11d48" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--muted))" />
                  <XAxis dataKey="day" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} interval="preserveStartEnd" />
                  <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} width={40} />
                  <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v) => [(v as number).toLocaleString(), "Total"]} />
                  <Area type="monotone" dataKey="total" stroke="#e11d48" strokeWidth={2} fill="url(#gradCumulativeOverview)" dot={false} />
                </AreaChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-5">
              <p className="text-sm font-medium text-muted-foreground mb-4">Clicks by day of week</p>
              <ResponsiveContainer width="100%" height={185}>
                <BarChart data={dowData} barCategoryGap="30%">
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--muted))" vertical={false} />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} />
                  <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} width={36} allowDecimals={false} />
                  <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v) => [(v as number).toLocaleString(), "Clicks"]} />
                  <Bar dataKey="clicks" fill="#e11d48" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Top links by clicks (horizontal bar) */}
      {topLinksData.length > 0 && (
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm font-medium text-muted-foreground mb-4">Top links by clicks</p>
            <ResponsiveContainer width="100%" height={topLinksData.length * 44 + 16}>
              <BarChart
                data={topLinksData}
                layout="vertical"
                margin={{ top: 0, right: 24, left: 0, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="hsl(var(--border))" />
                <XAxis type="number" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                <YAxis type="category" dataKey="name" width={90} tick={{ fontSize: 12 }} tickLine={false} axisLine={false} />
                <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v) => [(v as number).toLocaleString(), "Clicks"]} cursor={{ fill: "hsl(var(--muted))", opacity: 0.4 }} />
                <Bar dataKey="clicks" radius={[0, 4, 4, 0]} maxBarSize={28}>
                  {topLinksData.map((_, i) => (
                    <Cell key={i} fill="#e11d48" fillOpacity={TOP_LINK_COLORS[i] ?? 0.1} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}

      {/* Device + Browser donut charts */}
      {hasClicks && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <DonutChart title="Device breakdown" items={deviceItems} />
          <DonutChart title="Browser breakdown" items={browserItems} />
        </div>
      )}

      {/* Top referrers */}
      {hasClicks && (
        <BreakdownList
          title="Top referrers"
          items={refererItems}
          formatName={(name) => (name === "direct" ? "Direct / None" : name)}
        />
      )}

      {!hasClicks && (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <p className="font-medium">No clicks recorded yet</p>
            <p className="text-sm mt-1">Share your short links and charts will appear here.</p>
          </CardContent>
        </Card>
      )}

      <p className="text-xs text-muted-foreground text-center">
        Analytics are aggregated and anonymised. Raw visitor signals (IP addresses, user agents) are
        never stored.
      </p>
    </div>
  )
}
