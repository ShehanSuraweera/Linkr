import Link from "next/link"
import { ChevronLeft } from "lucide-react"
import StatsContent from "@/components/StatsContent"
import { getLinkStats } from "@/lib/api"
import type { LinkStats } from "@/lib/types"

export default async function StatsPage({
  params,
}: {
  params: Promise<{ code: string }>
}) {
  const { code } = await params

  let initialStats: LinkStats | undefined
  try {
    initialStats = await getLinkStats(code)
  } catch {
    // Let StatsContent handle the error (404, 403, network, etc.)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Link
          href="/dashboard"
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ChevronLeft className="h-4 w-4" />
          Dashboard
        </Link>
        <span className="text-border select-none">/</span>
        <span className="text-sm font-semibold">Stats for /{code}</span>
      </div>

      <StatsContent code={code} initialStats={initialStats} />
    </div>
  )
}
