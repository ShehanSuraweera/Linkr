"use client"

import { useState } from "react"
import Link from "next/link"
import type { Link as LinkType } from "@/lib/types"
import CreateLinkForm from "./CreateLinkForm"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

interface Props {
  initialLinks: LinkType[]
  initialHasMore: boolean
  initialNextCursor?: string
}

export default function LinkTable({ initialLinks, initialHasMore, initialNextCursor }: Props) {
  const [links, setLinks] = useState<LinkType[]>(initialLinks)
  const [hasMore, setHasMore] = useState(initialHasMore)
  const [cursor, setCursor] = useState(initialNextCursor)
  const [loadingMore, setLoadingMore] = useState(false)

  const handleCreated = (link: LinkType) => {
    setLinks((prev) => [link, ...prev])
  }

  const loadMore = async () => {
    if (!cursor) return
    setLoadingMore(true)
    try {
      const res = await fetch(`/api/links?cursor=${cursor}`)
      const data = await res.json()
      setLinks((prev) => [...prev, ...data.items])
      setHasMore(data.has_more)
      setCursor(data.next_cursor)
    } finally {
      setLoadingMore(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-800">Your links</h2>
        <CreateLinkForm onCreated={handleCreated} />
      </div>

      {links.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-lg">No links yet</p>
          <p className="text-sm mt-1">Create your first short link above.</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Short link</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Original URL</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Created</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Analytics</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {links.map((link) => (
                <tr key={link.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3">
                    <a
                      href={`${API_BASE}/${link.short_code}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-indigo-600 font-mono hover:underline"
                    >
                      /{link.short_code}
                    </a>
                  </td>
                  <td className="px-4 py-3 text-gray-600 max-w-xs truncate">
                    {link.original_url}
                  </td>
                  <td className="px-4 py-3 text-gray-400 text-right whitespace-nowrap">
                    {new Date(link.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Link
                      href={`/links/${link.short_code}`}
                      className="text-indigo-600 hover:underline"
                    >
                      Stats →
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {hasMore && (
            <div className="px-4 py-3 border-t border-gray-100 text-center">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="text-indigo-600 text-sm hover:underline disabled:opacity-50"
              >
                {loadingMore ? "Loading…" : "Load more"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
