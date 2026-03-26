import { DollarSign, Activity, Key, Database } from 'lucide-react'
import { StatCard } from '@/components/dashboard/StatCard'
import { ProviderHealth } from '@/components/dashboard/ProviderHealth'
import { RecentRequests } from '@/components/dashboard/RecentRequests'
import { SpendChart } from '@/components/dashboard/SpendChart'
import { ModelBreakdown } from '@/components/analytics/ModelBreakdown'
import { Card } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { formatCurrency, formatNumber, formatPercentage } from '@/lib/utils'
import { useHealth, useUsage, useModels, useCacheStats } from '@/hooks/useAxisApi'

export function Overview() {
  const { data: providerHealth, loading: healthLoading, error: healthError } = useHealth()
  const { data: usageData, loading: usageLoading } = useUsage()
  const { data: modelsData } = useModels()
  const { data: cacheData } = useCacheStats()

  const loading = healthLoading || usageLoading
  const totals = usageData?.totals

  if (loading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  return (
    <div className="space-y-6">
      {/* Hero Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="This Month"
          value={formatCurrency(totals?.costUsd ?? 0)}
          icon={<DollarSign className="h-5 w-5" />}
        />
        <StatCard
          label="Requests Today"
          value={formatNumber(totals?.requests ?? 0)}
          icon={<Activity className="h-5 w-5" />}
        />
        <StatCard
          label="Active Keys"
          value="—"
          icon={<Key className="h-5 w-5" />}
        />
        <StatCard
          label="Cache Hit Rate"
          value={formatPercentage(cacheData?.hitRate ?? 0)}
          icon={<Database className="h-5 w-5" />}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1">
          <SpendChart data={(usageData?.usage ?? []).map((u) => ({ date: u.date, requests: u.requests, cost: u.costUsd }))} />
        </div>
        <div className="lg:col-span-2">
          <ProviderHealth providers={(providerHealth ?? []) as any} />
        </div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ModelBreakdown models={(modelsData ?? []).slice(0, 5).map((m) => ({
          id: m.id,
          name: m.name,
          provider: m.provider,
          requests: m.requests ?? 0,
          cost: 0,
          color: '',
        })) as any} />
        <RecentRequests requests={[]} />
      </div>
    </div>
  )
}
