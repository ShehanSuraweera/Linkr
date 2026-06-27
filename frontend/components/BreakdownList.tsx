import { Card, CardContent } from "@/components/ui/card"

interface BreakdownItem {
  name: string
  count: number
}

interface Props {
  title: string
  items: BreakdownItem[]
  emptyMessage?: string
  formatName?: (name: string) => string
}

// Renders a ranked list with proportional fill bars — same pattern used by
// Plausible and Fathom for small categorical breakdowns.
export default function BreakdownList({
  title,
  items,
  emptyMessage = "No data yet",
  formatName,
}: Props) {
  const total = items.reduce((s, i) => s + i.count, 0)
  const max = items[0]?.count ?? 1

  return (
    <Card>
      <CardContent className="pt-5">
        <p className="text-sm font-medium text-muted-foreground mb-3">{title}</p>

        {items.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center">{emptyMessage}</p>
        ) : (
          <ul className="space-y-2">
            {items.map((item) => {
              const pct = total > 0 ? Math.round((item.count / total) * 100) : 0
              const barWidth = max > 0 ? (item.count / max) * 100 : 0
              const label = formatName ? formatName(item.name) : item.name

              return (
                <li key={item.name} className="group">
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="font-medium capitalize truncate max-w-[60%]">{label}</span>
                    <span className="text-muted-foreground tabular-nums">
                      {item.count.toLocaleString()}
                      <span className="ml-1.5 text-xs">({pct}%)</span>
                    </span>
                  </div>
                  <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                    <div
                      className="h-full rounded-full bg-primary transition-all duration-300"
                      style={{ width: `${barWidth}%` }}
                    />
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
