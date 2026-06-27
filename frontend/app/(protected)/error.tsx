"use client"

import { useEffect } from "react"
import { AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"

export default function Error({
  error,
  unstable_retry,
}: {
  error: Error & { digest?: string }
  unstable_retry: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="flex flex-col items-center justify-center py-24 rounded-xl border border-dashed bg-muted/20">
      <div className="p-3 rounded-full bg-muted mb-4">
        <AlertCircle className="h-8 w-8 text-destructive" />
      </div>
      <p className="font-semibold">Something went wrong</p>
      <p className="text-sm text-muted-foreground mt-1 mb-5">
        {error.message ?? "An unexpected error occurred."}
      </p>
      <Button variant="outline" size="sm" onClick={unstable_retry}>
        Try again
      </Button>
    </div>
  )
}
