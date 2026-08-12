import { useState } from 'react'
import { useAuthStore } from '@/store/auth'
import { NavLink, Outlet } from 'react-router'
import { Button } from '@/components/ui/button'
import { LogOut, LayoutDashboard, Map, Search, MapPin, X, Menu } from 'lucide-react'
import { cn } from '@/lib/utils'

const NAV_ITEMS = [
  { to: '/app/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/app/wilayah', label: 'Wilayah', icon: Map },
  { to: '/app/search', label: 'Pencarian', icon: Search },
  { to: '/app/reverse', label: 'Reverse Geocode', icon: MapPin },
]

export function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const signOut = useAuthStore((s) => s.signOut)

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    cn(
      'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
      isActive
        ? 'bg-accent text-accent-foreground'
        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
    )

  return (
    <div className="flex h-screen overflow-hidden">
      {mobileOpen && (
        <div className="fixed inset-0 z-30 bg-black/50 md:hidden" onClick={() => setMobileOpen(false)} />
      )}
      <aside
        className={cn(
          'bg-background border-r border-border fixed inset-y-0 left-0 z-40 flex h-screen w-64 flex-col transition-transform duration-300',
          'md:static md:z-auto md:translate-x-0 md:w-64',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <div className="flex h-14 items-center border-b border-border px-4">
          <p className="flex-1 truncate text-sm font-semibold">Geomap Indonesia</p>
          <Button
            variant="ghost"
            size="icon"
            className="ml-auto h-8 w-8 md:hidden"
            onClick={() => setMobileOpen(false)}
            aria-label="Tutup menu"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        <nav className="flex-1 overflow-y-auto py-4 px-2">
          {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} end className={linkClass}>
              <Icon className="h-4 w-4 shrink-0" />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-border p-3">
          <div className="flex items-center gap-2">
            <div className="h-7 w-7 rounded-full bg-muted flex items-center justify-center text-xs font-medium">G</div>
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium truncate">API Dashboard</p>
            </div>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={signOut} aria-label="Keluar">
              <LogOut className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto">
        <div className="flex items-center border-b border-border px-3 py-2 md:hidden">
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setMobileOpen(true)} aria-label="Buka menu">
            <Menu className="h-5 w-5" />
          </Button>
        </div>
        <div className="mx-auto max-w-6xl p-4 md:p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}