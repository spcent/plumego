import { useCallback, useState } from 'react'

export interface SqlHistoryEntry {
  id: string
  sql: string
  database: string
  connectionName: string
  executedAt: string       // ISO 8601
  success: boolean
  executionTimeMs: number
}

const HISTORY_KEY = 'dbadmin_sql_history'
const SETTINGS_KEY = 'dbadmin_settings'
const MAX_ENTRIES = 100

interface Settings {
  sqlHistoryEnabled: boolean
}

function loadSettings(): Settings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (raw) return { sqlHistoryEnabled: true, ...JSON.parse(raw) }
  } catch {}
  return { sqlHistoryEnabled: true }
}

function saveSettings(s: Settings) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(s))
}

function loadEntries(): SqlHistoryEntry[] {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    if (raw) return JSON.parse(raw) as SqlHistoryEntry[]
  } catch {}
  return []
}

function saveEntries(entries: SqlHistoryEntry[]) {
  localStorage.setItem(HISTORY_KEY, JSON.stringify(entries))
}

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
}

export function useSqlHistory() {
  const [entries, setEntries] = useState<SqlHistoryEntry[]>(() => loadEntries())
  const [enabled, setEnabledState] = useState<boolean>(() => loadSettings().sqlHistoryEnabled)

  const add = useCallback((entry: Omit<SqlHistoryEntry, 'id' | 'executedAt'>) => {
    if (!loadSettings().sqlHistoryEnabled) return
    const newEntry: SqlHistoryEntry = {
      ...entry,
      id: generateId(),
      executedAt: new Date().toISOString(),
    }
    setEntries(prev => {
      const next = [newEntry, ...prev].slice(0, MAX_ENTRIES)
      saveEntries(next)
      return next
    })
  }, [])

  const remove = useCallback((id: string) => {
    setEntries(prev => {
      const next = prev.filter(e => e.id !== id)
      saveEntries(next)
      return next
    })
  }, [])

  const clear = useCallback(() => {
    setEntries([])
    saveEntries([])
  }, [])

  const setEnabled = useCallback((val: boolean) => {
    setEnabledState(val)
    const s = loadSettings()
    saveSettings({ ...s, sqlHistoryEnabled: val })
  }, [])

  return { entries, add, remove, clear, enabled, setEnabled }
}

// Standalone settings helpers for use outside QueryPage (e.g. SettingsPage).
export const sqlHistorySettings = {
  get: (): Settings => loadSettings(),
  setEnabled: (v: boolean) => saveSettings({ ...loadSettings(), sqlHistoryEnabled: v }),
}
