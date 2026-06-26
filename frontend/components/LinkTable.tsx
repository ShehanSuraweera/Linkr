"use client"

import { useState, useEffect, useCallback } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { useInfiniteQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query"
import type { Link as LinkType, ListLinksResponse } from "@/lib/types"
import CreateLinkForm from "./CreateLinkForm"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Copy, Check, BarChart2, Link2, RefreshCw,
  Search, X, AlertCircle, LayoutList, LayoutGrid,
  Power, Trash2,
} from "lucide-react"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"
const PAGE_SIZE = 20
const SEARCH_DEBOUNCE_MS = 300
const VIEW_STORAGE_KEY = "linkr_view"

type ViewMode = "card" | "table"

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return debounced
}

function getFavicon(url: string): string {
  try {
    const { hostname } = new URL(url)
    return `https://www.google.com/s2/favicons?domain=${hostname}&sz=32`
  } catch {
    return ""
  }
}

function getHostname(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "")
  } catch {
    return url
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short", day: "numeric", year: "numeric",
  })
}

interface Props {
  initialLinks: LinkType[]
  initialHasMore: boolean
  initialNextCursor?: string
}

function PaginationSkeleton({ view }: { view: ViewMode }) {
  if (view === "table") {
    return (
      <div className="border-t divide-y">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 px-4 py-3 animate-pulse">
            <div className="h-4 w-28 bg-muted rounded" />
            <div className="flex-1 h-4 bg-muted rounded" />
            <div className="h-4 w-24 bg-muted rounded" />
            <div className="h-4 w-12 bg-muted rounded" />
          </div>
        ))}
      </div>
    )
  }
  return (
    <div className="space-y-3 mt-3">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 rounded-xl border bg-card px-4 py-3.5 animate-pulse">
          <div className="h-9 w-9 rounded-lg bg-muted shrink-0" />
          <div className="flex-1 space-y-2 min-w-0">
            <div className="h-4 w-32 bg-muted rounded" />
            <div className="h-3 w-48 bg-muted rounded" />
          </div>
          <div className="shrink-0 space-y-2 flex flex-col items-end">
            <div className="h-4 w-12 bg-muted rounded" />
            <div className="h-3 w-16 bg-muted rounded" />
          </div>
        </div>
      ))}
    </div>
  )
}

