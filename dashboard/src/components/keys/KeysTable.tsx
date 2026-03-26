import { useState } from 'react'
import { Table } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { formatCurrency, formatNumber, getRelativeTime } from '@/lib/utils'
import { KeyDetail } from './KeyDetail'

export interface KeyData {
  id: string
  name: string
  owner: string
  team: string
  created: Date
  lastUsed: Date
  spend: number
  rpm: number
  tpm: number
  status: string
}

interface KeysTableProps {
  keys: KeyData[]
  onCreateKey?: () => void
}

export function KeysTable({ keys, onCreateKey }: KeysTableProps) {
  const [selectedKey, setSelectedKey] = useState<KeyData | null>(null)

  const columns = [
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (key: KeyData) => (
        <span className="font-medium text-text-primary">{key.name}</span>
      ),
    },
    {
      key: 'owner',
      header: 'Owner',
      sortable: true,
    },
    {
      key: 'team',
      header: 'Team',
      sortable: true,
    },
    {
      key: 'spend',
      header: 'Spend',
      sortable: true,
      render: (key: KeyData) => (
        <span className="font-mono">{formatCurrency(key.spend)}</span>
      ),
    },
    {
      key: 'rpm',
      header: 'RPM',
      sortable: true,
      render: (key: KeyData) => (
        <span className="font-mono">{formatNumber(key.rpm)}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (key: KeyData) => (
        <Badge variant={key.status === 'active' ? 'success' : key.status === 'expired' ? 'warning' : 'danger'}>
          {key.status}
        </Badge>
      ),
    },
    {
      key: 'lastUsed',
      header: 'Last Used',
      sortable: true,
      render: (key: KeyData) => (
        <span className="text-text-muted">{getRelativeTime(key.lastUsed)}</span>
      ),
    },
  ]

  return (
    <>
      <Table
        columns={columns}
        data={keys}
        keyExtractor={(key) => key.id}
        onRowClick={(key) => setSelectedKey(key)}
        pageSize={10}
      />

      {selectedKey && (
        <KeyDetail selectedKeyData={selectedKey} onClose={() => setSelectedKey(null)} />
      )}
    </>
  )
}
