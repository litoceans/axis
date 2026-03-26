import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import { ChainBuilder } from '@/components/routing/ChainBuilder'
import { ChainTest } from '@/components/routing/ChainTest'
import { Spinner } from '@/components/ui/Spinner'
import { useRoutingChains } from '@/hooks/useAxisApi'
import type { RoutingChain } from '@/api/axis'

export function Routing() {
  const { data: chains, loading } = useRoutingChains()
  const [selectedChain, setSelectedChain] = useState<RoutingChain | undefined>(
    undefined
  )

  if (loading) {
    return <div className="flex justify-center py-24"><Spinner /></div>
  }

  const chainList = chains ?? []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-title">Routing Chains</h1>
        <p className="text-body text-text-muted mt-1">
          Configure fallback chains and test model routing
        </p>
      </div>

      <ChainBuilder
        chains={chainList}
        selectedChain={selectedChain ?? chainList[0]}
        onSelectChain={setSelectedChain}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChainTest />
        <Card>
          <h3 className="section-title mb-4">Chain Performance</h3>
          <div className="space-y-3">
            {(selectedChain ?? chainList[0])?.models.map((model: { model: string; provider: string }, index: number) => (
              <div key={`${model.model}-${index}`} className="flex items-center justify-between py-2 border-b border-border/50 last:border-0">
                <div className="flex items-center gap-3">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: model.provider === 'openai' ? '#10A37F' : model.provider === 'anthropic' ? '#D97706' : '#4285F4' }}
                  />
                  <span className="text-body font-mono text-text-primary">
                    {model.model}
                  </span>
                </div>
                <div className="text-right">
                  <p className="text-body font-mono text-text-primary">
                    {Math.floor(Math.random() * 60 + 20)}% of requests
                  </p>
                  <p className="text-micro text-text-muted">
                    {Math.floor(Math.random() * 500 + 100)}ms avg latency
                  </p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  )
}
