import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { api, ApiError, type QueryResult, type HistoryEntry } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'

interface ConfirmState {
  reason: string
}

// Mirror of backend classifySQL for client-side pre-check.
function clientClassify(sql: string): { dangerous: boolean; reason: string } {
  const upper = sql.trimStart().toUpperCase()
  if (/^DROP\b/.test(upper)) return { dangerous: true, reason: 'DROP destroys data irreversibly' }
  if (/^TRUNCATE\b/.test(upper)) return { dangerous: true, reason: 'TRUNCATE removes all rows' }
  if (/^ALTER\b/.test(upper)) return { dangerous: true, reason: 'ALTER TABLE modifies table structure' }
  if (/^DELETE\b/.test(upper) && !/\bWHERE\b/.test(upper))
    return { dangerous: true, reason: 'DELETE without WHERE affects all rows' }
  if (/^UPDATE\b/.test(upper) && !/\bWHERE\b/.test(upper))
    return { dangerous: true, reason: 'UPDATE without WHERE affects all rows' }
  return { dangerous: false, reason: '' }
}

export default function QueryPage() {
  const { connId } = useParams<{ connId: string }>()
  const [databases, setDatabases] = useState<string[]>([])
  const [selectedDb, setSelectedDb] = useState('')
  const [sql, setSql] = useState('SELECT 1')
  const [result, setResult] = useState<QueryResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [running, setRunning] = useState(false)
  const [activeTab, setActiveTab] = useState<'result' | 'history'>('result')
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)
  const [serverHistory, setServerHistory] = useState<HistoryEntry[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const editorRef = useRef<unknown>(null)
  const runRef = useRef<() => void>(() => {})
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const isReadonly = currentConn?.readonly ?? false

  useEffect(() => {
    if (!connId) return
    api.resources(connId)
      .then(nodes => setDatabases(nodes.filter(n => n.type === 'sql_database').map(n => n.name)))
      .catch(e => showToast(e.message))
  }, [connId])

  const loadHistory = useCallback(async () => {
    if (!connId) return
    setHistoryLoading(true)
    try {
      const entries = await api.listHistory(connId)
      setServerHistory(entries ?? [])
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Failed to load history')
    } finally {
      setHistoryLoading(false)
    }
  }, [connId])

  // Load history when switching to history tab.
  useEffect(() => {
    if (activeTab === 'history') {
      loadHistory()
    }
  }, [activeTab, loadHistory])

  const handleRun = useCallback(async (confirmDangerous = false) => {
    if (!connId || !sql.trim()) return
    setError(null)

    if (!confirmDangerous) {
      const { dangerous, reason } = clientClassify(sql)
      if (dangerous) {
        setConfirmState({ reason })
        return
      }
    }

    setRunning(true)
    try {
      const r = await api.executeQuery(connId, selectedDb, sql, { confirmDangerous })
      setResult(r)
      setActiveTab('result')
    } catch (e) {
      if (e instanceof ApiError && e.details?.confirm_required) {
        setConfirmState({ reason: String(e.details.reason ?? 'dangerous operation') })
      } else {
        const msg = e instanceof Error ? e.message : 'Query failed'
        setError(msg)
      }
    } finally {
      setRunning(false)
    }
  }, [connId, selectedDb, sql])

  // Keep runRef in sync so Monaco keybinding never captures a stale closure.
  useEffect(() => {
    runRef.current = () => handleRun(false)
  }, [handleRun])

  function handleEditorMount(editor: unknown) {
    editorRef.current = editor
    ;(editor as { addAction: (a: unknown) => void }).addAction({
      id: 'run-query',
      label: 'Run Query',
      keybindings: [2051], // KeyMod.CtrlCmd | KeyCode.Enter
      run: () => runRef.current(),
    })
  }

  function handleCopy() {
    navigator.clipboard.writeText(sql).then(
      () => showToast('Copied', 'success'),
      () => showToast('Copy failed'),
    )
  }

  async function handleDeleteHistory(entryId: string) {
    if (!connId) return
    try {
      await api.deleteHistory(connId, entryId)
      setServerHistory(prev => prev.filter(e => e.id !== entryId))
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Failed to delete entry')
    }
  }

  async function handleClearHistory() {
    if (!connId) return
    try {
      await api.clearHistory(connId)
      setServerHistory([])
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Failed to clear history')
    }
  }

  const tabLabel = result
    ? `(${result.type === 'result_set' ? result.rows.length : result.rowsAffected} ${
        result.type === 'result_set' ? 'rows' : 'affected'
      })`
    : ''

  return (
    <div className="flex flex-col h-full p-4 gap-3">
      {/* Toolbar */}
      <div className="flex items-center gap-2 flex-wrap">
        <h1 className="text-lg font-bold mr-1 flex items-center gap-2" style={{ color: 'var(--text-strong)' }}>
          {t('query.title')}
          {isReadonly && (
            <span className="text-xs px-1.5 py-0.5 rounded font-mono font-normal"
              style={{ background: 'var(--bg-muted)', color: 'var(--warning)', border: '1px solid var(--border-strong)' }}>
              {t('readonly.badge')}
            </span>
          )}
        </h1>
        <select
          value={selectedDb}
          onChange={e => setSelectedDb(e.target.value)}
          className="border rounded px-2 py-1 text-sm"
          style={{ borderColor: 'var(--border-strong)', background: 'var(--bg-surface)', color: 'var(--text-default)' }}
        >
          <option value="">{t('query.db_placeholder')}</option>
          {databases.map(db => <option key={db} value={db}>{db}</option>)}
        </select>
        <button
          onClick={() => handleRun(false)}
          disabled={running}
          className="disabled:opacity-50 text-white px-4 py-1.5 rounded text-sm"
          style={{ background: 'var(--accent)' }}
        >
          {running ? t('query.running') : t('query.run')}
        </button>
        <button
          onClick={handleCopy}
          className="text-sm px-3 py-1.5 rounded"
          style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)', border: '1px solid var(--border-subtle)' }}
        >{t('query.copy')}</button>
        <button
          onClick={() => setSql('')}
          className="text-sm px-3 py-1.5 rounded"
          style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)', border: '1px solid var(--border-subtle)' }}
        >{t('query.clear')}</button>
      </div>

      {/* Editor */}
      <div className="rounded-lg overflow-hidden shrink-0" style={{ height: 240, border: '1px solid var(--border-subtle)' }}>
        <Editor
          defaultLanguage="sql"
          value={sql}
          onChange={v => setSql(v || '')}
          onMount={handleEditorMount}
          theme="vs-dark"
          options={{
            fontSize: 13,
            minimap: { enabled: false },
            lineNumbers: 'on',
            wordWrap: 'on',
            scrollBeyondLastLine: false,
          }}
        />
      </div>

      {/* Error banner */}
      {error && (
        <div className="px-3 py-2 rounded text-sm shrink-0"
          style={{ background: 'var(--bg-muted)', border: '1px solid var(--danger)', color: 'var(--danger)' }}>
          {error}
        </div>
      )}

      {/* Result / History tabs */}
      <div className="flex-1 overflow-auto min-h-0">
        <div className="flex gap-3 mb-3" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
          <button
            onClick={() => setActiveTab('result')}
            style={{
              padding: '6px 12px',
              fontSize: 13,
              fontWeight: activeTab === 'result' ? 500 : 400,
              borderBottom: activeTab === 'result' ? '2px solid var(--accent)' : '2px solid transparent',
              color: activeTab === 'result' ? 'var(--accent)' : 'var(--text-muted)',
              background: 'transparent',
              marginBottom: -1,
            }}
          >
            {t('query.tab.result')} {tabLabel}
          </button>
          <button
            onClick={() => setActiveTab('history')}
            style={{
              padding: '6px 12px',
              fontSize: 13,
              fontWeight: activeTab === 'history' ? 500 : 400,
              borderBottom: activeTab === 'history' ? '2px solid var(--accent)' : '2px solid transparent',
              color: activeTab === 'history' ? 'var(--accent)' : 'var(--text-muted)',
              background: 'transparent',
              marginBottom: -1,
            }}
          >
            {t('history.tab')} ({serverHistory.length})
          </button>
        </div>

        {activeTab === 'result' && result && (
          result.type === 'result_set' ? (
            <ResultTable
              columns={result.columns}
              rows={result.rows}
              duration={result.executionTimeMs}
              truncated={result.truncated}
            />
          ) : (
            <div className="p-4 rounded text-sm"
              style={{ background: 'var(--bg-muted)', border: '1px solid var(--success)', color: 'var(--success)' }}>
              ✓ {result.rowsAffected} {t('query.rows_affected')} · {result.executionTimeMs}ms
              {result.lastInsertId > 0 && (
                <span className="ml-2">
                  · {t('query.last_insert_id')}: {result.lastInsertId}
                </span>
              )}
            </div>
          )
        )}

        {activeTab === 'history' && (
          <div>
            {historyLoading ? (
              <div className="text-center py-6 text-sm" style={{ color: 'var(--text-subtle)' }}>Loading…</div>
            ) : (
              <>
                {serverHistory.length > 0 && (
                  <div className="flex justify-end mb-2">
                    <button
                      onClick={handleClearHistory}
                      className="text-xs px-2 py-1"
                      style={{ color: 'var(--text-subtle)' }}
                    >{t('history.clear_all')}</button>
                  </div>
                )}
                <div className="space-y-1">
                  {serverHistory.map(h => (
                    <div
                      key={h.id}
                      className="group relative p-3 rounded"
                      style={{ border: '1px solid var(--border-subtle)' }}
                    >
                      <div onClick={() => setSql(h.sql)} className="cursor-pointer">
                        <div className="flex items-center justify-between mb-0.5">
                          <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
                            {new Date(h.created_at).toLocaleString()}
                            {h.database ? ` · ${h.database}` : ''}
                          </span>
                          <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>{h.duration_ms}ms</span>
                        </div>
                        <pre className="text-xs font-mono truncate" style={{ color: 'var(--text-default)' }}>{h.sql}</pre>
                        {h.error && (
                          <div className="text-xs mt-0.5" style={{ color: 'var(--danger)' }}>✗ {h.error}</div>
                        )}
                      </div>
                      <button
                        onClick={() => handleDeleteHistory(h.id)}
                        className="absolute top-2 right-2 text-xs px-1 opacity-0 group-hover:opacity-100"
                        style={{ color: 'var(--text-subtle)' }}
                        title={t('history.delete')}
                      >×</button>
                    </div>
                  ))}
                  {serverHistory.length === 0 && (
                    <div className="text-center py-6 text-sm" style={{ color: 'var(--text-subtle)' }}>
                      {t('history.empty')}
                    </div>
                  )}
                </div>
              </>
            )}
          </div>
        )}
      </div>

      {/* Dangerous SQL confirmation dialog */}
      <ConfirmDialog
        open={!!confirmState}
        title={t('query.confirm_danger')}
        message={`${t('query.danger_reason')}: ${confirmState?.reason ?? ''}`}
        confirmLabel={t('query.confirm_proceed')}
        dangerous
        onConfirm={() => { setConfirmState(null); handleRun(true) }}
        onCancel={() => setConfirmState(null)}
      />
    </div>
  )
}

