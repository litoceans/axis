import { useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { KeysTable } from '@/components/keys/KeysTable'
import { CreateKeyModal } from '@/components/keys/CreateKeyModal'
import { Spinner } from '@/components/ui/Spinner'
import { useKeys, useCreateKey } from '@/hooks/useAxisApi'
import type { CreateKeyRequest } from '@/api/axis'

export function Keys() {
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const { data: keys, loading, error, refetch } = useKeys()
  const { createKey, loading: creating } = useCreateKey()

  const handleCreateKey = async (data: CreateKeyRequest) => {
    await createKey(data)
    refetch()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">API Keys</h1>
          <p className="text-body text-text-muted mt-1">
            Manage your API keys and monitor usage
          </p>
        </div>
        <Button onClick={() => setCreateModalOpen(true)}>
          <Plus className="h-4 w-4" />
          Create Key
        </Button>
      </div>

      {loading && <div className="flex justify-center py-12"><Spinner /></div>}
      {error && <p className="text-body text-danger">Failed to load keys: {error.message}</p>}

      <Card>
        {!loading && !error && (
          <KeysTable
            keys={(keys || []).map((k) => ({
              id: k.id,
              name: k.name,
              owner: k.memberId || k.orgId || 'Unknown',
              team: k.teamId || 'Default',
              created: new Date(k.createdAt),
              lastUsed: k.lastUsedAt ? new Date(k.lastUsedAt) : new Date(0),
              spend: 0,
              rpm: 0,
              tpm: 0,
              status: k.revokedAt ? 'revoked' : 'active',
            }))}
            onCreateKey={() => setCreateModalOpen(true)}
          />
        )}
      </Card>

      <CreateKeyModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSubmit={handleCreateKey}
      />
    </div>
  )
}
