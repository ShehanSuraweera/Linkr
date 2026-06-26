import { Suspense } from "react"
import Link from "next/link"
import { getLinks } from "@/lib/api"
import { ApiError } from "@/lib/api"
import LinkTable from "@/components/LinkTable"
import DashboardStats from "@/components/DashboardStats"

async function Links() {
  try {
    const data = await getLinks()
    return (
      <div className="space-y-6">
        <DashboardStats links={data.items} />
        <LinkTable
          initialLinks={data.items}
          initialHasMore={data.has_more}
          initialNextCursor={data.next_cursor}
        />
      </div>
    )
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return (
        <div className="text-center py-16 text-muted-foreground">
          <p>Session expired. Please <Link href="/login" className="text-primary underline">sign in again</Link>.</p>
        </div>
      )
    }
    return (
      <div className="text-center py-16 text-destructive">
        <p>Failed to load links. Is the API running?</p>
      </div>
    )
  }
}

export default function DashboardPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-24 bg-muted rounded-xl animate-pulse" />
            ))}
          </div>
          <div className="h-8 w-48 bg-muted rounded-lg animate-pulse" />
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-12 bg-muted rounded-lg animate-pulse" />
            ))}
          </div>
        </div>
      }
    >
      <Links />
    </Suspense>
  )
}
