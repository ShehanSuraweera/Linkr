import { Suspense } from "react"
import { redirect } from "next/navigation"
import { getLinks, ApiError } from "@/lib/api"
import LinkTable from "@/components/LinkTable"

async function LinksData() {
  let data
  try {
    data = await getLinks()
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/api/auth/logout")
    }
    return (
      <div className="text-center py-16 text-destructive">
        <p>Failed to load links. Is the API running?</p>
      </div>
    )
  }
  return (
    <LinkTable
      initialLinks={data.items}
      initialHasMore={data.has_more}
      initialNextCursor={data.next_cursor}
    />
  )
}

export default function DashboardPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-10 w-64 bg-muted rounded-lg animate-pulse" />
          <div className="h-10 bg-muted rounded-lg animate-pulse" />
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-12 bg-muted rounded-lg animate-pulse" />
            ))}
          </div>
        </div>
      }
    >
      <LinksData />
    </Suspense>
  )
}
