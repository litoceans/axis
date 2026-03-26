import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'
import { getProviderColor } from '@/lib/utils'

interface ProviderHealthData {
  provider: string
  name: string
  status: 'healthy' | 'degraded' | 'down'
  uptime: number
  p99Latency: number
  errorRate: number
}

interface ProviderHealthProps {
  providers: ProviderHealthData[]
}

export function ProviderHealth({ providers }: ProviderHealthProps) {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'healthy':
        return <Badge variant="success">Healthy</Badge>
      case 'degraded':
        return <Badge variant="warning">Degraded</Badge>
      case 'down':
        return <Badge variant="danger">Down</Badge>
      default:
        return <Badge variant="muted">{status}</Badge>
    }
  }

  return (
    <Card>
      <h3 className="section-title mb-4">Provider Health</h3>
      <div className="space-y-3">
        {providers.map((provider) => (
          <div
            key={provider.provider}
            className="flex items-center justify-between py-2 border-b border-border/50 last:border-0"
          >
            <div className="flex items-center gap-3">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: getProviderColor(provider.provider) }}
              />
              <span className="text-body font-medium text-text-primary">
                {provider.name}
              </span>
            </div>
            <div className="flex items-center gap-6">
              {getStatusBadge(provider.status)}
              <div className="text-right">
                <p className="text-micro text-text-muted">Uptime</p>
                <p className="text-body font-mono text-text-primary">
                  {provider.uptime.toFixed(1)}%
                </p>
              </div>
              <div className="text-right">
                <p className="text-micro text-text-muted">P99</p>
                <p className="text-body font-mono text-text-primary">
                  {provider.p99Latency >= 1000
                    ? (provider.p99Latency / 1000).toFixed(1) + 's'
                    : provider.p99Latency + 'ms'}
                </p>
              </div>
              <div className="text-right">
                <p className="text-micro text-text-muted">Error Rate</p>
                <p className={cn(
                  'text-body font-mono',
                  provider.errorRate > 1 ? 'text-danger' : 'text-text-primary'
                )}>
                  {provider.errorRate.toFixed(1)}%
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
