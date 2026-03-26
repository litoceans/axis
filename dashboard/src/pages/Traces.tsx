import { useState } from 'react'
import { Activity, ChevronRight, ChevronDown, Clock, CheckCircle, AlertCircle } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { getRelativeTime } from '@/lib/utils'

interface TraceSpan {
  id: string
  operation: string
  duration: number
  status: 'success' | 'error' | 'pending'
  timestamp: string
  parentId?: string
  children?: TraceSpan[]
  model?: string
  provider?: string
  tokens?: number
  cost?: number
}

interface Trace {
  id: string
  operation: string
  duration: number
  status: 'success' | 'error' | 'pending'
  timestamp: string
  spans: TraceSpan[]
}

// Mock data for now - will be replaced with real API
const mockTraces: Trace[] = [
  {
    id: 'trace_001',
    operation: 'chat.completions.create',
    duration: 1250,
    status: 'success',
    timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    spans: [
      { id: 'span_001', operation: 'router.select', duration: 45, status: 'success', timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString() },
      { id: 'span_002', operation: 'provider.request', duration: 1150, status: 'success', timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(), parentId: 'span_001', model: 'gpt-4o-mini', provider: 'openai', tokens: 1024, cost: 0.0002 },
      { id: 'span_003', operation: 'cache.lookup', duration: 12, status: 'success', timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(), parentId: 'span_001' },
    ],
  },
  {
    id: 'trace_002',
    operation: 'chat.completions.create',
    duration: 2340,
    status: 'success',
    timestamp: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
    spans: [
      { id: 'span_004', operation: 'router.select', duration: 52, status: 'success', timestamp: new Date(Date.now() - 10 * 60 * 1000).toISOString() },
      { id: 'span_005', operation: 'provider.request', duration: 2200, status: 'success', timestamp: new Date(Date.now() - 10 * 60 * 1000).toISOString(), parentId: 'span_004', model: 'claude-3-5-sonnet', provider: 'anthropic', tokens: 2048, cost: 0.0055 },
    ],
  },
  {
    id: 'trace_003',
    operation: 'embeddings.create',
    duration: 450,
    status: 'error',
    timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
    spans: [
      { id: 'span_006', operation: 'provider.request', duration: 400, status: 'error', timestamp: new Date(Date.now() - 15 * 60 * 1000).toISOString(), model: 'text-embedding-3-small', provider: 'openai' },
    ],
  },
]

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'success':
      return <CheckCircle className="h-4 w-4 text-success" />
    case 'error':
      return <AlertCircle className="h-4 w-4 text-danger" />
    default:
      return <Clock className="h-4 w-4 text-text-muted" />
  }
}

function StatusBadge({ status }: { status: string }) {
  const variant = status === 'success' ? 'success' : status === 'error' ? 'danger' : 'muted'
  return <Badge variant={variant}>{status}</Badge>
}

function SpanTimeline({ spans }: { spans: TraceSpan[] }) {
  const [expandedSpans, setExpandedSpans] = useState<Set<string>>(new Set())

  const toggleSpan = (spanId: string) => {
    const newExpanded = new Set(expandedSpans)
    if (newExpanded.has(spanId)) {
      newExpanded.delete(spanId)
    } else {
      newExpanded.add(spanId)
    }
    setExpandedSpans(newExpanded)
  }

  const rootSpans = spans.filter((s) => !s.parentId)
  const childSpans = spans.filter((s) => s.parentId)

  const renderSpan = (span: TraceSpan, depth = 0) => {
    const children = childSpans.filter((s) => s.parentId === span.id)
    const isExpanded = expandedSpans.has(span.id)
    const hasChildren = children.length > 0

    return (
      <div key={span.id}>
        <div
          className={`flex items-center gap-2 py-2 px-3 hover:bg-surface-hover rounded-md cursor-pointer ${depth > 0 ? 'ml-6' : ''}`}
          onClick={() => hasChildren && toggleSpan(span.id)}
        >
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="h-4 w-4 text-text-muted" />
            ) : (
              <ChevronRight className="h-4 w-4 text-text-muted" />
            )
          ) : (
            <span className="w-4" />
          )}
          <StatusIcon status={span.status} />
          <span className="text-body text-text-primary flex-1">{span.operation}</span>
          <span className="text-micro text-text-muted font-mono">{span.duration}ms</span>
          {span.model && (
            <Badge variant="muted" className="text-micro">
              {span.model}
            </Badge>
          )}
        </div>
        {isExpanded && children.map((child) => renderSpan(child, depth + 1))}
      </div>
    )
  }

  return (
    <div className="space-y-1">
      <h4 className="text-micro font-semibold text-text-muted uppercase tracking-wider mb-2">Span Timeline</h4>
      {rootSpans.map((span) => renderSpan(span))}
    </div>
  )
}

function TraceDetail({ trace, onClose }: { trace: Trace; onClose: () => void }) {
  return (
    <Card className="p-6">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-section font-semibold text-text-primary mb-1">Trace Details</h3>
          <p className="text-micro font-mono text-text-muted">{trace.id}</p>
        </div>
        <button onClick={onClose} className="text-text-muted hover:text-text-primary">
          ✕
        </button>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div>
          <p className="text-micro text-text-muted">Operation</p>
          <p className="text-body text-text-primary font-medium">{trace.operation}</p>
        </div>
        <div>
          <p className="text-micro text-text-muted">Duration</p>
          <p className="text-body text-text-primary font-medium">{trace.duration}ms</p>
        </div>
        <div>
          <p className="text-micro text-text-muted">Status</p>
          <div className="mt-1">
            <StatusBadge status={trace.status} />
          </div>
        </div>
        <div>
          <p className="text-micro text-text-muted">Timestamp</p>
          <p className="text-body text-text-primary font-medium">{getRelativeTime(trace.timestamp)}</p>
        </div>
      </div>

      <SpanTimeline spans={trace.spans} />
    </Card>
  )
}

export function Traces() {
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null)
  const [loading] = useState(false)

  if (loading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-title">Request Traces</h1>
        <p className="text-body text-text-muted mt-1">
          View detailed request traces and span timelines
        </p>
      </div>

      {selectedTrace ? (
        <TraceDetail trace={selectedTrace} onClose={() => setSelectedTrace(null)} />
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Trace ID
                  </th>
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Operation
                  </th>
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Duration
                  </th>
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Status
                  </th>
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Timestamp
                  </th>
                  <th className="text-left text-micro font-semibold text-text-muted uppercase tracking-wider py-3 px-4">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody>
                {mockTraces.map((trace) => (
                  <tr key={trace.id} className="border-b border-surface-border last:border-b-0 hover:bg-surface-hover">
                    <td className="py-3 px-4">
                      <span className="text-micro font-mono text-text-primary">{trace.id}</span>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <Activity className="h-4 w-4 text-accent" />
                        <span className="text-body text-text-primary">{trace.operation}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-body text-text-primary font-mono">{trace.duration}ms</span>
                    </td>
                    <td className="py-3 px-4">
                      <StatusBadge status={trace.status} />
                    </td>
                    <td className="py-3 px-4">
                      <span className="text-micro text-text-muted">{getRelativeTime(trace.timestamp)}</span>
                    </td>
                    <td className="py-3 px-4">
                      <Button variant="secondary" size="sm" onClick={() => setSelectedTrace(trace)}>
                        View Details
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}
    </div>
  )
}
