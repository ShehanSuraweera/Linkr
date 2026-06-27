"use client"

import { useRouter } from "next/navigation"
import { BarChart2, Check, Copy, Link2, Power, Trash2 } from "lucide-react"
import type { Link as LinkType } from "@/lib/types"
import { cn, getFavicon, getHostname, formatDate } from "@/lib/utils"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

interface LinkCardProps {
  link: LinkType
  copiedId: number | null
  onCopy: (code: string, id: number) => void
  onToggle: (link: LinkType) => void
  onDelete: (link: LinkType) => void
  isToggling: boolean
  isDeleting: boolean
}

export default function LinkCard({
  link,
  copiedId,
  onCopy,
  onToggle,
  onDelete,
  isToggling,
  isDeleting,
}: LinkCardProps) {
  const router = useRouter()
  const shortUrl = `${API_BASE}/${link.short_code}`
  const favicon = getFavicon(link.original_url)
  const hostname = getHostname(link.original_url)

  return (
    <div
      className="flex items-center gap-4 rounded-xl border bg-card px-4 py-3.5 group hover:bg-muted/30 transition-colors cursor-pointer"
      onClick={() => router.push(`/links/${link.short_code}`)}
    >
      {/* Favicon */}
      <div className="shrink-0 h-9 w-9 rounded-lg bg-muted flex items-center justify-center overflow-hidden">
        {favicon
          ? <img src={favicon} alt="" width={20} height={20} className="h-5 w-5" onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = "none" }} />
          : <Link2 className="h-4 w-4 text-muted-foreground" />}
      </div>

      {/* Middle — main info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className={cn("h-2 w-2 rounded-full shrink-0", link.is_active ? "bg-green-500" : "bg-muted-foreground/40")} />
          <a
            href={shortUrl}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            className="text-primary font-mono text-sm font-medium hover:underline truncate"
          >
            /{link.short_code}
          </a>
          <button
            onClick={(e) => { e.stopPropagation(); e.preventDefault(); onCopy(link.short_code, link.id) }}
            className="opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-foreground shrink-0"
            aria-label="Copy link"
          >
            {copiedId === link.id
              ? <Check className="h-3.5 w-3.5 text-green-600" />
              : <Copy className="h-3.5 w-3.5" />}
          </button>
        </div>
        <p className="text-xs text-muted-foreground truncate mt-0.5" title={link.original_url}>
          {hostname}
        </p>
      </div>

      {/* Right — clicks + date + actions */}
      <div className="shrink-0 flex flex-col items-end gap-1">
        <span className="inline-flex items-center gap-1.5 text-sm text-muted-foreground">
          <BarChart2 className="h-3.5 w-3.5 shrink-0" />
          <span className="font-medium tabular-nums">{link.total_clicks.toLocaleString()}</span>
        </span>
        <span className="text-[11px] text-muted-foreground/70">{formatDate(link.created_at)}</span>
        <div className="flex items-center gap-1 opacity-100 transition-opacity mt-0.5">
          <button
            onClick={(e) => { e.stopPropagation(); e.preventDefault(); onToggle(link) }}
            disabled={isToggling}
            title={link.is_active ? "Deactivate" : "Activate"}
            className={cn(
              "p-1 rounded-md transition-colors",
              link.is_active
                ? "text-green-600 hover:bg-green-50 dark:hover:bg-green-950"
                : "text-muted-foreground hover:text-foreground hover:bg-muted"
            )}
          >
            <Power className="h-3.5 w-3.5" />
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); e.preventDefault(); onDelete(link) }}
            disabled={isDeleting}
            title="Delete"
            className="p-1 rounded-md transition-colors text-muted-foreground hover:text-destructive hover:bg-muted"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  )
}
