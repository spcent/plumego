import { useState, useEffect, useCallback } from 'react'
import { api, type RedisHistoryEntry } from '../api'

export function useRedisHistory(connId?: string) {
  const [entries, setEntries] = useState<RedisHistoryEntry[]>([])
  const [loading, setLoading] = useState(false)

  const reload = useCallback(async () => {
    if (!connId) {
      setEntries([])
      return
    }
    setLoading(true)
    try {
      setEntries(await api.redisListHistory(connId))
    } finally {
      setLoading(false)
    }
  }, [connId])

  useEffect(() => {
    queueMicrotask(() => {
      void reload()
    })
  }, [reload])

  const removeEntry = useCallback(async (id: string) => {
    if (!connId) return
    await api.redisDeleteHistoryEntry(connId, id)
    setEntries(prev => prev.filter(e => e.id !== id))
  }, [connId])

  const clearHistory = useCallback(async () => {
    if (!connId) return
    await api.redisClearHistory(connId)
    setEntries([])
  }, [connId])

  return {
    entries,
    loading,
    reload,
    removeEntry,
    clearHistory,
  }
}
