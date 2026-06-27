"use client"

import { useState } from "react"
import { useQueryClient, type InfiniteData } from "@tanstack/react-query"
import type { Link as LinkType, ListLinksResponse } from "@/lib/types"

export function useLinksActions() {
  const queryClient = useQueryClient()
  const [togglingId, setTogglingId] = useState<number | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [mutationError, setMutationError] = useState<string | null>(null)

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
    patchAllPages((item) => (item.id === link.id ? { ...item, is_active: next } : item))
    try {
      const res = await fetch(`/api/links/${link.short_code}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_active: next }),
      })
      if (!res.ok) throw new Error("Failed to update link")
      const updated: LinkType = await res.json()
      patchAllPages((item) => (item.id === link.id ? { ...item, is_active: updated.is_active } : item))
    } catch {
      patchAllPages((item) => (item.id === link.id ? { ...item, is_active: link.is_active } : item))
      setMutationError("Failed to update link status.")
    } finally {
      setTogglingId(null)
    }
  }

  const handleDelete = async (link: LinkType) => {
    if (deletingId !== null) return
    setDeletingId(link.id)
    try {
      const res = await fetch(`/api/links/${link.short_code}`, { method: "DELETE" })
      if (!res.ok) throw new Error("Failed to delete link")
      patchAllPages((item) => (item.id === link.id ? null : item))
    } catch {
      setMutationError("Failed to delete link.")
    } finally {
      setDeletingId(null)
    }
  }

  return {
    togglingId,
    deletingId,
    mutationError,
    setMutationError,
    toggleActive,
    handleDelete,
  }
}
