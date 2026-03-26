import { Card } from '@/components/ui/Card'
import { cn } from '@/lib/utils'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

export interface StatCardProps {
  label: string
  value: string | number
  trend?: number
  trendLabel?: string
  icon?: React.ReactNode
  className?: string
}

export function StatCard({ label, value, trend, trendLabel, icon, className }: StatCardProps) {
  const TrendIcon = trend === undefined || trend === 0 ? Minus : trend > 0 ? TrendingUp : TrendingDown
  const trendColor = trend === undefined || trend === 0 ? 'text-text-muted' : trend > 0 ? 'text-success' : 'text-danger'

  return (
    <Card className={cn('relative overflow-hidden', className)}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-micro text-text-muted uppercase tracking-wider mb-1">{label}</p>
          <p className="stat-number text-text-primary">{value}</p>
          {trend !== undefined && (
            <div className={cn('flex items-center gap-1 mt-2', trendColor)}>
              <TrendIcon className="h-3 w-3" />
              <span className="text-micro font-medium">
                {trend > 0 ? '+' : ''}{(trend * 100).toFixed(1)}%
              </span>
              {trendLabel && (
                <span className="text-micro text-text-muted ml-1">{trendLabel}</span>
              )}
            </div>
          )}
        </div>
        {icon && (
          <div className="p-2 rounded-lg bg-accent/10 text-accent">
            {icon}
          </div>
        )}
      </div>
    </Card>
  )
}
