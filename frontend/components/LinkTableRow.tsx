"use client"

import { BarChart2, Check, Copy, Power, Trash2 } from "lucide-react"
import {
  TableCell,
  TableRow,
} from "@/components/ui/table"
import type { Link as LinkType } from "@/lib/types"
import { cn, formatDate } from "@/lib/utils"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

interface LinkTableRowProps {
  link: LinkType
  copiedId: number | null
  onCopy: (code: string, id: number) => void
  onToggle: (link: LinkType) => void
  onDelete: (link: LinkType) => void
  isToggling: boolean
  isDeleting: boolean
  onNavigate: () => void
}

export default function LinkTableRow({
  link,
  copiedId,
  onCopy,
  onToggle,
  onDelete,
  isToggling,
  isDeleting,
  onNavigate,
}: LinkTableRowProps) {
  return (
    <TableRow className="group cursor-pointer" onClick={onNavigate}>
      <TableCell>
        <div className="flex items-center gap-1.5">
          <span className={cn("h-2 w-2 rounded-full shrink-0", link.is_active ? "bg-green-500" : "bg-muted-foreground/40")} />
          <a
            href={`${API_BASE}/${link.short_code}`}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            className="text-primary font-mono text-sm hover:underline"
          >
            /{link.short_code}
          </a>
          <button
            onClick={(e) => { e.stopPropagation(); onCopy(link.short_code, link.id) }}
            className="opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-foreground"
            aria-label="Copy link"
          >
            {copiedId === link.id
              ? <Check className="h-3.5 w-3.5 text-green-600" />
              : <Copy className="h-3.5 w-3.5" />}
          </button>
        </div>
      </TableCell>
      <TableCell>
        <span className="text-muted-foreground text-sm block truncate max-w-xs" title={link.original_url}>
          {link.original_url}
        </span>
      </TableCell>
      <TableCell className="text-right text-sm text-muted-foreground whitespace-nowrap">
        {formatDate(link.created_at)}
      </TableCell>
      <TableCell className="text-right">
        <span className="inline-flex items-center justify-end gap-1.5 text-sm text-muted-foreground">
          <BarChart2 className="h-3.5 w-3.5 shrink-0" />
          <span className="font-medium tabular-nums">{link.total_clicks.toLocaleString()}</span>
        </span>
      </TableCell>
      <TableCell>
        <div className="flex items-center justify-end gap-1 opacity-100 transition-opacity">
          <button
            onClick={(e) => { e.stopPropagation(); onToggle(link) }}
            disabled={isToggling}
            title={link.is_active ? "Deactivate" : "Activate"}
            className={cn(
              "p-1.5 rounded-md transition-colors",
              link.is_active
                ? "text-green-600 hover:bg-green-50 dark:hover:bg-green-950"
                : "text-muted-foreground hover:text-foreground hover:bg-muted"
            )}
          >
            <Power className="h-3.5 w-3.5" />
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); onDelete(link) }}
            disabled={isDeleting}
            title="Delete"
            className="p-1.5 rounded-md transition-colors text-muted-foreground hover:text-destructive hover:bg-muted"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </TableCell>
    </TableRow>
  )
}
