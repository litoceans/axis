export function Spinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizes = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-8 w-8' }
  return (
    <div className={`animate-spin ${sizes[size]} border-2 border-accent border-t-transparent rounded-full`} />
  )
}
