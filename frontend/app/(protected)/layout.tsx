import Sidebar from "@/components/Sidebar"
import AppHeader from "@/components/AppHeader"
import { UserProvider } from "@/context/UserContext"
import { getMe } from "@/lib/api"

export default async function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const user = await getMe().catch(() => null)

  return (
    <UserProvider initialUser={user}>
      <div className="flex h-screen overflow-hidden bg-muted/30">
        <Sidebar />
        <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
          {/* Mobile top-bar offset — AppHeader sits above on desktop */}
          <div className="lg:hidden h-14 shrink-0" />
          <AppHeader />
          <main className="flex-1 overflow-y-auto">
            <div className="max-w-5xl mx-auto px-6 py-8">
              {children}
            </div>
          </main>
        </div>
      </div>
    </UserProvider>
  )
}
