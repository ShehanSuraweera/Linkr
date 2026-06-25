import { Suspense } from "react"
import { getLinks } from "@/lib/api"
import { ApiError } from "@/lib/api"
import LinkTable from "@/components/LinkTable"
import LogoutButton from "@/components/LogoutButton"

async function Links() {
  try {
    const data = await getLinks()
    return (
      <LinkTable
        initialLinks={data.items}
        initialHasMore={data.has_more}
        initialNextCursor={data.next_cursor}
      />
    )
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      return (
        <div className="text-center py-16 text-gray-400">
          <p>Session expired. Please sign in again.</p>
        </div>
      )
    }
    return (
      <div className="text-center py-16 text-red-500">
        <p>Failed to load links. Is the API running?</p>
      </div>
    )
  }
}

export default function DashboardPage() {
  return (
    <div className="min-h-screen">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">Linkr</h1>
          <LogoutButton />
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <Suspense
          fallback={
            <div className="space-y-3">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="h-12 bg-gray-100 rounded-lg animate-pulse" />
              ))}
            </div>
          }
        >
          <Links />
        </Suspense>
      </main>
    </div>
  )
}
