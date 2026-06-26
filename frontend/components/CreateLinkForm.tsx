"use client"

import { useState } from "react"
import { useForm, Controller } from "react-hook-form"
import { toast } from "sonner"
import { zodResolver } from "@hookform/resolvers/zod"
import { createLinkSchema, type CreateLinkInput } from "@/lib/schemas"
import type { Link } from "@/lib/types"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import DateTimePicker from "@/components/DateTimePicker"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (link: Link) => void
}

export default function CreateLinkForm({ open, onOpenChange, onCreated }: Props) {
  const [serverError, setServerError] = useState("")

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isSubmitting },
  } = useForm<CreateLinkInput>({ resolver: zodResolver(createLinkSchema) })

  const handleOpenChange = (next: boolean) => {
    onOpenChange(next)
    if (!next) { reset(); setServerError("") }
  }

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
        // Server validation errors (alias taken, invalid URL, etc.) stay inline
        // so the user can correct the input without losing the form.
        setServerError(body.error ?? "Failed to create link")
        return
      }
      toast.success("Short link created", {
        description: `/${(body as Link).short_code} is ready to share.`,
      })
      onCreated(body as Link)
      reset()
      onOpenChange(false)
    } catch {
      toast.error("Network error", { description: "Could not reach the API. Is the server running?" })
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New short link</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-1" noValidate>

          {/* Destination URL */}
          <div className="space-y-1.5">
            <Label htmlFor="url">Destination URL</Label>
            <Input
              id="url"
              {...register("url")}
              type="url"
              placeholder="https://example.com/very/long/url"
              aria-invalid={!!errors.url}
              autoFocus
            />
            {errors.url && (
              <p className="text-destructive text-xs">{errors.url.message}</p>
            )}
          </div>

          {/* Custom alias */}
          <div className="space-y-1.5">
            <Label htmlFor="alias">
              Custom alias{" "}
              <span className="text-muted-foreground font-normal">(optional)</span>
            </Label>
            <Input
              id="alias"
              {...register("alias")}
              placeholder="my-link"
              aria-invalid={!!errors.alias}
            />
            {errors.alias && (
              <p className="text-destructive text-xs">{errors.alias.message}</p>
            )}
          </div>

          {/* Expiry — date + time pickers */}
          <div className="space-y-1.5">
            <Label>
              Expires at{" "}
              <span className="text-muted-foreground font-normal">(optional)</span>
            </Label>
            <Controller
              control={control}
              name="expires_at"
              render={({ field }) => (
                <DateTimePicker
                  value={field.value ?? ""}
                  onChange={field.onChange}
                />
              )}
            />
            <p className="text-muted-foreground text-xs">
              Time defaults to 23:59 if only a date is chosen.
            </p>
          </div>

          {serverError && (
            <p className="text-destructive text-sm bg-destructive/10 rounded-lg px-3 py-2">
              {serverError}
            </p>
          )}

          <div className="flex gap-2 justify-end pt-1">
            <Button type="button" variant="ghost" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Creating…" : "Create link"}
            </Button>
          </div>

        </form>
      </DialogContent>
    </Dialog>
  )
}
