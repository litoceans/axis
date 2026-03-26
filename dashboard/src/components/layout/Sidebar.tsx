import { NavLink, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Key,
  BarChart3,
  Boxes,
  Route,
  Database,
  Bell,
  Settings,
  Activity,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { useState } from 'react'

const navItems = [
  { path: '/', label: 'Overview', icon: LayoutDashboard },
  { path: '/keys', label: 'Keys', icon: Key },
  { path: '/analytics', label: 'Analytics', icon: BarChart3 },
  { path: '/models', label: 'Models', icon: Boxes },
  { path: '/routing', label: 'Routing', icon: Route },
  { path: '/cache', label: 'Cache', icon: Database },
  { path: '/alerts', label: 'Alerts', icon: Bell },
  { path: '/traces', label: 'Traces', icon: Activity },
  { path: '/settings', label: 'Settings', icon: Settings },
]

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 h-screen bg-surface border-r border-border',
        'flex flex-col transition-all duration-150 z-40',
        collapsed ? 'w-16' : 'w-56'
      )}
    >
      {/* Logo */}
      <div className="h-14 flex items-center px-4 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <svg viewBox="0 0 24 24" className="w-5 h-5 text-white" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M6 12L12 6L18 12L12 18L6 12Z" strokeLinejoin="round" />
              <circle cx="12" cy="12" r="2" fill="currentColor" />
            </svg>
          </div>
          {!collapsed && (
            <span className="font-semibold text-section text-text-primary">Axis</span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 px-2">
        <ul className="space-y-1">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path
            return (
              <li key={item.path}>
                <NavLink
                  to={item.path}
                  className={cn(
                    'flex items-center gap-3 px-3 py-2 rounded-md text-body',
                    'transition-colors duration-150',
                    isActive
                      ? 'bg-accent/10 text-accent'
                      : 'text-text-muted hover:text-text-primary hover:bg-surface'
                  )}
                >
                  <item.icon className="h-4 w-4 flex-shrink-0" />
                  {!collapsed && <span>{item.label}</span>}
                </NavLink>
              </li>
            )
          })}
        </ul>
      </nav>

      {/* Collapse Button */}
      <div className="p-2 border-t border-border">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-md text-text-muted hover:text-text-primary hover:bg-surface transition-colors"
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <>
              <ChevronLeft className="h-4 w-4" />
              <span className="text-micro">Collapse</span>
            </>
          )}
        </button>
      </div>
    </aside>
  )
}
