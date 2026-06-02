import { useState, useCallback } from 'react'
import type { ESHistoryEntry } from '../api'

const STORAGE_KEY = 'elasticsearch_query_history'
const MAX_ENTRIES = 100

function loadEntries(): ESHistoryEntry[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored) as ESHistoryEntry[]
    }
  } catch {
    localStorage.removeItem(STORAGE_KEY)
  }
  return []
}

export function useEsHistory() {
  const [entries, setEntries] = useState<ESHistoryEntry[]>(() => loadEntries())
  const [enabled, setEnabled] = useState(true)

  const addEntry = useCallback((entry: Omit<ESHistoryEntry, 'id' | 'created_at'>) => {
    if (!enabled) return
    setEntries(prev => {
      const newEntry: ESHistoryEntry = {
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
