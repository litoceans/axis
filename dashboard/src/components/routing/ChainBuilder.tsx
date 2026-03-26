import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { getProviderColor } from '@/lib/utils'

interface ChainModel {
  model: string
  provider: string
  maxLatency: number
  retries: number
  weight: number
  failOpen?: boolean
}

interface Chain {
  id: string
  name: string
  isDefault: boolean
  stickySession?: boolean
  models: ChainModel[]
}

interface ChainBuilderProps {
  chains: Chain[]
  selectedChain?: Chain
  onSelectChain: (chain: Chain) => void
  onSave?: (chain: Chain) => void
}

export function ChainBuilder({ chains, selectedChain, onSelectChain, onSave }: ChainBuilderProps) {
  const [stickySession, setStickySession] = useState(selectedChain?.stickySession ?? false)

  const handleSelectChain = (chain: Chain) => {
    setStickySession(chain.stickySession ?? false)
    onSelectChain(chain)
  }

  const handleToggleSticky = () => {
    const next = !stickySession
    setStickySession(next)
    if (selectedChain && onSave) {
      onSave({ ...selectedChain, stickySession: next })
    }
  }

  return (
    <div className="grid grid-cols-3 gap-6">
      {/* Chain List */}
      <Card className="col-span-1">
        <h3 className="section-title mb-4">Routing Chains</h3>
        <div className="space-y-2">
          {chains.map((chain) => (
            <button
              key={chain.id}
              onClick={() => handleSelectChain(chain)}
              className={`w-full text-left px-3 py-2 rounded-md transition-colors ${
                selectedChain?.id === chain.id
                  ? 'bg-accent/10 text-accent'
                  : 'hover:bg-surface text-text-muted hover:text-text-primary'
              }`}
            >
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-body font-medium">{chain.name}</span>
                {chain.isDefault && (
                  <Badge variant="accent">Default</Badge>
                )}
                {chain.stickySession && (
                  <Badge variant="warning">Sticky</Badge>
                )}
              </div>
              <p className="text-micro text-text-muted mt-1">
                {chain.models.length} models
              </p>
            </button>
          ))}
        </div>
      </Card>

      {/* Chain Editor */}
      <Card className="col-span-2">
        {selectedChain ? (
          <>
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <h3 className="section-title">{selectedChain.name}</h3>
                <Badge variant={selectedChain.isDefault ? 'accent' : 'muted'}>
                  {selectedChain.isDefault ? 'Default' : 'Custom'}
                </Badge>
              </div>

              {/* Sticky Session Toggle */}
              <button
                onClick={handleToggleSticky}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-body transition-colors ${
                  stickySession
                    ? 'bg-warning/10 text-warning border border-warning/30'
                    : 'bg-surface text-text-muted border border-border hover:border-warning/30 hover:text-warning'
                }`}
              >
                <span className="text-micro uppercase tracking-wider">Sticky Session</span>
                <div
                  className={`w-8 h-4 rounded-full transition-colors relative ${
                    stickySession ? 'bg-warning' : 'bg-border'
                  }`}
                >
                  <div
                    className={`absolute top-0.5 w-3 h-3 rounded-full bg-white transition-transform ${
                      stickySession ? 'translate-x-4' : 'translate-x-0.5'
                    }`}
                  />
                </div>
              </button>
            </div>

            <div className="space-y-4">
              <p className="text-body text-text-muted">
                Request flow — drag to reorder priority
                {stickySession && (
                  <span className="ml-2 text-warning">
                    • Sticky session enabled — routes by <code className="font-mono text-micro">axis_session_id</code>
                  </span>
                )}
              </p>

              {selectedChain.models.map((model, index) => (
                <div
                  key={`${model.model}-${index}`}
                  className="flex items-center gap-4 p-4 bg-bg rounded-lg border border-border"
                >
                  <div className="flex items-center justify-center w-8 h-8 rounded-full bg-accent/10 text-accent font-mono text-body">
                    {index + 1}
                  </div>

                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: getProviderColor(model.provider) }}
                      />
                      <span className="text-body font-medium text-text-primary font-mono">
                        {model.model}
                      </span>
                    </div>
                    <p className="text-micro text-text-muted mt-1">
                      via {model.provider} • max {model.maxLatency}ms • {model.retries} retries
                    </p>
                  </div>

                  {model.failOpen && (
                    <Badge variant="success">Fail Open</Badge>
                  )}

                  <div className="flex items-center gap-1 text-text-muted cursor-move">
                    <svg className="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                      <path d="M7 2a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 2zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 8zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 14zm6-8a2 2 0 1 0 .001 4.001A2 2 0 0 0 13 6zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 13 12zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 13 18z" />
                    </svg>
                  </div>
                </div>
              ))}
            </div>
          </>
        ) : (
          <div className="flex items-center justify-center h-64 text-text-muted">
            Select a chain to view details
          </div>
        )}
      </Card>
    </div>
  )
}
