import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { Card } from '@/components/ui/Card'

interface LatencyChartProps {
  data: Array<{
    date: string
    latencyP50: number
    latencyP95: number
    latencyP99: number
  }>
}

export function LatencyChart({ data }: LatencyChartProps) {
  return (
    <Card>
      <h3 className="section-title mb-4">Latency Percentiles</h3>
      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#2A2A2F" />
            <XAxis
              dataKey="date"
              stroke="#71717A"
              fontSize={11}
              tickFormatter={(value) => {
                const date = new Date(value)
                return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
              }}
            />
            <YAxis
              stroke="#71717A"
              fontSize={11}
              tickFormatter={(value) => `${value}ms`}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: '#17171A',
                border: '1px solid #2A2A2F',
                borderRadius: '8px',
                color: '#EDEDEF',
                fontSize: '12px',
              }}
              labelStyle={{ color: '#71717A' }}
              formatter={(value: number) => [`${value}ms`]}
            />
            <Legend
              wrapperStyle={{ fontSize: '12px', color: '#71717A' }}
              formatter={(value) => <span style={{ color: '#71717A' }}>{value}</span>}
            />
            <Line
              type="monotone"
              dataKey="latencyP50"
              stroke="#22C55E"
              strokeWidth={2}
              dot={false}
              name="P50"
            />
            <Line
              type="monotone"
              dataKey="latencyP95"
              stroke="#F59E0B"
              strokeWidth={2}
              dot={false}
              name="P95"
            />
            <Line
              type="monotone"
              dataKey="latencyP99"
              stroke="#EF4444"
              strokeWidth={2}
              dot={false}
              name="P99"
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </Card>
  )
}
