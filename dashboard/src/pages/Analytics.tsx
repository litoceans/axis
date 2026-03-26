import { useState } from 'react'
import { Download, TrendingUp, TrendingDown } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Toggle } from '@/components/ui/Toggle'
import { AnalyticsSpendChart } from '@/components/analytics/SpendChart'
import { LatencyChart } from '@/components/analytics/LatencyChart'
import { StreamingCounter } from '@/components/analytics/StreamingCounter'
import { formatCurrency } from '@/lib/utils'
import { useUsage, useCosts, useCostReconciliation } from '@/hooks/useAxisApi'
import type { CostReconciliation } from '@/api/axis'

export function Analytics() {
  const [dateRange, setDateRange] = useState('30d')
  const [modelFilter, setModelFilter] = useState('all')
  const [providerFilter, setProviderFilter] = useState('all')

  const { data: usageData, loading: usageLoading } = useUsage()
  const { data: costsData } = useCosts({ by: 'day' })

  const usage = usageData?.usage ?? []
  const totalCost = usage.reduce((sum, u) => sum + u.costUsd, 0)

  // Simple cost prediction: extrapolate based on days with data
  const daysWithData = usage.length
  const avgDailyCost = daysWithData > 0 ? totalCost / daysWithData : 0
  const daysInMonth = 30
  const predictedMonthSpend = avgDailyCost * daysInMonth
  const currentMonthSpend = totalCost
  const costTrend = predictedMonthSpend > currentMonthSpend ? 'up' : 'down'

  const handleExport = () => {
    const csv = [
      ['Date', 'Requests', 'Cost', 'Tokens In', 'Tokens Out'],
      ...usage.map((u) => [
        u.date,
        u.requests,
        u.costUsd,
        u.tokensIn,
        u.tokensOut,
      ]),
    ]
      .map((row) => row.join(','))
      .join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `axis-analytics-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Analytics</h1>
          <p className="text-body text-text-muted mt-1">
            Track usage, costs, and performance
          </p>
        </div>
        <Button variant="secondary" onClick={handleExport}>
          <Download className="h-4 w-4" />
          Export CSV
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 flex-wrap">
        <select
          className="input w-40"
          value={dateRange}
          onChange={(e) => setDateRange(e.target.value)}
        >
          <option value="7d">Last 7 days</option>
          <option value="30d">Last 30 days</option>
          <option value="90d">Last 90 days</option>
        </select>
        <select
          className="input w-40"
          value={modelFilter}
          onChange={(e) => setModelFilter(e.target.value)}
        >
          <option value="all">All Models</option>
          <option value="gpt-4o-mini">GPT-4o Mini</option>
          <option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
        </select>
        <select
          className="input w-40"
          value={providerFilter}
          onChange={(e) => setProviderFilter(e.target.value)}
        >
          <option value="all">All Providers</option>
          <option value="openai">OpenAI</option>
          <option value="anthropic">Anthropic</option>
          <option value="google">Google</option>
        </select>
      </div>

      {/* Cost Prediction */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-micro text-text-muted uppercase tracking-wider">
              Cost Prediction
            </p>
            <p className="stat-number mt-1">
              {formatCurrency(predictedMonthSpend)}
            </p>
            <p className="text-body text-text-muted mt-1">
              Based on {daysWithData} days of data
            </p>
          </div>
          <div className={`flex items-center gap-2 ${costTrend === 'up' ? 'text-danger' : 'text-success'}`}>
            {costTrend === 'up' ? (
              <TrendingUp className="h-5 w-5" />
            ) : (
              <TrendingDown className="h-5 w-5" />
            )}
            <span className="text-section font-semibold">
              {costTrend === 'up' ? '+' : ''}
              {daysWithData > 0 && currentMonthSpend > 0
                ? ((predictedMonthSpend - currentMonthSpend) / currentMonthSpend * 100).toFixed(1)
                : '0.0'}%
            </span>
          </div>
        </div>
      </Card>

      {/* Streaming Token Counter */}
      <StreamingCounter />

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <AnalyticsSpendChart data={usage.map((u) => ({ date: u.date, cost: u.costUsd }))} />
        <LatencyChart data={[]} />
      </div>

      {/* Cost Reconciliation */}
      <CostReconciliationSection />
    </div>
  )
}

function CostReconciliationSection() {
  const [showDiscrepanciesOnly, setShowDiscrepanciesOnly] = useState(false)
  const { data: reconciliations, loading } = useCostReconciliation({ showDiscrepanciesOnly })

  const filteredData = showDiscrepanciesOnly
    ? reconciliations?.filter((r) => r.status === 'discrepancy')
    : reconciliations

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="section-title">Cost Reconciliation</h2>
        <Toggle
          checked={showDiscrepanciesOnly}
          onCheckedChange={setShowDiscrepanciesOnly}
          label="Show discrepancies only"
        />
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Request ID
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Model
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Provider
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Reported Cost
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Calculated Cost
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Difference
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Status
                </th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-body text-text-muted">
                    Loading...
                  </td>
                </tr>
              ) : filteredData && filteredData.length > 0 ? (
                filteredData.map((rec) => (
                  <tr key={rec.requestId} className="border-b border-surface-border last:border-b-0 hover:bg-surface-hover">
                    <td className="py-3 px-4">
                      <span className="text-micro font-mono text-text-primary">{rec.requestId}</span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary">{rec.model}</span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-secondary">{rec.provider}</span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary font-mono">
                        {formatCurrency(rec.reportedCost)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary font-mono">
                        {formatCurrency(rec.calculatedCost)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={`text-body font-mono ${rec.difference >= 0 ? 'text-warning' : 'text-warning'}`}>
                        {rec.difference >= 0 ? '+' : ''}{formatCurrency(rec.difference)} ({rec.differencePercent.toFixed(2)}%)
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <Badge variant={rec.status === 'match' ? 'success' : 'warning'}>
                        {rec.status === 'match' ? 'Match' : 'Discrepancy'}
                      </Badge>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-body text-text-muted">
                    {showDiscrepanciesOnly
                      ? 'No discrepancies found'
                      : 'No reconciliation data available'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}
