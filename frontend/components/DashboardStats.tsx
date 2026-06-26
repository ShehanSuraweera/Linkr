"use client"

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts"
import { Card, CardContent } from "@/components/ui/card"
import type { Link } from "@/lib/types"

interface Props {
  links: Link[]
}

export default function DashboardStats({ links }: Props) {
  const totalClicks = links.reduce((sum, l) => sum + l.total_clicks, 0)
  const activeLinks = links.filter((l) => l.is_active).length

  const topLinks = [...links]
    .sort((a, b) => b.total_clicks - a.total_clicks)
    .slice(0, 5)
    .map((l) => ({ name: `/${l.short_code}`, clicks: l.total_clicks }))

  const hasClicks = totalClicks > 0

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Links</p>
            <p className="text-3xl font-bold">{links.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Total clicks</p>
            <p className="text-3xl font-bold">{totalClicks.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-xs text-muted-foreground uppercase tracking-wide mb-1">Active</p>
            <p className="text-3xl font-bold">{activeLinks}</p>
          </CardContent>
        </Card>
      </div>

      {hasClicks && (
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm font-medium text-muted-foreground mb-4">Top links by clicks</p>
            <ResponsiveContainer width="100%" height={topLinks.length * 44 + 16}>
              <BarChart
                data={topLinks}
                layout="vertical"
                margin={{ top: 0, right: 24, left: 0, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} className="stroke-border" />
                <XAxis
                  type="number"
                  tick={{ fontSize: 11 }}
                  className="fill-muted-foreground"
                  allowDecimals={false}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  width={90}
                  tick={{ fontSize: 12 }}
                  className="fill-muted-foreground"
                />
                <Tooltip
                  formatter={(v) => [v, "Clicks"]}
                  cursor={{ fill: "var(--muted)", opacity: 0.4 }}
                />
                <Bar dataKey="clicks" radius={[0, 4, 4, 0]} maxBarSize={28}>
                  {topLinks.map((_, i) => (
                    <Cell
                      key={i}
                      fill={i === 0 ? "var(--primary)" : "var(--primary)"}
                      opacity={1 - i * 0.14}
                    />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
