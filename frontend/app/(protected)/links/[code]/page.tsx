import Link from "next/link"
import { ChevronLeft } from "lucide-react"
import StatsContent from "@/components/StatsContent"

export default async function StatsPage({
  params,
}: {
  params: Promise<{ code: string }>
}) {
  const { code } = await params

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

      <StatsContent code={code} />
    </div>
  )
}
