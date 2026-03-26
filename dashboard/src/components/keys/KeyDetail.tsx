import { formatCurrency, formatNumber, getRelativeTime } from '@/lib/utils'
import { KeyData } from '@/components/keys/KeysTable'

export interface KeyDetailProps {
  selectedKeyData: KeyData
  onClose: () => void
}

export function KeyDetail({ selectedKeyData: key, onClose }: KeyDetailProps) {
  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-surface border border-border rounded-lg w-full max-w-2xl max-h-[80vh] overflow-y-auto">
        <div className="p-6 border-b border-border">
          <div className="flex items-start justify-between">
            <div>
              <h2 className="text-title font-semibold text-text-primary">{key.name}</h2>
              <p className="text-body text-text-muted font-mono mt-1">{key.id}</p>
            </div>
            <button
              onClick={onClose}
              className="btn-ghost p-2 rounded-md hover:bg-bg"
            >
              ×
            </button>
          </div>
        </div>

        <div className="p-6 space-y-6">
          {/* Stats */}
          <div className="grid grid-cols-3 gap-4">
            <div className="card">
              <p className="text-micro text-text-muted mb-1">This Month</p>
              <p className="stat-number text-text-primary">{formatCurrency(key.spend)}</p>
            </div>
            <div className="card">
              <p className="text-micro text-text-muted mb-1">RPM</p>
              <p className="stat-number text-text-primary">{formatNumber(key.rpm)}</p>
            </div>
            <div className="card">
              <p className="text-micro text-text-muted mb-1">TPM</p>
              <p className="stat-number text-text-primary">{formatNumber(key.tpm)}</p>
            </div>
          </div>

          {/* Details */}
          <div className="space-y-4">
            <h3 className="section-title">Details</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-micro text-text-muted">Owner</p>
                <p className="text-body text-text-primary">{key.owner}</p>
              </div>
              <div>
                <p className="text-micro text-text-muted">Team</p>
                <p className="text-body text-text-primary">{key.team}</p>
              </div>
              <div>
                <p className="text-micro text-text-muted">Created</p>
                <p className="text-body text-text-primary">{getRelativeTime(key.created)}</p>
              </div>
              <div>
                <p className="text-micro text-text-muted">Last Used</p>
                <p className="text-body text-text-primary">{getRelativeTime(key.lastUsed)}</p>
              </div>
            </div>
          </div>

          {/* Danger Zone */}
          <div className="border border-danger/30 rounded-lg p-4">
            <h3 className="text-body font-semibold text-danger mb-3">Danger Zone</h3>
            <div className="flex gap-3">
              <button className="btn btn-secondary btn-sm">
                Rotate Key
              </button>
              <button className="btn btn-danger btn-sm">
                Delete Key
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
