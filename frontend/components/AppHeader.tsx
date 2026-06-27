"use client"

import UserMenu from "@/components/UserMenu"

export default function AppHeader() {
  return (
    <header className="h-18 shrink-0 border-b bg-background hidden lg:flex items-center justify-end px-6">
      <UserMenu />
    </header>
  )
}
