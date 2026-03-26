import { Card } from '@/components/ui/Card'
import { format } from 'date-fns'

interface SpendChartProps {
  data: Array<{
    date: string
    requests: number
    cost: number
  }>
}

export function SpendChart({ data }: SpendChartProps) {
  const maxCost = Math.max(...data.map((d) => d.cost))
  const minCost = Math.min(...data.map((d) => d.cost))
  const range = maxCost - minCost || 1

  const points = data.map((d, i) => ({
    x: (i / (data.length - 1)) * 100,
    y: 100 - ((d.cost - minCost) / range) * 80 - 10,
    cost: d.cost,
    date: d.date,
  }))

  const pathD = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`)
    .join(' ')

  return (
    <Card>
      <h3 className="section-title mb-4">7-Day Spend</h3>
      <div className="h-32 relative">
        <svg
          viewBox="0 0 100 100"
          className="w-full h-full"
          preserveAspectRatio="none"
        >
          {/* Grid lines */}
          <line x1="0" y1="25" x2="100" y2="25" stroke="#2A2A2F" strokeWidth="0.2" />
          <line x1="0" y1="50" x2="100" y2="50" stroke="#2A2A2F" strokeWidth="0.2" />
          <line x1="0" y1="75" x2="100" y2="75" stroke="#2A2A2F" strokeWidth="0.2" />

          {/* Area fill */}
          <path
            d={`${pathD} L 100 100 L 0 100 Z`}
            fill="url(#spendGradient)"
            opacity="0.3"
          />

          {/* Line */}
          <path
            d={pathD}
            fill="none"
            stroke="#6366F1"
            strokeWidth="0.5"
            vectorEffect="non-scaling-stroke"
          />

          {/* Points */}
          {points.map((p, i) => (
            <circle
              key={i}
              cx={p.x}
              cy={p.y}
              r="1"
              fill="#6366F1"
              vectorEffect="non-scaling-stroke"
            />
          ))}

          <defs>
            <linearGradient id="spendGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#6366F1" stopOpacity="0.4" />
              <stop offset="100%" stopColor="#6366F1" stopOpacity="0" />
            </linearGradient>
          </defs>
        </svg>

        {/* Y-axis labels */}
        <div className="absolute left-0 top-0 h-full flex flex-col justify-between text-micro text-text-muted font-mono pointer-events-none">
          <span>${maxCost.toFixed(0)}</span>
          <span>${minCost.toFixed(0)}</span>
        </div>

        {/* X-axis labels */}
        <div className="absolute bottom-0 left-0 right-0 flex justify-between text-micro text-text-muted font-mono pointer-events-none">
          <span>{format(new Date(data[0]?.date || new Date()), 'MMM d')}</span>
          <span>{format(new Date(data[data.length - 1]?.date || new Date()), 'MMM d')}</span>
        </div>
      </div>
    </Card>
  )
}
