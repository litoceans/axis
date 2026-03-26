import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { formatLatency, formatCurrency, getProviderColor } from '@/lib/utils'

interface TestResult {
  model: string
  provider: string
  latency: number
  cost: number
  success: boolean
  error?: string
}

interface ChainTestProps {
  selectedModel?: string
  prompt?: string
  onPromptChange?: (prompt: string) => void
}

export function ChainTest({ prompt, onPromptChange }: ChainTestProps) {
  const [loading, setLoading] = useState(false)
  const [results, setResults] = useState<TestResult[]>([])

  const handleTest = async () => {
    setLoading(true)
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1500))
    setResults([
      { model: 'gpt-4o-mini', provider: 'openai', latency: 420, cost: 0.00023, success: true },
      { model: 'claude-3-5-haiku', provider: 'anthropic', latency: 890, cost: 0.00112, success: true },
      { model: 'gemini-1.5-flash', provider: 'google', latency: 680, cost: 0.00009, success: true },
    ])
    setLoading(false)
  }

  return (
    <Card>
      <h3 className="section-title mb-4">Test Panel</h3>

      <div className="space-y-4">
        <div>
          <label className="label">Test Prompt</label>
          <textarea
            className="input min-h-[120px] resize-none"
            placeholder="Enter a test prompt to see which model handles it..."
            value={prompt}
            onChange={(e) => onPromptChange?.(e.target.value)}
          />
        </div>

        <Button onClick={handleTest} loading={loading} className="w-full">
          Run Test
        </Button>

        {results.length > 0 && (
          <div className="space-y-3 mt-4">
            <p className="text-micro text-text-muted uppercase tracking-wider">
              Results
            </p>
            {results.map((result, index) => (
              <div
                key={result.model}
                className="flex items-center gap-4 p-3 bg-bg rounded-lg border border-border"
              >
                <div className="flex items-center justify-center w-6 h-6 rounded-full bg-accent/10 text-accent text-micro font-mono">
                  {index + 1}
                </div>

                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getProviderColor(result.provider) }}
                    />
                    <span className="text-body font-mono text-text-primary">
                      {result.model}
                    </span>
                  </div>
                </div>

                <div className="text-right">
                  <p className="text-body font-mono text-text-primary">
                    {formatLatency(result.latency)}
                  </p>
                  <p className="text-micro text-text-muted">
                    {formatCurrency(result.cost)}
                  </p>
                </div>

                {result.success ? (
                  <Badge variant="success">Success</Badge>
                ) : (
                  <Badge variant="danger">{result.error}</Badge>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </Card>
  )
}
