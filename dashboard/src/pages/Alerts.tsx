import { useState } from 'react'
import { Bell, CheckCircle2, AlertTriangle, TrendingUp } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Toggle } from '@/components/ui/Toggle'
import { getRelativeTime, formatCurrency } from '@/lib/utils'
import { useBudgetAlerts, useAcknowledgeAlert } from '@/hooks/useAxisApi'
import type { BudgetAlert } from '@/api/axis'

function ThresholdBadge({ threshold }: { threshold: number }) {
  const variant =
    threshold >= 100 ? 'danger' :
    threshold >= 80 ? 'warning' :
    'muted'
  
  const color =
    threshold >= 100 ? 'text-danger' :
    threshold >= 80 ? 'text-warning' :
    threshold >= 50 ? 'text-warning' :
    'text-text-muted'

  return (
    <Badge variant={variant} className={color}>
      {threshold}%
    </Badge>
  )
}

function StatusBadge({ acknowledged, threshold }: { acknowledged: boolean; threshold: number }) {
  if (acknowledged) {
    return (
      <Badge variant="muted" className="text-text-muted">
        Acknowledged
      </Badge>
    )
  }
  
  const variant =
    threshold >= 100 ? 'danger' :
    threshold >= 80 ? 'warning' :
    'muted'
  
  const label =
    threshold >= 100 ? 'Critical' :
    threshold >= 80 ? 'Warning' :
    'Info'

  return (
    <Badge variant={variant}>
      {label}
    </Badge>
  )
}

export function Alerts() {
  const [showAcknowledged, setShowAcknowledged] = useState(false)
  const { data: alerts, loading, refetch } = useBudgetAlerts()
  const { acknowledge, loading: acknowledging } = useAcknowledgeAlert()

  const handleAcknowledge = async (alertId: string) => {
    try {
      await acknowledge(alertId)
      await refetch()
    } catch (err) {
      console.error('Failed to acknowledge alert:', err)
    }
  }

  const filteredAlerts = alerts?.filter((alert) => 
    showAcknowledged ? true : !alert.acknowledged
  )

  if (loading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-title">Budget Alerts</h1>
        <p className="text-body text-text-muted mt-1">
          Monitor budget threshold alerts and acknowledge notifications
        </p>
      </div>

      {/* Filter Toggle */}
      <div className="flex items-center gap-3">
        <Toggle
          checked={showAcknowledged}
          onCheckedChange={setShowAcknowledged}
          label="Show acknowledged"
        />
      </div>

      {/* Alert History Table */}
      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Key
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Threshold
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Spent
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Limit
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Triggered At
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Status
                </th>
                <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                  Action
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredAlerts && filteredAlerts.length > 0 ? (
                filteredAlerts.map((alert) => (
                  <tr key={alert.id} className="border-b border-surface-border last:border-b-0 hover:bg-surface-hover">
                    <td className="py-3 px-4">
                      <div className="text-body font-medium text-text-primary">
                        {alert.keyName || alert.keyId}
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <ThresholdBadge threshold={alert.thresholdPercent} />
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary font-mono">
                        {formatCurrency(alert.spentUsd)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary font-mono">
                        {formatCurrency(alert.limitUsd)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-micro text-text-muted">
                        {getRelativeTime(alert.triggeredAt)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <StatusBadge acknowledged={alert.acknowledged} threshold={alert.thresholdPercent} />
                    </td>
                    <td className="py-3 px-4">
                      {!alert.acknowledged && (
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleAcknowledge(alert.id)}
                          disabled={acknowledging}
                        >
                          {acknowledging ? <Spinner size="sm" /> : 'Acknowledge'}
                        </Button>
                      )}
                      {alert.acknowledged && (
                        <CheckCircle2 className="h-4 w-4 text-success" />
                      )}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-body text-text-muted">
                    {showAcknowledged 
                      ? 'No acknowledged alerts' 
                      : 'No active alerts'}
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
