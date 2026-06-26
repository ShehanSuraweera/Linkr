"use client"

import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts"
import { Card, CardContent } from "@/components/ui/card"

const PALETTE = ["#e11d48", "#f43f5e", "#fb7185", "#fda4af", "#fecdd3", "#94a3b8"]

interface Item {
  name: string
  count: number
}

interface Props {
  title: string
  items: Item[]
  emptyMessage?: string
  formatName?: (name: string) => string
}

export default function DonutChart({
  title,
  items,
  emptyMessage = "No data yet",
  formatName,
}: Props) {
  const total = items.reduce((s, i) => s + i.count, 0)
  const data = items.map((item) => ({
    name: formatName ? formatName(item.name) : item.name,
    value: item.count,
  }))

  return (
    <Card>
      <CardContent className="pt-5">
        <p className="text-sm font-medium text-muted-foreground mb-2">{title}</p>
        {data.length === 0 ? (
          <p className="text-sm text-muted-foreground py-10 text-center">{emptyMessage}</p>
        ) : (
          <ResponsiveContainer width="100%" height={215}>
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="44%"
                innerRadius={52}
                outerRadius={78}
                paddingAngle={2}
                dataKey="value"
              >
                {data.map((_, i) => (
                  <Cell key={i} fill={PALETTE[i % PALETTE.length]} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(value) => {
                  const n = value as number
                  const pct = total > 0 ? Math.round((n / total) * 100) : 0
                  return [`${n.toLocaleString()} (${pct}%)`]
                }}
              />
              <Legend
                iconType="circle"
                iconSize={8}
                wrapperStyle={{ fontSize: 12, textTransform: "capitalize" }}
              />
            </PieChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
