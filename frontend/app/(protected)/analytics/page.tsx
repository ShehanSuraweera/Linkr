import { Suspense } from "react"
import Link from "next/link"
import { getOverview, ApiError } from "@/lib/api"
import DashboardStats from "@/components/DashboardStats"
import AnalyticsRefreshButton from "@/components/AnalyticsRefreshButton"
import type { OverviewStats } from "@/lib/types"

async function AnalyticsData() {
  let initialStats: OverviewStats | undefined
  try {
    initialStats = await getOverview()
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return (
        <div className="text-center py-16 text-muted-foreground">
          <p>
            Session expired. Please{" "}
            <Link href="/login" className="text-primary underline">
              sign in again
            </Link>
            .
          </p>
        </div>
      )
    }
    return (
      <div className="text-center py-16 text-destructive">
        <p>Failed to load analytics. Is the API running?</p>
      </div>
    )
  }

  return <DashboardStats initialStats={initialStats} />
}

export default function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Analytics</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Overview of your link performance</p>
        </div>
        <AnalyticsRefreshButton />
      </div>

      <Suspense
        fallback={
          <div className="space-y-4">
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-24 bg-muted rounded-xl animate-pulse" />
              ))}
            </div>
            <div className="h-56 bg-muted rounded-xl animate-pulse" />
            <div className="h-56 bg-muted rounded-xl animate-pulse" />
          </div>
        }
      >
        <AnalyticsData />
      </Suspense>
    </div>
  )
}
