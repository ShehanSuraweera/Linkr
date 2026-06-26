"use client"

import { useState, useEffect, useCallback } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { useInfiniteQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query"
import Link from "next/link"
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
import { Copy, Check, BarChart2, Link2, RefreshCw, Search, X } from "lucide-react"
import { Input } from "@/components/ui/input"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

const PAGE_SIZE = 20
const SEARCH_DEBOUNCE_MS = 300

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return debounced
}

interface Props {
  initialLinks: LinkType[]
  initialHasMore: boolean
  initialNextCursor?: string
}

export default function LinkTable({ initialLinks, initialHasMore, initialNextCursor }: Props) {
  const queryClient = useQueryClient()
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [search, setSearch] = useState("")
  const debouncedSearch = useDebounce(search, SEARCH_DEBOUNCE_MS)
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "inactive">("all")

  const searchParams = useSearchParams()
  const router = useRouter()

  const redirectToLogin = useCallback(() => {
    router.push("/login")
  }, [router])

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
  } = useInfiniteQuery({
    queryKey: ["links", debouncedSearch],
    queryFn: async ({ pageParam }: { pageParam: string | null }) => {
      const params = new URLSearchParams({ limit: String(PAGE_SIZE) })
      if (pageParam) params.set("cursor", pageParam)
      if (debouncedSearch) params.set("q", debouncedSearch)
      const res = await fetch(`/api/links?${params}`)
      if (res.status === 401) {
        redirectToLogin()
        throw new Error("Session expired")
      }
      if (!res.ok) throw new Error("Failed to fetch links")
      return res.json() as Promise<ListLinksResponse>
    },
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) =>
      lastPage.has_more && lastPage.next_cursor ? lastPage.next_cursor : null,
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
        pages: [
          { ...old.pages[0], items: [link, ...old.pages[0].items] },
          ...old.pages.slice(1),
        ],
      }
    })
    setDialogOpen(false)
  }

  const copyLink = async (shortCode: string, id: number) => {
    await navigator.clipboard.writeText(`${API_BASE}/${shortCode}`)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">My Links</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Create and manage your short links</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => refetch()}
            disabled={isRefetching}
            aria-label="Refresh links"
          >
            <RefreshCw className={`h-4 w-4 ${isRefetching ? "animate-spin" : ""}`} />
          </Button>
          <Button className="gap-1.5" onClick={() => setDialogOpen(true)}>
            <Link2 className="h-4 w-4" />
            Create link
          </Button>
        </div>
      </div>

      <div className="relative">
        {isFetching && debouncedSearch ? (
          <RefreshCw className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground animate-spin" />
        ) : (
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
        )}
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

      <div className="flex items-center gap-1.5">
        {(["all", "active", "inactive"] as const).map((f) => (
          <button
            key={f}
            onClick={() => setStatusFilter(f)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition-colors ${
              statusFilter === f
                ? "bg-foreground text-background"
                : "bg-muted text-muted-foreground hover:text-foreground"
            }`}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
        {statusFilter !== "all" && (
          <span className="ml-1 text-xs text-muted-foreground">
            {links.length} of {allLinks.length}
          </span>
        )}
      </div>

      {links.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 rounded-xl border border-dashed bg-muted/20">
          <div className="p-3 rounded-full bg-muted mb-4">
            <Link2Icon />
          </div>
          {debouncedSearch ? (
            <>
              <p className="font-semibold text-foreground">No results for &ldquo;{debouncedSearch}&rdquo;</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">Try a different URL or short code.</p>
              <Button variant="outline" size="sm" onClick={() => setSearch("")}>Clear search</Button>
            </>
          ) : statusFilter !== "all" ? (
            <>
              <p className="font-semibold text-foreground">No {statusFilter} links</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">Try a different filter.</p>
              <Button variant="outline" size="sm" onClick={() => setStatusFilter("all")}>Show all</Button>
            </>
          ) : (
            <>
              <p className="font-semibold text-foreground">No links yet</p>
              <p className="text-sm text-muted-foreground mt-1 mb-5">
                Create your first short link to get started.
              </p>
              <Button className="gap-1.5" onClick={() => setDialogOpen(true)}>
                <Link2 className="h-4 w-4" />
                Create link
              </Button>
            </>
          )}
        </div>
      ) : (
        <div className="rounded-xl border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/40 hover:bg-muted/40">
                <TableHead className="w-44">Short link</TableHead>
                <TableHead>Target URL</TableHead>
                <TableHead className="text-right w-32">Created</TableHead>
                <TableHead className="text-right w-28">Clicks</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {links.map((link) => (
                <TableRow key={link.id} className="group">
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      <a
                        href={`${API_BASE}/${link.short_code}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary font-mono text-sm hover:underline"
                      >
                        /{link.short_code}
                      </a>
                      <button
                        onClick={() => copyLink(link.short_code, link.id)}
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
                    <span
                      className="text-muted-foreground text-sm block truncate max-w-xs"
                      title={link.original_url}
                    >
                      {link.original_url}
                    </span>
                  </TableCell>
                  <TableCell className="text-right text-sm text-muted-foreground whitespace-nowrap">
                    {new Date(link.created_at).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                      year: "numeric",
                    })}
                  </TableCell>
                  <TableCell className="text-right">
                    <Link
                      href={`/links/${link.short_code}`}
                      className="inline-flex items-center justify-end gap-1.5 text-sm text-muted-foreground hover:text-primary transition-colors"
                    >
                      <BarChart2 className="h-3.5 w-3.5 shrink-0" />
                      <span className="font-medium tabular-nums">
                        {link.total_clicks.toLocaleString()}
                      </span>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {hasNextPage && (
            <div className="px-4 py-3 border-t bg-muted/20 text-center">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => fetchNextPage()}
                disabled={isFetchingNextPage}
              >
                {isFetchingNextPage ? "Loading…" : "Load more"}
              </Button>
            </div>
          )}
        </div>
      )}

      <CreateLinkForm
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onCreated={handleCreated}
      />
    </div>
  )
}

function Link2Icon() {
  return (
    <svg className="h-5 w-5 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
    </svg>
  )
}
