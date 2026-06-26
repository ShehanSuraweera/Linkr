"use client"

import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts"
import type { DailyStat } from "@/lib/types"

export default function ClicksChart({ daily }: { daily: DailyStat[] }) {
  if (daily.length === 0) {
    return (
      <div className="flex items-center justify-center h-40 text-muted-foreground text-sm">
        No clicks yet
      </div>
    )
  }

  const data = daily.map((d) => ({ date: d.day, clicks: d.count }))

  return (
    <ResponsiveContainer width="100%" height={220}>
      <AreaChart data={data} margin={{ top: 4, right: 8, left: -16, bottom: 0 }}>
        <defs>
          <linearGradient id="clicksFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.18} />
            <stop offset="95%" stopColor="var(--primary)" stopOpacity={0.02} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 11 }}
          className="fill-muted-foreground"
          tickFormatter={(v) =>
            new Date(v + "T00:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })
          }
        />
        <YAxis tick={{ fontSize: 11 }} className="fill-muted-foreground" allowDecimals={false} />
        <Tooltip
          formatter={(v) => [v, "Clicks"]}
          labelFormatter={(l) => new Date(l + "T00:00:00").toLocaleDateString()}
        />
        <Area
          type="monotone"
          dataKey="clicks"
          stroke="var(--primary)"
          strokeWidth={2}
          fill="url(#clicksFill)"
          dot={false}
          activeDot={{ r: 5, fill: "var(--primary)" }}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
