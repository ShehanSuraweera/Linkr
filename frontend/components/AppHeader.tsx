"use client"

import { useState, useRef, useEffect } from "react"
import { useRouter } from "next/navigation"
import { ChevronDown, LogOut } from "lucide-react"
import { useUser } from "@/context/UserContext"

export default function AppHeader() {
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const router = useRouter()
  const { user } = useUser()

  const initial = user?.email?.[0]?.toUpperCase() ?? "?"

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const handleLogout = async () => {
    setDropdownOpen(false)
    await fetch("/api/auth/logout", { method: "POST" })
    router.push("/login")
    router.refresh()
  }

  return (
    <header className="h-18 shrink-0 border-b bg-background hidden lg:flex items-center justify-end px-6">
      <div className="relative" ref={dropdownRef}>
        <button
          onClick={() => setDropdownOpen((v) => !v)}
          className="flex items-center gap-1.5 rounded-full hover:bg-muted p-1 transition-colors"
          aria-label="User menu"
          aria-expanded={dropdownOpen}
        >
          <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center text-primary-foreground text-sm font-bold select-none">
            {initial}
          </div>
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        </button>

        {dropdownOpen && (
          <div className="absolute right-0 top-full mt-2 w-56 bg-popover border border-border rounded-xl shadow-lg overflow-hidden z-50">
            <div className="px-4 py-3 border-b border-border">
              <div className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center text-primary-foreground text-sm font-bold shrink-0">
                  {initial}
                </div>
                <div className="min-w-0">
                  {user?.email
                    ? <p className="text-xs font-medium text-foreground truncate">{user.email}</p>
                    : <div className="h-3 w-28 bg-muted rounded animate-pulse" />}

                  <p className="text-[11px] text-muted-foreground">Free plan</p>
                </div>
              </div>
            </div>

            <div className="py-1">
              <button
                onClick={handleLogout}
                className="flex w-full items-center gap-2.5 px-4 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
              >
                <LogOut className="h-4 w-4 shrink-0" />
                Sign out
              </button>
            </div>
          </div>
        )}
      </div>
    </header>
  )
}
