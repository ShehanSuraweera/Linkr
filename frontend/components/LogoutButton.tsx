"use client"

import { useRouter } from "next/navigation"

export default function LogoutButton() {
  const router = useRouter()

  const handleLogout = async () => {
    await fetch("/api/auth/logout", { method: "POST" })
    router.push("/login")
    router.refresh()
  }

  return (
    <button
      onClick={handleLogout}
      className="text-sm text-gray-500 hover:text-gray-800 transition-colors"
    >
      Sign out
    </button>
  )
}
