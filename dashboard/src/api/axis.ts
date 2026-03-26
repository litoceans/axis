const API_BASE = import.meta.env.VITE_AXIS_URL || 'http://localhost:8080'

export interface ApiKey {
  id: string
  name: string
  keyPrefix: string
  orgId: string
  teamId?: string
  memberId?: string
  scopes: string[]
  models?: string[]
  rpmLimit?: number
  tpmLimit?: number
  monthlyBudgetUsd?: number
  environments: string[]
  expiresAt?: string
  lastUsedAt?: string
  revokedAt?: string
  createdAt: string
}

export interface CreateKeyRequest {
  name: string
  orgId?: string
  teamId?: string
  scopes: string[]
  models?: string[]
  rpmLimit?: number
  tpmLimit?: number
  monthlyBudgetUsd?: number
  expiresAt?: string
  environments?: string[]
}

export interface UsageData {
  date: string
  requests: number
  tokensIn: number
  tokensOut: number
  costUsd: number
}

export interface UsageResponse {
  usage: UsageData[]
  totals: {
    requests: number
    tokensIn: number
    tokensOut: number
    costUsd: number
  }
}

export interface CostBreakdown {
  model: string
  provider: string
  requests: number
  tokensIn: number
  tokensOut: number
  costUsd: number
}

export interface Model {
  id: string
  name: string
  provider: string
  inputPrice: number
  outputPrice: number
  latencyP50?: number
  requests?: number
  maxContextTokens?: number
}

export interface BudgetStatus {
  limit: number
  spent: number
  remaining: number
}

export interface BudgetAlert {
  id: string
  keyId: string
  keyName?: string
  orgId: string
  thresholdPercent: number
  spentUsd: number
  limitUsd: number
  triggeredAt: string
  acknowledged: boolean
}

export interface CostReconciliation {
  requestId: string
  model: string
  provider: string
  reportedCost: number
  calculatedCost: number
  difference: number
  differencePercent: number
  status: 'match' | 'discrepancy'
}

export interface ProviderHealth {
  provider: string
  name: string
  status: 'healthy' | 'degraded' | 'down'
  uptime: number
  p99Latency: number
  errorRate: number
}

export interface RoutingChain {
  id: string
  name: string
  isDefault: boolean
  stickySession?: boolean
  models: ChainModel[]
}

export interface ChainModel {
  model: string
  provider: string
  maxLatency: number
  retries: number
  weight: number
  failOpen?: boolean
}

export interface CacheStats {
  hitRate: number
  totalHits: number
  totalMisses: number
  savings: number
  topCachedPrompts: Array<{ hash: string; count: number; lastUsed: string }>
}

export interface RequestLog {
  id: string
  keyId: string
  keyName: string
  model: string
  provider: string
  status: number
  latency: number
  cost: number
  tokens: number
  timestamp: string
}

class AxisApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public type?: string
  ) {
    super(message)
    this.name = 'AxisApiError'
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('axis_token')

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new AxisApiError(
      errorData.error?.message || `Request failed with status ${response.status}`,
      response.status,
      errorData.error?.type
    )
  }

  if (response.status === 204) {
    return {} as T
  }

  return response.json()
}

// Auth / Keys
export const keysApi = {
  list: () =>
    request<ApiKey[]>('/v1/keys'),

  create: (data: CreateKeyRequest) =>
    request<ApiKey>('/v1/keys', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  delete: (keyId: string) =>
    request<void>(`/v1/keys/${keyId}`, {
      method: 'DELETE',
    }),

  rotate: (keyId: string) =>
    request<{ newKey: string }>(`/v1/keys/${keyId}/rotate`, {
      method: 'POST',
    }),
}

// Usage
export const usageApi = {
  getUsage: (params: {
    keyId?: string
    from?: string
    to?: string
    granularity?: 'hour' | 'day'
  }) => {
    const searchParams = new URLSearchParams()
    if (params.keyId) searchParams.set('key_id', params.keyId)
    if (params.from) searchParams.set('from', params.from)
    if (params.to) searchParams.set('to', params.to)
    if (params.granularity) searchParams.set('granularity', params.granularity)
    return request<UsageResponse>(`/v1/usage?${searchParams}`)
  },

  getCosts: (params: {
    orgId?: string
    by?: 'model' | 'key' | 'day' | 'provider'
  }) => {
    const searchParams = new URLSearchParams()
    if (params.orgId) searchParams.set('org_id', params.orgId)
    if (params.by) searchParams.set('by', params.by)
    return request<CostBreakdown[]>(`/v1/costs?${searchParams}`)
  },
}

// Models
export const modelsApi = {
  list: () =>
    request<Model[]>('/v1/models'),
}

// Health
export const healthApi = {
  getProviderHealth: () =>
    request<ProviderHealth[]>('/v1/health'),
}

// Routing
export const routingApi = {
  listChains: () =>
    request<RoutingChain[]>('/v1/routing/chains'),

  createChain: (data: Omit<RoutingChain, 'id'>) =>
    request<RoutingChain>('/v1/routing/chains', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}

// Budget
export const budgetApi = {
  get: (keyId: string) =>
    request<BudgetStatus>(`/v1/keys/${keyId}/budget`),
}

// Budget Alerts
export const alertsApi = {
  list: (params?: { keyId?: string; acknowledged?: boolean }) => request<BudgetAlert[]>(`/v1/alerts?${new URLSearchParams(params as any)}`),
  acknowledge: (alertId: string) => request<void>(`/v1/alerts/${alertId}/acknowledge`, { method: 'POST' }),
}

// Cost Reconciliation
export const reconciliationApi = {
  list: (params?: { showDiscrepanciesOnly?: boolean }) => request<CostReconciliation[]>(`/v1/costs/reconciliation?${new URLSearchParams(params as any)}`),
}

// Cache
export const cacheApi = {
  getStats: () =>
    request<CacheStats>('/v1/cache/stats'),

  clear: (keyId?: string) => {
    const searchParams = keyId ? `?key_id=${keyId}` : ''
    return request<void>(`/v1/cache${searchParams}`, {
      method: 'DELETE',
    })
  },
}

// Streaming helper for future use
export async function* streamChatCompletions(
  endpoint: string,
  data: Record<string, unknown>
): AsyncGenerator<string, void, unknown> {
  const token = localStorage.getItem('axis_token')
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    },
    body: JSON.stringify(data),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new AxisApiError(
      errorData.error?.message || `Request failed with status ${response.status}`,
      response.status
    )
  }

  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')

  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') return
          yield data
        }
      }
    }
  } finally {
    reader.releaseLock()
  }
}

export { AxisApiError }
