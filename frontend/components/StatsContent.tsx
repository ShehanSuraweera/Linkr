"use client"

import { useQuery } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import type { LinkStats } from "@/lib/types"
import ClicksChart from "@/components/ClicksChart"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { RefreshCw } from "lucide-react"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

interface Props {
  code: string
  initialStats?: LinkStats
}

export default function StatsContent({ code, initialStats }: Props) {
  const router = useRouter()

  const { data: stats, isLoading, isError, error, refetch, isRefetching } = useQuery<LinkStats>({
    queryKey: ["stats", code],
    queryFn: async () => {
      const res = await fetch(`/api/links-proxy/${code}/stats`)
      if (res.status === 401) {
        router.push("/login")
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

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="h-24 bg-muted rounded-xl animate-pulse" />
          <div className="h-24 bg-muted rounded-xl animate-pulse" />
        </div>
        <div className="h-56 bg-muted rounded-xl animate-pulse" />
      </div>
    )
  }

  if (isError) {
    const status = (error as { status?: number }).status
    if (status === 404) return <p className="text-muted-foreground">Link not found.</p>
    if (status === 403) return <p className="text-muted-foreground">You don&apos;t own this link.</p>
    return <p className="text-destructive">Failed to load stats. Is the API running?</p>
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground mb-1">Short link</p>
            <a
              href={`${API_BASE}/${code}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary font-mono hover:underline"
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
      </div>

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
    </div>
  )
}
