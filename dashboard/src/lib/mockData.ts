import { format, subDays, subHours } from 'date-fns'

// Provider Health
export const providerHealth = [
  { provider: 'openai', name: 'OpenAI', status: 'healthy', uptime: 99.8, p99Latency: 890, errorRate: 0.2 },
  { provider: 'anthropic', name: 'Anthropic', status: 'healthy', uptime: 99.9, p99Latency: 1200, errorRate: 0.1 },
  { provider: 'google', name: 'Google', status: 'degraded', uptime: 98.1, p99Latency: 3400, errorRate: 1.9 },
  { provider: 'ollama', name: 'Ollama', status: 'healthy', uptime: 100, p99Latency: 240, errorRate: 0 },
]

// Models
export const models = [
  { id: 'gpt-4o-mini', name: 'GPT-4o Mini', provider: 'openai', inputPrice: 0.15, outputPrice: 0.60, latencyP50: 420, requests: 234000, color: '#10A37F' },
  { id: 'gpt-4o', name: 'GPT-4o', provider: 'openai', inputPrice: 2.50, outputPrice: 10.00, latencyP50: 890, requests: 89000, color: '#10A37F' },
  { id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', provider: 'anthropic', inputPrice: 3.00, outputPrice: 15.00, latencyP50: 1100, requests: 156000, color: '#D97706' },
  { id: 'claude-3-5-haiku', name: 'Claude 3.5 Haiku', provider: 'anthropic', inputPrice: 0.80, outputPrice: 4.00, latencyP50: 520, requests: 198000, color: '#D97706' },
  { id: 'gemini-1.5-flash', name: 'Gemini 1.5 Flash', provider: 'google', inputPrice: 0.075, outputPrice: 0.30, latencyP50: 680, requests: 124000, color: '#4285F4' },
  { id: 'gemini-1.5-pro', name: 'Gemini 1.5 Pro', provider: 'google', inputPrice: 1.25, outputPrice: 5.00, latencyP50: 1400, requests: 45000, color: '#4285F4' },
  { id: 'llama3.3-70b', name: 'Llama 3.3 70B', provider: 'ollama', inputPrice: 0.00, outputPrice: 0.00, latencyP50: 180, requests: 312000, color: '#7C3AED' },
  { id: 'qwen2.5-72b', name: 'Qwen 2.5 72B', provider: 'ollama', inputPrice: 0.00, outputPrice: 0.00, latencyP50: 210, requests: 278000, color: '#7C3AED' },
]

// API Keys
export const apiKeys = [
  { id: 'key_abc123', name: 'Production API', owner: 'alice@godlabs.io', team: 'Engineering', created: subDays(new Date(), 45), lastUsed: subHours(new Date(), 2), spend: 1247.82, rpm: 847, tpm: 8900000, status: 'active' },
  { id: 'key_def456', name: 'Staging API', owner: 'bob@godlabs.io', team: 'Engineering', created: subDays(new Date(), 30), lastUsed: subHours(new Date(), 18), spend: 89.34, rpm: 120, tpm: 1200000, status: 'active' },
  { id: 'key_ghi789', name: 'Analytics Service', owner: 'carol@godlabs.io', team: 'Data', created: subDays(new Date(), 20), lastUsed: subHours(new Date(), 0.5), spend: 456.21, rpm: 423, tpm: 4200000, status: 'active' },
  { id: 'key_jkl012', name: 'Legacy v2 Key', owner: 'dave@godlabs.io', team: 'Platform', created: subDays(new Date(), 90), lastUsed: subDays(new Date(), 15), spend: 2341.56, rpm: 0, tpm: 0, status: 'expired' },
  { id: 'key_mno345', name: 'Mobile App', owner: 'eve@godlabs.io', team: 'Mobile', created: subDays(new Date(), 10), lastUsed: subHours(new Date(), 1), spend: 234.67, rpm: 234, tpm: 2300000, status: 'active' },
  { id: 'key_pqr678', name: 'CI/CD Pipeline', owner: 'frank@godlabs.io', team: 'Platform', created: subDays(new Date(), 5), lastUsed: subHours(new Date(), 0.1), spend: 12.45, rpm: 56, tpm: 560000, status: 'active' },
]

// Recent Requests
export const recentRequests = [
  { id: 'req_001', keyId: 'key_abc123', keyName: 'Production API', model: 'gpt-4o-mini', provider: 'openai', status: 200, latency: 342, cost: 0.00023, tokens: 45, timestamp: subHours(new Date(), 0) },
  { id: 'req_002', keyId: 'key_ghi789', keyName: 'Analytics Service', model: 'claude-3-5-haiku', provider: 'anthropic', status: 200, latency: 1100, cost: 0.00182, tokens: 128, timestamp: subHours(new Date(), 0.02) },
  { id: 'req_003', keyId: 'key_abc123', keyName: 'Production API', model: 'gemini-1.5-flash', provider: 'google', status: 429, latency: 120, cost: 0.00000, tokens: 0, timestamp: subHours(new Date(), 0.03) },
  { id: 'req_004', keyId: 'key_mno345', keyName: 'Mobile App', model: 'llama3.3-70b', provider: 'ollama', status: 200, latency: 180, cost: 0.00000, tokens: 89, timestamp: subHours(new Date(), 0.05) },
  { id: 'req_005', keyId: 'key_pqr678', keyName: 'CI/CD Pipeline', model: 'gpt-4o-mini', provider: 'openai', status: 200, latency: 380, cost: 0.00019, tokens: 38, timestamp: subHours(new Date(), 0.08) },
  { id: 'req_006', keyId: 'key_def456', keyName: 'Staging API', model: 'claude-3-5-sonnet', provider: 'anthropic', status: 200, latency: 2100, cost: 0.00452, tokens: 234, timestamp: subHours(new Date(), 0.1) },
  { id: 'req_007', keyId: 'key_abc123', keyName: 'Production API', model: 'gpt-4o', provider: 'openai', status: 200, latency: 1200, cost: 0.00823, tokens: 456, timestamp: subHours(new Date(), 0.15) },
  { id: 'req_008', keyId: 'key_ghi789', keyName: 'Analytics Service', model: 'qwen2.5-72b', provider: 'ollama', status: 200, latency: 220, cost: 0.00000, tokens: 67, timestamp: subHours(new Date(), 0.2) },
  { id: 'req_009', keyId: 'key_mno345', keyName: 'Mobile App', model: 'gemini-1.5-pro', provider: 'google', status: 500, latency: 800, cost: 0.00000, tokens: 0, timestamp: subHours(new Date(), 0.25) },
  { id: 'req_010', keyId: 'key_abc123', keyName: 'Production API', model: 'claude-3-5-haiku', provider: 'anthropic', status: 200, latency: 480, cost: 0.00112, tokens: 98, timestamp: subHours(new Date(), 0.3) },
]

// Usage Data (7 days)
export const usageData = Array.from({ length: 7 }, (_, i) => {
  const date = subDays(new Date(), 6 - i)
  return {
    date: format(date, 'yyyy-MM-dd'),
    requests: Math.floor(Math.random() * 50000) + 80000,
    tokensIn: Math.floor(Math.random() * 5000000) + 8000000,
    tokensOut: Math.floor(Math.random() * 20000000) + 30000000,
    cost: Math.random() * 300 + 400,
  }
})

// Analytics Data (30 days)
export const analyticsData = Array.from({ length: 30 }, (_, i) => {
  const date = subDays(new Date(), 29 - i)
  return {
    date: format(date, 'yyyy-MM-dd'),
    requests: Math.floor(Math.random() * 60000) + 70000,
    cost: Math.random() * 350 + 380,
    latencyP50: Math.random() * 200 + 400,
    latencyP95: Math.random() * 800 + 800,
    latencyP99: Math.random() * 1500 + 1200,
    errorRate: Math.random() * 2,
  }
})

// Cache Stats
export const cacheStats = {
  hitRate: 0.673,
  hitRateTrend: 0.023,
  totalHits: 1284592,
  totalMisses: 624891,
  savings: 3847.23,
  savingsTrend: 0.156,
  topCachedPrompts: [
    { hash: 'abc123...', count: 2341, lastUsed: subHours(new Date(), 2) },
    { hash: 'def456...', count: 1892, lastUsed: subHours(new Date(), 5) },
    { hash: 'ghi789...', count: 1456, lastUsed: subHours(new Date(), 8) },
    { hash: 'jkl012...', count: 1234, lastUsed: subHours(new Date(), 12) },
    { hash: 'mno345...', count: 987, lastUsed: subHours(new Date(), 18) },
  ],
}

// Routing Chains
export const routingChains = [
  {
    id: 'chain_balanced',
    name: 'Reliable Balanced',
    isDefault: true,
    models: [
      { model: 'gpt-4o-mini', provider: 'openai', maxLatency: 3000, retries: 2, weight: 1 },
      { model: 'claude-3-5-haiku', provider: 'anthropic', maxLatency: 4000, retries: 2, weight: 1 },
      { model: 'gemini-1.5-flash', provider: 'google', maxLatency: 5000, retries: 2, weight: 1, failOpen: true },
    ],
  },
  {
    id: 'chain_quality',
    name: 'Quality First',
    isDefault: false,
    models: [
      { model: 'claude-3-5-sonnet', provider: 'anthropic', maxLatency: 15000, retries: 3, weight: 1 },
      { model: 'gpt-4o', provider: 'openai', maxLatency: 15000, retries: 3, weight: 1 },
    ],
  },
  {
    id: 'chain_fast',
    name: 'Fast Local',
    isDefault: false,
    models: [
      { model: 'llama3.3-70b', provider: 'ollama', maxLatency: 5000, retries: 1, weight: 1 },
      { model: 'qwen2.5-72b', provider: 'ollama', maxLatency: 8000, retries: 1, weight: 1 },
    ],
  },
]

// Teams
export const teams = [
  { id: 'team_eng', name: 'Engineering', members: 8, keys: 12 },
  { id: 'team_data', name: 'Data', members: 3, keys: 5 },
  { id: 'team_platform', name: 'Platform', members: 4, keys: 7 },
  { id: 'team_mobile', name: 'Mobile', members: 2, keys: 3 },
]

// Alerts
export const alerts = [
  { id: 'alert_001', type: 'budget_threshold', severity: 'warning', message: 'Production API key has used 80% of monthly budget', keyId: 'key_abc123', createdAt: subHours(new Date(), 4), acknowledged: false },
  { id: 'alert_002', type: 'provider_degraded', severity: 'warning', message: 'Google provider experiencing elevated error rates (1.9%)', createdAt: subHours(new Date(), 2), acknowledged: false },
  { id: 'alert_003', type: 'high_latency', severity: 'info', message: 'Claude 3.5 Sonnet P99 latency above 2s threshold', createdAt: subHours(new Date(), 8), acknowledged: true },
]

// Summary Stats
export const summaryStats = {
  totalSpend: 3847.23,
  spendTrend: 0.12,
  totalRequests: 1234567,
  requestsTrend: 0.08,
  activeKeys: 5,
  activeKeysTrend: 0,
  cacheHitRate: 0.673,
  cacheHitTrend: 0.023,
  totalCostSavings: 3847.23,
}

// Cost Prediction
export const costPrediction = {
  currentMonthSpend: 3847.23,
  predictedMonthSpend: 4823.50,
  trend: 'up',
  basedOnDays: 25,
}

// Sparkline Data
export const sparklineData = [420, 380, 510, 470, 620, 580, 720, 680, 750, 710, 820, 890]
