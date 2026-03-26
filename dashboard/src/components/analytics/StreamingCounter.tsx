import { useEffect, useState } from 'react'
import { Activity, Zap } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'

interface TokenStream {
  requestId: string
  model: string
  tokensIn: number
  tokensOut: number
  totalTokens: number
  isStreaming: boolean
  startTime: number
  lastUpdate: number
}

// Placeholder data - will be replaced with real WebSocket streaming
const mockStreams: TokenStream[] = [
  {
    requestId: 'req_stream_001',
    model: 'gpt-4o-mini',
    tokensIn: 256,
    tokensOut: 512,
    totalTokens: 768,
    isStreaming: true,
    startTime: Date.now() - 5000,
    lastUpdate: Date.now(),
  },
  {
    requestId: 'req_stream_002',
    model: 'claude-3-5-sonnet',
    tokensIn: 1024,
    tokensOut: 2048,
    totalTokens: 3072,
    isStreaming: true,
    startTime: Date.now() - 12000,
    lastUpdate: Date.now() - 100,
  },
  {
    requestId: 'req_stream_003',
    model: 'gemini-1.5-flash',
    tokensIn: 512,
    tokensOut: 128,
    totalTokens: 640,
    isStreaming: false,
    startTime: Date.now() - 30000,
    lastUpdate: Date.now() - 25000,
  },
]

interface StreamStats {
  activeStreams: number
  totalTokensPerSecond: number
  avgLatency: number
}

function calculateStats(streams: TokenStream[]): StreamStats {
  const activeStreams = streams.filter((s) => s.isStreaming).length
  const totalTokens = streams.reduce((sum, s) => sum + s.totalTokens, 0)
  const totalDuration = streams.reduce((sum, s) => sum + (Date.now() - s.startTime), 0) / streams.length
  
  return {
    activeStreams,
    totalTokensPerSecond: Math.round(totalTokens / (totalDuration / 1000)),
    avgLatency: Math.round(totalDuration / streams.length),
  }
}

export function StreamingCounter() {
  const [streams, setStreams] = useState<TokenStream[]>(mockStreams)
  const stats = calculateStats(streams)

  // Simulate live token aggregation (placeholder for real WebSocket)
  useEffect(() => {
    const interval = setInterval(() => {
      setStreams((prev) =>
        prev.map((stream) => {
          if (!stream.isStreaming) return stream
          
          // Simulate token accumulation
          const newTokensOut = stream.tokensOut + Math.floor(Math.random() * 10)
          return {
            ...stream,
            tokensOut: newTokensOut,
            totalTokens: stream.tokensIn + newTokensOut,
            lastUpdate: Date.now(),
          }
        })
      )
    }, 500)

    return () => clearInterval(interval)
  }, [])

  return (
    <Card className="p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-accent" />
          <h3 className="text-section font-semibold text-text-primary">Live Token Stream</h3>
        </div>
        <Badge variant={stats.activeStreams > 0 ? 'success' : 'muted'}>
          {stats.activeStreams} active
        </Badge>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-3 gap-4 mb-4">
        <div className="text-center">
          <p className="text-micro text-text-muted">Active Streams</p>
          <p className="text-section font-bold text-text-primary">{stats.activeStreams}</p>
        </div>
        <div className="text-center">
          <p className="text-micro text-text-muted">Tokens/sec</p>
          <p className="text-section font-bold text-accent">{stats.totalTokensPerSecond}</p>
        </div>
        <div className="text-center">
          <p className="text-micro text-text-muted">Avg Latency</p>
          <p className="text-section font-bold text-text-primary">{stats.avgLatency}ms</p>
        </div>
      </div>

      {/* Active Streams List */}
      <div className="space-y-3">
        {streams.map((stream) => (
          <div
            key={stream.requestId}
            className="flex items-center justify-between p-3 rounded-md bg-surface hover:bg-surface-hover transition-colors"
          >
            <div className="flex items-center gap-3">
              {stream.isStreaming ? (
                <Zap className="h-4 w-4 text-accent animate-pulse" />
              ) : (
                <Activity className="h-4 w-4 text-text-muted" />
              )}
              <div>
                <p className="text-body font-medium text-text-primary">{stream.model}</p>
                <p className="text-micro text-text-muted font-mono">{stream.requestId}</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-micro text-text-muted">Tokens</p>
                <p className="text-body font-mono text-text-primary">
                  {stream.tokensIn} → {stream.tokensOut}
                </p>
              </div>
              <div className="text-right">
                <p className="text-micro text-text-muted">Total</p>
                <p className="text-body font-mono text-accent">{stream.totalTokens}</p>
              </div>
              {stream.isStreaming && (
                <Badge variant="success" className="text-micro">
                  Streaming
                </Badge>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Note about WebSocket integration */}
      <p className="text-micro text-text-muted mt-4 pt-3 border-t border-surface-border">
        📡 Placeholder data — WebSocket streaming to be wired in Phase 3 backend integration
      </p>
    </Card>
  )
}
