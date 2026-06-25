import { Suspense } from "react"
import Link from "next/link"
import { getLinkStats } from "@/lib/api"
import { ApiError } from "@/lib/api"
import ClicksChart from "@/components/ClicksChart"
import LogoutButton from "@/components/LogoutButton"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

async function StatsContent({ code }: { code: string }) {
  try {
    const stats = await getLinkStats(code)
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <p className="text-sm text-gray-500 mb-1">Short link</p>
            <a
              href={`${API_BASE}/${code}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-indigo-600 font-mono hover:underline"
            >
              /{code}
            </a>
          </div>
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <p className="text-sm text-gray-500 mb-1">Total clicks</p>
            <p className="text-3xl font-bold text-gray-900">{stats.total_clicks}</p>
          </div>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h2 className="text-sm font-medium text-gray-600 mb-4">Clicks over time</h2>
          <ClicksChart daily={stats.daily ?? []} />
        </div>
      </div>
    )
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return <p className="text-gray-500">Link not found.</p>
    }
    if (err instanceof ApiError && err.status === 403) {
      return <p className="text-gray-500">You don&apos;t own this link.</p>
    }
    return <p className="text-red-500">Failed to load stats. Is the API running?</p>
  }
}

export default async function StatsPage({
  params,
}: {
  params: Promise<{ code: string }>
}) {
  const { code } = await params

  return (
    <div className="min-h-screen">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/dashboard" className="text-gray-400 hover:text-gray-600 text-sm">
              ← Dashboard
            </Link>
            <h1 className="text-xl font-bold text-gray-900">Stats</h1>
          </div>
          <LogoutButton />
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <Suspense
          fallback={
            <div className="space-y-4">
              <div className="h-24 bg-gray-100 rounded-xl animate-pulse" />
              <div className="h-56 bg-gray-100 rounded-xl animate-pulse" />
            </div>
          }
        >
          <StatsContent code={code} />
        </Suspense>
      </main>
    </div>
  )
}
