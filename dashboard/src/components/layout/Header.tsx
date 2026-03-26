import { Search, Bell, User } from 'lucide-react'
import { useState } from 'react'

export function Header() {
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <header className="h-14 bg-surface border-b border-border flex items-center justify-between px-6">
      <div className="flex items-center gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
          <input
            type="text"
            placeholder="Search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input pl-9 w-64 text-micro"
          />
        </div>
      </div>

      <div className="flex items-center gap-4">
        <button className="relative p-2 rounded-md text-text-muted hover:text-text-primary hover:bg-surface transition-colors">
          <Bell className="h-4 w-4" />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-danger rounded-full" />
        </button>
        <button className="flex items-center gap-2 p-1.5 rounded-md hover:bg-surface transition-colors">
          <div className="w-8 h-8 rounded-full bg-accent/20 flex items-center justify-center">
            <User className="h-4 w-4 text-accent" />
          </div>
        </button>
      </div>
    </header>
  )
}
