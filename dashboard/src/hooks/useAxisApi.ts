import { useState, useEffect, useCallback } from 'react'
import { keysApi, usageApi, modelsApi, healthApi, routingApi, cacheApi, budgetApi, alertsApi, reconciliationApi } from '@/api/axis'
import type {CreateKeyRequest, BudgetAlert, CostReconciliation} from '@/api/axis'

interface UseApiOptions<T> {
  initialData?: T
  enabled?: boolean
}

export function useApi<T>(
  fetcher: () => Promise<T>,
  options: UseApiOptions<T> = {}
) {
  const { initialData, enabled = true } = options
  const [data, setData] = useState<T | undefined>(initialData)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const refetch = useCallback(async () => {
    if (!enabled) return
    setLoading(true)
    setError(null)
    try {
      const result = await fetcher()
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'))
    } finally {
      setLoading(false)
    }
  }, [fetcher, enabled])

  useEffect(() => {
    refetch()
  }, [refetch])

  return { data, loading, error, refetch }
}

// Hooks for specific endpoints
export function useKeys() {
  return useApi(() => keysApi.list())
}

export function useCreateKey() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const createKey = useCallback(async (data: CreateKeyRequest) => {
    setLoading(true)
    setError(null)
    try {
      const result = await keysApi.create(data)
      return result
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'))
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  return { createKey, loading, error }
}

export function useUsage(params?: { keyId?: string; from?: string; to?: string }) {
  return useApi(() => usageApi.getUsage(params || {}))
}

export function useCosts(params?: { orgId?: string; by?: 'model' | 'key' | 'day' | 'provider' }) {
  return useApi(() => usageApi.getCosts(params || {}))
}

export function useModels() {
  return useApi(() => modelsApi.list())
}

export function useHealth() {
  return useApi(() => healthApi.getProviderHealth())
}

export function useRoutingChains() {
  return useApi(() => routingApi.listChains())
}

export function useCacheStats() {
  return useApi(() => cacheApi.getStats())
}

export function useBudget(keyId: string) {
  return useApi(() => budgetApi.get(keyId), { enabled: !!keyId })
}

export function useBudgetAlerts(params?: { keyId?: string; acknowledged?: boolean }) {
  return useApi(() => alertsApi.list(params))
}

export function useAcknowledgeAlert() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const acknowledge = useCallback(async (alertId: string) => {
    setLoading(true)
    setError(null)
    try {
      await alertsApi.acknowledge(alertId)
      return true
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'))
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  return { acknowledge, loading, error }
}

export function useCostReconciliation(params?: { showDiscrepanciesOnly?: boolean }) {
  return useApi(() => reconciliationApi.list(params))
}

// Re-export types and APIs used by pages
export type { ApiKey, CreateKeyRequest, UsageData, UsageResponse, CostBreakdown, Model, ProviderHealth, RoutingChain, ChainModel, CacheStats, RequestLog, BudgetStatus, BudgetAlert, CostReconciliation } from '@/api/axis'
export { cacheApi, budgetApi, alertsApi, reconciliationApi }
