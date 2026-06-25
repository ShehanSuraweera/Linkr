"use client"

import {
  BarChart,
  Bar,
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
      <div className="flex items-center justify-center h-40 text-gray-400 text-sm">
        No clicks yet
      </div>
    )
  }

  const data = daily.map((d) => ({ date: d.Day, clicks: d.Count }))

  return (
    <ResponsiveContainer width="100%" height={220}>
      <BarChart data={data} margin={{ top: 4, right: 8, left: -16, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 11, fill: "#9ca3af" }}
          tickFormatter={(v) =>
            new Date(v + "T00:00:00").toLocaleDateString("en-US", { month: "short", day: "numeric" })
          }
        />
        <YAxis tick={{ fontSize: 11, fill: "#9ca3af" }} allowDecimals={false} />
        <Tooltip
          formatter={(v) => [v, "Clicks"]}
          labelFormatter={(l) => new Date(l + "T00:00:00").toLocaleDateString()}
        />
        <Bar dataKey="clicks" fill="#6366f1" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  )
}
