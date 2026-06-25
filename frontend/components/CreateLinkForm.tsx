"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { createLinkSchema, type CreateLinkInput } from "@/lib/schemas"
import type { Link } from "@/lib/types"

interface Props {
  onCreated: (link: Link) => void
}

export default function CreateLinkForm({ onCreated }: Props) {
  const [serverError, setServerError] = useState("")
  const [open, setOpen] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateLinkInput>({ resolver: zodResolver(createLinkSchema) })

  const onSubmit = async (data: CreateLinkInput) => {
    setServerError("")
    try {
      const payload = {
        url: data.url,
        ...(data.alias ? { alias: data.alias } : {}),
        ...(data.expires_at ? { expires_at: new Date(data.expires_at).toISOString() } : {}),
      }
      const res = await fetch("/api/links", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
      const body = await res.json()
      if (!res.ok) {
        setServerError(body.error ?? "Failed to create link")
        return
      }
      onCreated(body as Link)
      reset()
      setOpen(false)
    } catch {
      setServerError("Network error — is the API running?")
    }
  }

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors"
      >
        + New link
      </button>
    )
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="bg-white border border-gray-200 rounded-xl p-5 space-y-3 shadow-sm"
      noValidate
    >
      <h2 className="font-semibold text-gray-800">New short link</h2>

      <div>
        <input
          {...register("url")}
          type="url"
          placeholder="https://example.com/very/long/url"
          className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        {errors.url && <p className="text-red-500 text-xs mt-1">{errors.url.message}</p>}
      </div>

      <div className="flex gap-2">
        <div className="flex-1">
          <input
            {...register("alias")}
            placeholder="Custom alias (optional)"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          {errors.alias && <p className="text-red-500 text-xs mt-1">{errors.alias.message}</p>}
        </div>
        <div className="flex-1">
          <input
            {...register("expires_at")}
            type="datetime-local"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-500"
          />
        </div>
      </div>

      {serverError && (
        <p className="text-red-600 text-sm bg-red-50 rounded-lg px-3 py-2">{serverError}</p>
      )}

      <div className="flex gap-2 justify-end">
        <button
          type="button"
          onClick={() => { setOpen(false); reset(); setServerError("") }}
          className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
        >
          {isSubmitting ? "Creating…" : "Create"}
        </button>
      </div>
    </form>
  )
}
