import { useState, useEffect, useCallback } from 'react'

const STORAGE_KEY = 'redis_command_history'
const MAX_ENTRIES = 100

export interface RedisHistoryEntry {
  id: string
  command: string
  duration_ms: number
  success: boolean
  created_at: string
}

export function useRedisHistory() {
  const [entries, setEntries] = useState<RedisHistoryEntry[]>([])
  const [enabled, setEnabled] = useState(true)

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      try {
        setEntries(JSON.parse(stored))
      } catch {
        localStorage.removeItem(STORAGE_KEY)
      }
    }
  }, [])

  const addEntry = useCallback((entry: Omit<RedisHistoryEntry, 'id' | 'created_at'>) => {
    if (!enabled) return
    setEntries(prev => {
      const newEntry: RedisHistoryEntry = {
        ...entry,
        id: Date.now().toString(),
        created_at: new Date().toISOString(),
      }
      const updated = [newEntry, ...prev].slice(0, MAX_ENTRIES)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
      return updated
    })
  }, [enabled])

  const removeEntry = useCallback((id: string) => {
    setEntries(prev => {
      const updated = prev.filter(e => e.id !== id)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
      return updated
    })
  }, [])

  const clearHistory = useCallback(() => {
    setEntries([])
    localStorage.removeItem(STORAGE_KEY)
  }, [])

  const toggleEnabled = useCallback((value: boolean) => {
    setEnabled(value)
  }, [])

  return {
    entries,
    addEntry,
    removeEntry,
    clearHistory,
    enabled,
    toggleEnabled,
  }
}