function ResultTable({
  columns, rows, duration, truncated,
}: {
  columns: string[]
  rows: Record<string, unknown>[]
  duration: number
  truncated: boolean
}) {
  const { t } = useI18n()
  return (
    <div>
      {truncated && (
        <div className="mb-2 px-3 py-1.5 rounded text-xs"
          style={{ background: 'var(--bg-muted)', border: '1px solid var(--warning)', color: 'var(--warning)' }}>
          {t('query.truncated')}
        </div>
      )}
      <div className="text-xs mb-2" style={{ color: 'var(--text-subtle)' }}>{rows.length} rows · {duration}ms</div>
      <div className="overflow-auto rounded-lg" style={{ border: '1px solid var(--border-subtle)' }}>
        <table style={{ width: '100%', minWidth: 'max-content', fontSize: 12, borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ background: 'var(--bg-muted)', borderBottom: '1px solid var(--border-subtle)' }}>
              {columns.map(c => (
                <th key={c} className="text-left px-3 py-2 font-mono whitespace-nowrap"
                  style={{ fontWeight: 500, color: 'var(--text-muted)', fontSize: 11 }}>
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i} className="hover:bg-[var(--bg-hover)]"
                style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                {columns.map(c => (
                  <td key={c} className="px-3 py-1.5 max-w-xs truncate"
                    style={{ color: row[c] === null || row[c] === undefined ? 'var(--text-subtle)' : 'var(--text-default)' }}>
                    {row[c] === null || row[c] === undefined
                      ? <span className="italic font-mono">NULL</span>
                      : String(row[c])}
                  </td>
                ))}
              </tr>
            ))}
            {rows.length === 0 && (
              <tr>
                <td colSpan={columns.length || 1} className="px-3 py-6 text-center"
                  style={{ color: 'var(--text-subtle)' }}>
                  No rows
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
