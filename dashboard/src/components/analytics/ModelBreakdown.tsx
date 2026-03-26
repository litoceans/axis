import { Card } from '@/components/ui/Card'
import { formatCurrency, getProviderColor } from '@/lib/utils'

interface ModelBreakdownProps {
  models: Array<{
    id: string
    name: string
    provider: string
    requests: number
    cost: number
    color: string
  }>
}

export function ModelBreakdown({ models }: ModelBreakdownProps) {
  const totalCost = models.reduce((sum, m) => sum + m.cost, 0)

  return (
    <Card>
      <h3 className="section-title mb-4">Top Models by Spend</h3>
      <div className="space-y-3">
        {models.slice(0, 5).map((model) => (
          <div key={model.id} className="space-y-1">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div
                  className="w-2 h-2 rounded-full"
                  style={{ backgroundColor: model.color || getProviderColor(model.provider) }}
                />
                <span className="text-body font-medium text-text-primary">
                  {model.name}
                </span>
              </div>
              <div className="flex items-center gap-4">
                <span className="text-body font-mono text-text-primary">
                  {formatCurrency(model.cost)}
                </span>
                <span className="text-micro text-text-muted w-16 text-right">
                  {((model.cost / totalCost) * 100).toFixed(1)}%
                </span>
              </div>
            </div>
            <div className="h-1.5 bg-bg rounded-full overflow-hidden">
              <div
                className="h-full rounded-full transition-all duration-300"
                style={{
                  width: `${(model.cost / totalCost) * 100}%`,
                  backgroundColor: model.color || getProviderColor(model.provider),
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
