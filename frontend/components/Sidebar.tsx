"use client"

import { useState } from "react"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { Link2, BarChart2, Settings, Menu, X, Plus, ChevronLeft } from "lucide-react"
import LinkrLogoIcon from "@/components/LinkrLogoIcon"
import { cn } from "@/lib/utils"

const navItems = [
  { href: "/dashboard", icon: Link2, label: "Links" },
  { href: "/analytics", icon: BarChart2, label: "Analytics", disabled: true },
  { href: "/settings", icon: Settings, label: "Settings", disabled: true },
]

interface NavContentProps {
  pathname: string
  collapsed: boolean
  onNavigate?: () => void
  onToggleCollapse?: () => void
  isMobile?: boolean
}

function NavContent({ pathname, collapsed, onNavigate, onToggleCollapse, isMobile }: NavContentProps) {
  const router = useRouter()

  return (
    <div className="flex flex-col h-full">
      {/* Logo + collapse toggle */}
      <div className={cn(
        "flex items-center border-b border-sidebar-border shrink-0",
        collapsed ? "justify-center px-3 py-4" : "justify-between px-4 py-4"
      )}>
        {!collapsed && (
          <Link href="/dashboard" onClick={onNavigate} className="inline-block">
            <LinkrLogoIcon />
          </Link>
        )}
        {!isMobile && (
          <button
            onClick={onToggleCollapse}
            className="text-sidebar-foreground/50 hover:text-sidebar-foreground transition-colors"
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            <ChevronLeft className={cn("h-4 w-4 transition-transform", collapsed && "rotate-180")} />
          </button>
        )}
        {isMobile && (
          <button
            onClick={onNavigate}
            className="text-sidebar-foreground/50 hover:text-sidebar-foreground ml-auto"
            aria-label="Close menu"
          >
            <X className="h-5 w-5" />
          </button>
        )}
      </div>

      {/* Create new button */}
      <div className={cn("px-3 py-4 shrink-0", collapsed && "flex justify-center")}>
        <Link
          href="/dashboard?new=1"
          onClick={onNavigate}
          className={cn(
            "flex items-center justify-center gap-2 bg-foreground text-background text-sm font-semibold rounded-lg transition-colors hover:bg-foreground/80",
            collapsed ? "h-9 w-9" : "w-full px-4 py-2"
          )}
        >
          <Plus className="h-4 w-4 shrink-0" />
          {!collapsed && "Create new"}
        </Link>
      </div>

      {/* Nav items */}
      <nav className="flex-1 px-3 space-y-0.5 overflow-y-auto">
        {navItems.map(({ href, icon: Icon, label, disabled }) => {
          const isActive = pathname === href || (!disabled && pathname.startsWith(href + "/"))

          if (disabled) {
            return (
              <span
                key={href}
                title={collapsed ? label : undefined}
                className={cn(
                  "flex items-center rounded-lg text-sm font-medium text-sidebar-foreground/30 cursor-not-allowed select-none",
                  collapsed ? "justify-center px-2 py-2.5" : "gap-3 px-3 py-2.5"
                )}
              >
                <Icon className="h-4 w-4 shrink-0" />
                {!collapsed && (
                  <>
                    {label}
                    <span className="ml-auto text-[10px] font-normal bg-sidebar-accent text-sidebar-accent-foreground px-1.5 py-0.5 rounded">
                      Soon
                    </span>
                  </>
                )}
              </span>
            )
          }

          return (
            <Link
              key={href}
              href={href}
              onClick={onNavigate}
              title={collapsed ? label : undefined}
              className={cn(
                "flex items-center rounded-lg text-sm font-medium transition-colors",
                collapsed ? "justify-center px-2 py-2.5" : "gap-3 px-3 py-2.5",
                isActive
                  ? "bg-sidebar-accent text-sidebar-accent-foreground"
                  : "text-sidebar-foreground/60 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground"
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {!collapsed && label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}

export default function Sidebar() {
  const pathname = usePathname()
  const [collapsed, setCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <>
      {/* Desktop sidebar */}
      <aside
        className={cn(
          "hidden lg:flex shrink-0 flex-col h-screen sticky top-0 bg-sidebar border-r border-sidebar-border text-sidebar-foreground transition-all duration-200",
          collapsed ? "w-16" : "w-60"
        )}
      >
        <NavContent
          pathname={pathname}
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed((v) => !v)}
        />
      </aside>

      {/* Mobile topbar */}
      <div className="lg:hidden fixed top-0 inset-x-0 z-40 h-14 bg-background border-b flex items-center px-4 gap-3">
        <button
          onClick={() => setMobileOpen(true)}
          className="text-muted-foreground hover:text-foreground"
          aria-label="Open menu"
        >
          <Menu className="h-5 w-5" />
        </button>
        <Link href="/dashboard">
          <LinkrLogoIcon />
        </Link>
      </div>

      {/* Mobile slide-in */}
      {mobileOpen && (
        <div className="lg:hidden fixed inset-0 z-50 flex">
          <div className="w-64 bg-sidebar border-r border-sidebar-border flex flex-col text-sidebar-foreground">
            <NavContent
              pathname={pathname}
              collapsed={false}
              isMobile
              onNavigate={() => setMobileOpen(false)}
            />
          </div>
          <div
            className="flex-1 bg-black/40 backdrop-blur-sm"
            onClick={() => setMobileOpen(false)}
          />
        </div>
      )}
    </>
  )
}
