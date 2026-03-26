import { useMemo } from 'react'
import {
  summaryStats,
  sparklineData,
  usageData,
  analyticsData,
  recentRequests,
  cacheStats,
  costPrediction,
} from '@/lib/mockData'

export function useMetrics() {
  const stats = summaryStats
  const sparkline = sparklineData
  const usage = usageData
  const analytics = analyticsData
  const requests = recentRequests
  const cache = cacheStats
  const prediction = costPrediction

  const topModelsBySpend = useMemo(() => {
    return analytics
      .slice(-30)
      .reduce((acc, day) => {
        // Simulate model breakdown
        return [
          { id: 'gpt-4o-mini', name: 'GPT-4o Mini', provider: 'openai', cost: 847.23, color: '#10A37F' },
          { id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', provider: 'anthropic', cost: 923.45, color: '#D97706' },
          { id: 'claude-3-5-haiku', name: 'Claude 3.5 Haiku', provider: 'anthropic', cost: 634.12, color: '#D97706' },
          { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', provider: 'google', cost: 423.89, color: '#4285F4' },
          { id: 'gpt-4o', name: 'GPT-4o', provider: 'openai', cost: 312.45, color: '#10A37F' },
        ]
      }, [] as Array<{ id: string; name: string; provider: string; cost: number; color: string }>)
  }, [analytics])

  return {
    stats,
    sparkline,
    usage,
    analytics,
    requests,
    cache,
    prediction,
    topModelsBySpend,
  }
}
