import { useState } from 'react'
import { Modal, ModalFooter } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { CreateKeyRequest } from '@/api/axis'

interface CreateKeyModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreateKeyRequest) => void
}

export function CreateKeyModal({ open, onOpenChange, onSubmit }: CreateKeyModalProps) {
  const [name, setName] = useState('')
  const [rpmLimit, setRpmLimit] = useState('')
  const [tpmLimit, setTpmLimit] = useState('')
  const [monthlyBudget, setMonthlyBudget] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await onSubmit({
        name,
        scopes: ['chat', 'embeddings'],
        rpmLimit: rpmLimit ? parseInt(rpmLimit) : undefined,
        tpmLimit: tpmLimit ? parseInt(tpmLimit) : undefined,
        monthlyBudgetUsd: monthlyBudget ? parseFloat(monthlyBudget) : undefined,
      })
      setName('')
      setRpmLimit('')
      setTpmLimit('')
      setMonthlyBudget('')
      onOpenChange(false)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Create API Key"
      description="Create a new API key for your application"
    >
      <form onSubmit={handleSubmit}>
        <div className="space-y-4">
          <Input
            label="Key Name"
            placeholder="Production API Key"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="RPM Limit"
              type="number"
              placeholder="1000"
              value={rpmLimit}
              onChange={(e) => setRpmLimit(e.target.value)}
              helperText="Requests per minute"
            />
            <Input
              label="TPM Limit"
              type="number"
              placeholder="10000000"
              value={tpmLimit}
              onChange={(e) => setTpmLimit(e.target.value)}
              helperText="Tokens per minute"
            />
          </div>

          <Input
            label="Monthly Budget"
            type="number"
            step="0.01"
            placeholder="100.00"
            value={monthlyBudget}
            onChange={(e) => setMonthlyBudget(e.target.value)}
            helperText="Budget in USD"
          />
        </div>

        <ModalFooter>
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Create Key
          </Button>
        </ModalFooter>
      </form>
    </Modal>
  )
}
