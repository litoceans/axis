import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { formatCurrency, formatNumber, formatLatency, getProviderColor } from '@/lib/utils'
import { useModels, useHealth } from '@/hooks/useAxisApi'

const CONTEXT_SIZES: Record<string, number> = {
  'gpt-4o': 128000, 'gpt-4o-mini': 128000, 'gpt-4-turbo': 128000,
  'claude-3-5-sonnet': 200000, 'claude-3-5-haiku': 200000,
  'gemini-1.5-pro': 1000000, 'gemini-1.5-flash': 1000000,
  'llama3.3': 128000,
}

function formatContextTokens(n: number) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(0)}M`
  if (n >= 1_000) return `${Math.round(n / 1_000)}K`
  return String(n)
}

function isSmallContext(modelId: string, ctx?: number) {
  const size = ctx ?? CONTEXT_SIZES[modelId] ?? 0
  return size > 0 && size < 50_000
}

export function Models() {
  const [filter, setFilter] = useState('all')
  const { data: models, loading: modelsLoading } = useModels()
  const { data: healthData } = useHealth()

  // Map latency from health data by model id
  const latencyByModel: Record<string, number> = {}
  if (healthData) {
    // Provider health doesn't include per-model latency, use provider-level p99
    healthData.forEach((h) => {
      latencyByModel[h.provider] = h.p99Latency
    })
  }

  const filteredModels = filter === 'all'
    ? (models ?? [])
    : (models ?? []).filter((m) => m.provider === filter)

  if (modelsLoading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Models</h1>
          <p className="text-body text-text-muted mt-1">
            All available models and their pricing
          </p>
        </div>
        <select
          className="input w-40"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        >
          <option value="all">All Providers</option>
          <option value="openai">OpenAI</option>
          <option value="anthropic">Anthropic</option>
          <option value="google">Google</option>
          <option value="ollama">Ollama</option>
        </select>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredModels.map((model) => (
          <Card key={model.id} hover>
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: getProviderColor(model.provider) }}
                />
                <h3 className="text-body font-semibold text-text-primary font-mono">
                  {model.name}
                </h3>
              </div>
              <Badge variant="muted">{model.provider}</Badge>
            </div>

            <div className="space-y-2">
              <div className="flex justify-between text-body">
                <span className="text-text-muted">Input</span>
                <span className="font-mono text-text-primary">
                  {model.inputPrice === 0 ? 'Free' : `$${model.inputPrice}/1M`}
                </span>
              </div>
              <div className="flex justify-between text-body">
                <span className="text-text-muted">Output</span>
                <span className="font-mono text-text-primary">
                  {model.outputPrice === 0 ? 'Free' : `$${model.outputPrice}/1M`}
                </span>
              </div>
              <div className="flex justify-between text-body">
                <span className="text-text-muted">P50 Latency</span>
                <span className="font-mono text-text-primary">
                  {formatLatency(model.latencyP50 ?? latencyByModel[model.provider])}
                </span>
              </div>
              <div className="flex justify-between text-body">
                <span className="text-text-muted">Requests</span>
                <span className="font-mono text-text-primary">
                  {formatNumber(model.requests ?? 0)}
                </span>
              </div>
              <div className="flex justify-between text-body items-center">
                <span className="text-text-muted">Context</span>
                <div className="flex items-center gap-2">
                  {isSmallContext(model.id, model.maxContextTokens) && (
                    <Badge variant="warning">Small</Badge>
                  )}
                  <span className="font-mono text-text-primary">
                    {formatContextTokens(model.maxContextTokens ?? CONTEXT_SIZES[model.id] ?? 0) || '—'}
                  </span>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  )
}