export default function LinkTable({ initialLinks, initialHasMore, initialNextCursor }: Props) {
  const queryClient = useQueryClient()
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [search, setSearch] = useState("")
  const debouncedSearch = useDebounce(search, SEARCH_DEBOUNCE_MS)
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "inactive">("all")
  const [view, setView] = useState<ViewMode>("card")
  const [togglingId, setTogglingId] = useState<number | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [mutationError, setMutationError] = useState<string | null>(null)

  const searchParams = useSearchParams()
  const router = useRouter()

  // Hydrate view preference from localStorage after mount
  useEffect(() => {
    const saved = localStorage.getItem(VIEW_STORAGE_KEY) as ViewMode | null
    if (saved === "table" || saved === "card") setView(saved)
  }, [])

  const changeView = (v: ViewMode) => {
    setView(v)
    localStorage.setItem(VIEW_STORAGE_KEY, v)
  }

  const redirectToLogin = useCallback(() => window.location.replace("/api/auth/logout"), [])

  useEffect(() => {
    if (searchParams.get("new") === "1") {
      setDialogOpen(true)
      router.replace("/dashboard", { scroll: false })
    }
  }, [searchParams, router])

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isFetching,
    refetch,
    isRefetching,
    isError,
    error,
  } = useInfiniteQuery({
    queryKey: ["links", debouncedSearch],
    queryFn: async ({ pageParam }: { pageParam: string | null }) => {
      const params = new URLSearchParams({ limit: String(PAGE_SIZE) })
      if (pageParam) params.set("cursor", pageParam)
      if (debouncedSearch) params.set("q", debouncedSearch)
      const res = await fetch(`/api/links?${params}`)
      if (res.status === 401) { redirectToLogin(); throw new Error("Session expired") }
      if (!res.ok) throw new Error("Failed to fetch links")
      return res.json() as Promise<ListLinksResponse>
    },
    initialPageParam: null as string | null,
    getNextPageParam: (last) => last.has_more && last.next_cursor ? last.next_cursor : null,
    initialData: debouncedSearch ? undefined : {
      pages: [{ items: initialLinks, has_more: initialHasMore, next_cursor: initialNextCursor }],
      pageParams: [null],
    },
    staleTime: 30_000,
  })

  const allLinks = data?.pages.flatMap((p) => p.items) ?? []
  const links = allLinks.filter((l) => {
    if (statusFilter === "active") return l.is_active
    if (statusFilter === "inactive") return !l.is_active
    return true
  })

  const handleCreated = (link: LinkType) => {
    queryClient.setQueryData<InfiniteData<ListLinksResponse>>(["links"], (old) => {
      if (!old) return old
      return {
        ...old,
        pages: [{ ...old.pages[0], items: [link, ...old.pages[0].items] }, ...old.pages.slice(1)],
      }
    })
    setDialogOpen(false)
  }

  const patchAllPages = (updater: (item: LinkType) => LinkType | null) => {
    queryClient.setQueriesData<InfiniteData<ListLinksResponse>>(
      { queryKey: ["links"] },
      (old) => {
        if (!old) return old
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            items: page.items.flatMap((item) => {
              const next = updater(item)
              return next ? [next] : []
            }),
          })),
        }
      }
    )
  }

  const toggleActive = async (link: LinkType) => {
    if (togglingId !== null) return
    setTogglingId(link.id)
    const next = !link.is_active
    // optimistic
    patchAllPages((item) => (item.id === link.id ? { ...item, is_active: next } : item))
    try {
      const res = await fetch(`/api/links/${link.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_active: next }),
      })
      if (!res.ok) throw new Error("Failed to update link")
      const updated: LinkType = await res.json()
      patchAllPages((item) => (item.id === link.id ? { ...item, is_active: updated.is_active } : item))
    } catch {
      // roll back
      patchAllPages((item) => (item.id === link.id ? { ...item, is_active: link.is_active } : item))
      setMutationError("Failed to update link status.")
    } finally {
      setTogglingId(null)
    }
  }

  const handleDelete = async (id: number) => {
    if (deletingId !== null) return
    setDeletingId(id)
    try {
      const res = await fetch(`/api/links/${id}`, { method: "DELETE" })
      if (!res.ok) throw new Error("Failed to delete link")
      patchAllPages((item) => (item.id === id ? null : item))
    } catch {
      setMutationError("Failed to delete link.")
    } finally {
      setDeletingId(null)
    }
  }

  const copyLink = async (shortCode: string, id: number) => {
    await navigator.clipboard.writeText(`${API_BASE}/${shortCode}`)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const loadMoreFooter = (
    <>
      {isError && data && (
        <div className="px-4 py-3 border-t bg-destructive/10 flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {error?.message ?? "Failed to load more links."}
          </div>
          <Button variant="outline" size="sm" onClick={() => fetchNextPage()}>Retry</Button>
        </div>
      )}
      {hasNextPage && !isError && (
        isFetchingNextPage
          ? <PaginationSkeleton view={view} />
          : (
            <div className="px-4 py-3 border-t bg-muted/20 text-center">
              <Button variant="ghost" size="sm" onClick={() => fetchNextPage()}>Load more</Button>
            </div>
          )
      )}
    </>
  )

  return (
    <div className="space-y-6">

      {/* Mutation error banner */}
      {mutationError && (
        <div className="flex items-center justify-between gap-3 rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-2.5">
          <div className="flex items-center gap-2 text-sm text-destructive">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {mutationError}
          </div>
          <button onClick={() => setMutationError(null)} className="text-destructive hover:text-destructive/70">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">My Links</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Create and manage your short links</p>
        </div>
        <div className="flex items-center gap-2">
          {/* View toggle — desktop only */}
          <div className="hidden lg:flex items-center rounded-lg border bg-muted/40 p-0.5 gap-0.5">
            <button
              onClick={() => changeView("card")}
              aria-label="Card view"
              className={cn(
                "p-1.5 rounded-md transition-colors",
                view === "card"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <LayoutGrid className="h-4 w-4" />
            </button>
            <button
              onClick={() => changeView("table")}
              aria-label="Table view"
              className={cn(
                "p-1.5 rounded-md transition-colors",
                view === "table"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <LayoutList className="h-4 w-4" />
            </button>
          </div>

          <Button variant="ghost" size="sm" onClick={() => refetch()} disabled={isRefetching} aria-label="Refresh">
            <RefreshCw className={cn("h-4 w-4", isRefetching && "animate-spin")} />
          </Button>
          <Button className="gap-1.5" onClick={() => setDialogOpen(true)}>
            <Link2 className="h-4 w-4" />
            <span className="hidden sm:inline">Create link</span>
          </Button>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        {isFetching && debouncedSearch
          ? <RefreshCw className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground animate-spin" />
          : <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
        }
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by URL or short code…"
          className="pl-9 pr-9"
        />
        {search && (
          <button
            onClick={() => setSearch("")}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            aria-label="Clear search"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Status filters */}
      <div className="flex items-center gap-1.5">
        {(["all", "active", "inactive"] as const).map((f) => (
          <button
            key={f}
            onClick={() => setStatusFilter(f)}
            className={cn(
              "px-3 py-1 rounded-full text-xs font-medium transition-colors",
              statusFilter === f
                ? "bg-foreground text-background"
                : "bg-muted text-muted-foreground hover:text-foreground"
            )}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
        {statusFilter !== "all" && (
          <span className="ml-1 text-xs text-muted-foreground">{links.length} of {allLinks.length}</span>
        )}
      </div>

      {/* Content */}
      {isError && !data ? (
        <EmptyState icon={<AlertCircle className="h-8 w-8 text-destructive" />}>
          <p className="font-semibold">Failed to load links</p>
          <p className="text-sm text-muted-foreground mt-1 mb-5">{error?.message ?? "Something went wrong."}</p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button>
        </EmptyState>
      ) : links.length === 0 ? (
        <EmptyState icon={<Link2IconSvg />}>
          {debouncedSearch ? (
            <>
              <p className="font-semibold">No results for &ldquo;{debouncedSearch}&rdquo;</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">Try a different URL or short code.</p>
              <Button variant="outline" size="sm" onClick={() => setSearch("")}>Clear search</Button>
            </>
          ) : statusFilter !== "all" ? (
            <>
              <p className="font-semibold">No {statusFilter} links</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">Try a different filter.</p>
              <Button variant="outline" size="sm" onClick={() => setStatusFilter("all")}>Show all</Button>
            </>
          ) : (
            <>
              <p className="font-semibold">No links yet</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">Create your first short link to get started.</p>
              <Button className="gap-1.5" onClick={() => setDialogOpen(true)}>
                <Link2 className="h-4 w-4" />Create link
              </Button>
            </>
          )}
        </EmptyState>
      ) : (
        /* lg:card or lg:table; always card on mobile */
        <div>
          {/* Card view — always on mobile, conditional on desktop */}
          <div className={cn(view === "table" ? "lg:hidden" : "")}>
            <div className="space-y-3">
              {links.map((link) => (
                <LinkCard
                  key={link.id}
                  link={link}
                  copiedId={copiedId}
                  onCopy={copyLink}
                  onToggle={toggleActive}
                  onDelete={handleDelete}
                  isToggling={togglingId === link.id}
                  isDeleting={deletingId === link.id}
                />
              ))}
            </div>
            <div className="mt-3 rounded-xl border overflow-hidden">{loadMoreFooter}</div>
          </div>

          {/* Table view — desktop only, when table mode is active */}
          {view === "table" && (
            <div className="hidden lg:block rounded-xl border overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/40 hover:bg-muted/40">
                    <TableHead className="w-44">Short link</TableHead>
                    <TableHead>Target URL</TableHead>
                    <TableHead className="text-right w-32">Created</TableHead>
                    <TableHead className="text-right w-28">Clicks</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {links.map((link) => (
                    <TableRowWithActions
                      key={link.id}
                      link={link}
                      copiedId={copiedId}
                      onCopy={copyLink}
                      onToggle={toggleActive}
                      onDelete={handleDelete}
                      isToggling={togglingId === link.id}
                      isDeleting={deletingId === link.id}
                      onNavigate={() => router.push(`/links/${link.short_code}`)}
                    />
                  ))}
                </TableBody>
              </Table>
              {loadMoreFooter}
            </div>
          )}
        </div>
      )}

      <CreateLinkForm open={dialogOpen} onOpenChange={setDialogOpen} onCreated={handleCreated} />
    </div>
  )
}

// ── Card component ────────────────────────────────────────────────────────────

function LinkCard({
  link,
  copiedId,
  onCopy,
  onToggle,
  onDelete,
  isToggling,
  isDeleting,
}: {
  link: LinkType
  copiedId: number | null
  onCopy: (code: string, id: number) => void
  onToggle: (link: LinkType) => void
  onDelete: (id: number) => void
  isToggling: boolean
  isDeleting: boolean
}) {
  const router = useRouter()
  const [confirmDelete, setConfirmDelete] = useState(false)
  const shortUrl = `${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/${link.short_code}`
  const favicon = getFavicon(link.original_url)
  const hostname = getHostname(link.original_url)

  // Auto-reset delete confirmation after 3 s
  useEffect(() => {
    if (!confirmDelete) return
    const t = setTimeout(() => setConfirmDelete(false), 3000)
    return () => clearTimeout(t)
  }, [confirmDelete])

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    e.preventDefault()
    if (confirmDelete) {
      setConfirmDelete(false)
      onDelete(link.id)
    } else {
      setConfirmDelete(true)
    }
  }

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
        {/* Action buttons — visible on hover */}
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity mt-0.5">
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
            onClick={handleDeleteClick}
            disabled={isDeleting}
            title={confirmDelete ? "Click again to confirm" : "Delete"}
            className={cn(
              "p-1 rounded-md transition-colors",
              confirmDelete
                ? "text-destructive bg-destructive/10 hover:bg-destructive/20"
                : "text-muted-foreground hover:text-destructive hover:bg-muted"
            )}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Table row with actions ────────────────────────────────────────────────────

function TableRowWithActions({
  link,
  copiedId,
  onCopy,
  onToggle,
  onDelete,
  isToggling,
  isDeleting,
  onNavigate,
}: {
  link: LinkType
  copiedId: number | null
  onCopy: (code: string, id: number) => void
  onToggle: (link: LinkType) => void
  onDelete: (id: number) => void
  isToggling: boolean
  isDeleting: boolean
  onNavigate: () => void
}) {
  const [confirmDelete, setConfirmDelete] = useState(false)

  useEffect(() => {
    if (!confirmDelete) return
    const t = setTimeout(() => setConfirmDelete(false), 3000)
    return () => clearTimeout(t)
  }, [confirmDelete])

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirmDelete) {
      setConfirmDelete(false)
      onDelete(link.id)
    } else {
      setConfirmDelete(true)
    }
  }

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
        <div className="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
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
            onClick={handleDeleteClick}
            disabled={isDeleting}
            title={confirmDelete ? "Click again to confirm" : "Delete"}
            className={cn(
              "p-1.5 rounded-md transition-colors",
              confirmDelete
                ? "text-destructive bg-destructive/10 hover:bg-destructive/20"
                : "text-muted-foreground hover:text-destructive hover:bg-muted"
            )}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </TableCell>
    </TableRow>
  )
}

function EmptyState({ icon, children }: { icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 rounded-xl border border-dashed bg-muted/20">
      <div className="p-3 rounded-full bg-muted mb-4">{icon}</div>
      {children}
    </div>
  )
}

function Link2IconSvg() {
  return (
    <svg className="h-5 w-5 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
    </svg>
  )
}
