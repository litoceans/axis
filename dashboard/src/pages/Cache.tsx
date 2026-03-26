import { useState } from 'react'
import { Trash2, TrendingUp, DollarSign } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { formatCurrency, formatNumber, formatPercentage, getRelativeTime } from '@/lib/utils'
import { useCacheStats } from '@/hooks/useAxisApi'
import { cacheApi } from '@/api/axis'

export function Cache() {
  const [clearing, setClearing] = useState(false)
  const { data: cacheStats, loading } = useCacheStats()

  const handleClearCache = async () => {
    setClearing(true)
    try {
      await cacheApi.clear()
    } finally {
      setClearing(false)
    }
  }

  if (loading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  const stats = cacheStats

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Cache</h1>
          <p className="text-body text-text-muted mt-1">
            Monitor cache performance and manage cached entries
          </p>
        </div>
        <Button variant="danger" onClick={handleClearCache} loading={clearing}>
          <Trash2 className="h-4 w-4" />
          Clear Cache
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <p className="text-micro text-text-muted uppercase tracking-wider">Hit Rate</p>
          <p className="stat-number text-text-primary mt-1">{formatPercentage(stats?.hitRate ?? 0)}</p>
        </Card>

        <Card>
          <p className="text-micro text-text-muted uppercase tracking-wider">Estimated Savings</p>
          <p className="stat-number text-text-primary mt-1">{formatCurrency(stats?.savings ?? 0)}</p>
        </Card>

        <Card>
          <p className="text-micro text-text-muted uppercase tracking-wider">Total Hits</p>
          <p className="stat-number text-text-primary mt-1">{formatNumber(stats?.totalHits ?? 0)}</p>
          <p className="text-micro text-text-muted mt-2">vs {formatNumber(stats?.totalMisses ?? 0)} misses</p>
        </Card>
      </div>

      <Card>
        <h3 className="section-title mb-4">Top Cached Prompts</h3>
        <div className="space-y-3">
          {(stats?.topCachedPrompts ?? []).map((prompt, index) => (
            <div key={prompt.hash} className="flex items-center justify-between py-3 border-b border-border/50 last:border-0">
              <div className="flex items-center gap-4">
                <span className="text-micro text-text-muted w-6">#{index + 1}</span>
                <code className="text-body font-mono text-text-primary bg-bg px-3 py-1 rounded">
                  {prompt.hash.length > 8 ? prompt.hash.slice(0, 8) : prompt.hash}
                </code>
              </div>
              <div className="flex items-center gap-6">
                <span className="text-body text-text-muted">{formatNumber(prompt.count)} hits</span>
                <span className="text-micro text-text-muted">{getRelativeTime(new Date(prompt.lastUsed))}</span>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
