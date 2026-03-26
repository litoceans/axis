import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Card } from '@/components/ui/Card'

interface SpendChartProps {
  data: Array<{
    date: string
    cost: number
  }>
}

export function AnalyticsSpendChart({ data }: SpendChartProps) {
  return (
    <Card>
      <h3 className="section-title mb-4">Daily Spend</h3>
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
              tickFormatter={(value) => `$${value}`}
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
              formatter={(value: number) => [`$${value.toFixed(2)}`, 'Cost']}
            />
            <Line
              type="monotone"
              dataKey="cost"
              stroke="#6366F1"
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4, fill: '#6366F1' }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </Card>
  )
}
