"use client"

import { useQueryClient, useIsFetching } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { RefreshCw } from "lucide-react"

export default function AnalyticsRefreshButton() {
  const queryClient = useQueryClient()
  const fetching = useIsFetching({ queryKey: ["analytics-overview"] })

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={() => queryClient.refetchQueries({ queryKey: ["analytics-overview"] })}
      disabled={fetching > 0}
    >
      <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${fetching > 0 ? "animate-spin" : ""}`} />
      Refresh
    </Button>
  )
}
