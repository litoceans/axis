import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { formatCurrency, formatLatency, formatTime, getProviderColor } from '@/lib/utils'
import {} from '@/lib/utils'

interface RequestData {
  id: string
  keyId: string
  keyName: string
  model: string
  provider: string
  status: number
  latency: number
  cost: number
  tokens: number
  timestamp: Date
}

interface RecentRequestsProps {
  requests: RequestData[]
}

export function RecentRequests({ requests }: RecentRequestsProps) {
  const getStatusBadge = (status: number) => {
    if (status >= 200 && status < 300) {
      return <Badge variant="success">{status}</Badge>
    }
    if (status === 429) {
      return <Badge variant="warning">{status}</Badge>
    }
    if (status >= 500) {
      return <Badge variant="danger">{status}</Badge>
    }
    return <Badge variant="muted">{status}</Badge>
  }

  return (
    <Card>
      <h3 className="section-title mb-4">Recent Requests</h3>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="table-header py-3 px-3 text-left">Key</th>
              <th className="table-header py-3 px-3 text-left">Model</th>
              <th className="table-header py-3 px-3 text-left">Status</th>
              <th className="table-header py-3 px-3 text-left">Latency</th>
              <th className="table-header py-3 px-3 text-left">Cost</th>
              <th className="table-header py-3 px-3 text-left">Time</th>
            </tr>
          </thead>
          <tbody>
            {requests.slice(0, 10).map((req) => (
              <tr
                key={req.id}
                className="border-b border-border/50 hover:bg-surface/50 transition-colors"
              >
                <td className="py-2 px-3">
                  <span className="text-cell font-mono text-text-primary">
                    {req.keyName}
                  </span>
                </td>
                <td className="py-2 px-3">
                  <div className="flex items-center gap-2">
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getProviderColor(req.provider) }}
                    />
                    <span className="text-cell text-text-primary font-mono">
                      {req.model.length > 16 ? req.model.slice(0, 16) + '...' : req.model}
                    </span>
                  </div>
                </td>
                <td className="py-2 px-3">{getStatusBadge(req.status)}</td>
                <td className="py-2 px-3">
                  <span className="text-cell font-mono text-text-primary">
                    {formatLatency(req.latency)}
                  </span>
                </td>
                <td className="py-2 px-3">
                  <span className="text-cell font-mono text-text-primary">
                    {formatCurrency(req.cost)}
                  </span>
                </td>
                <td className="py-2 px-3">
                  <span className="text-cell text-text-muted">
                    {formatTime(req.timestamp)}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}
