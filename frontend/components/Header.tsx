import Link from "next/link"
import LinkrLogoIcon from "@/components/LinkrLogoIcon"
import LogoutButton from "@/components/LogoutButton"

export default function Header() {
  return (
    <header className="bg-background border-b sticky top-0 z-10">
      <div className="max-w-5xl mx-auto px-6 py-3 flex items-center justify-between">
        <Link href="/dashboard" aria-label="Go to dashboard">
          <LinkrLogoIcon />
        </Link>
        <LogoutButton />
      </div>
    </header>
  )
}
