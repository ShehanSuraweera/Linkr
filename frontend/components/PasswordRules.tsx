"use client"

import { Check, X } from "lucide-react"
import { cn } from "@/lib/utils"

const RULES = [
  { label: "At least 8 characters",  test: (p: string) => p.length >= 8 },
  { label: "One uppercase letter",   test: (p: string) => /[A-Z]/.test(p) },
  { label: "One number",             test: (p: string) => /[0-9]/.test(p) },
]

interface Props {
  password: string
  show: boolean
}

export default function PasswordRules({ password, show }: Props) {
  if (!show) return null

  const hasInput = password.length > 0

  return (
    <ul className="space-y-1.5 pt-1">
      {RULES.map((rule) => {
        const passed = rule.test(password)
        return (
          <li
            key={rule.label}
            className={cn(
              "flex items-center gap-2 text-xs transition-colors duration-150",
              !hasInput
                ? "text-muted-foreground"
                : passed
                  ? "text-green-600 dark:text-green-400"
                  : "text-destructive"
            )}
          >
            {!hasInput ? (
              <span className="h-3.5 w-3.5 shrink-0 rounded-full border border-current opacity-50" />
            ) : passed ? (
              <Check className="h-3.5 w-3.5 shrink-0" />
            ) : (
              <X className="h-3.5 w-3.5 shrink-0" />
            )}
            {rule.label}
          </li>
        )
      })}
    </ul>
  )
}
