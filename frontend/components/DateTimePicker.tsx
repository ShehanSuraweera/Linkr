"use client"

import { CalendarDays, Clock } from "lucide-react"
import { cn } from "@/lib/utils"

interface DateTimePickerProps {
  value: string                   // "" | "YYYY-MM-DDTHH:mm"
  onChange: (value: string) => void
  disabled?: boolean
}

// Shared classes that exactly mirror the shadcn Input component.
const inputBase =
  "h-8 w-full min-w-0 rounded-lg border border-input bg-transparent py-1 text-base transition-colors outline-none " +
  "placeholder:text-muted-foreground " +
  "focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 " +
  "disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 " +
  "dark:bg-input/30 dark:disabled:bg-input/80 " +
  "md:text-sm"

export default function DateTimePicker({ value, onChange, disabled }: DateTimePickerProps) {
  const hasT = value.includes("T")
  const datePart = hasT ? value.slice(0, 10) : (value.length === 10 ? value : "")
  const timePart = hasT ? value.slice(11, 16) : ""

  const today = new Date().toISOString().slice(0, 10)

  function emit(date: string, time: string) {
    if (!date) { onChange(""); return }
    onChange(`${date}T${time || "23:59"}`)
  }

  return (
    <div className="flex gap-2">
      {/* ── Date ── */}
      <div className="relative flex-1">
        <CalendarDays
          className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
          aria-hidden
        />
        <input
          type="date"
          value={datePart}
          min={today}
          disabled={disabled}
          onChange={(e) => emit(e.target.value, timePart)}
          className={cn(inputBase, "pl-8 pr-2.5")}
        />
      </div>

      {/* ── Time ── */}
      <div className="relative w-[118px] shrink-0">
        <Clock
          className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
          aria-hidden
        />
        <input
          type="time"
          value={timePart}
          disabled={disabled || !datePart}
          onChange={(e) => emit(datePart, e.target.value)}
          className={cn(inputBase, "pl-8 pr-2.5")}
        />
      </div>
    </div>
  )
}
