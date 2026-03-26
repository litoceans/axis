import { useState } from 'react'

interface ToggleProps {
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  label?: string
}

export function Toggle({ checked, onCheckedChange, label }: ToggleProps) {
  return (
    <label className="flex items-center gap-3 cursor-pointer">
      <div
        className={`relative w-10 h-5 rounded-full transition-colors ${
          checked ? 'bg-primary' : 'bg-surface-border'
        }`}
        onClick={() => onCheckedChange(!checked)}
      >
        <div
          className={`absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform ${
            checked ? 'translate-x-5' : 'translate-x-0'
          }`}
        />
      </div>
      {label && <span className="text-body text-text-secondary">{label}</span>}
    </label>
  )
}
