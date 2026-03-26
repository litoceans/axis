import { cn } from '@/lib/utils'

export interface BadgeProps {
  variant?: 'success' | 'warning' | 'danger' | 'muted' | 'accent'
  children: React.ReactNode
  className?: string
}

export function Badge({ variant = 'muted', children, className }: BadgeProps) {
  const variants = {
    success: 'badge-success',
    warning: 'badge-warning',
    danger: 'badge-danger',
    muted: 'badge-muted',
    accent: 'badge-accent',
  }

  return (
    <span className={cn(variants[variant], className)}>
      {children}
    </span>
  )
}
