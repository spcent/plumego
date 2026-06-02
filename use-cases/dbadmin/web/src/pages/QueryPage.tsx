import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { api, ApiError, type QueryResult } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import WorkbenchHeader from '../components/WorkbenchHeader'
import ErrorState from '../components/ErrorState'
import { useSqlHistory } from '../hooks/useSqlHistory'
import { XIcon } from '../components/Icons'

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

function makeQueryId(): string {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID()
  return `q_${Date.now().toString(36)}_${Math.random().toString(36).slice(2)}`
}

export default function QueryPage() {
  const { connId } = useParams<{ connId: string }>()
  const [databases, setDatabases] = useState<string[]>([])
  const [selectedDb, setSelectedDb] = useState('')
  const [sql, setSql] = useState('SELECT 1')
  const [result, setResult] = useState<QueryResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [running, setRunning] = useState(false)
  const [activeQueryId, setActiveQueryId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'result' | 'history'>('result')
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)
  const editorRef = useRef<unknown>(null)
  const runRef = useRef<() => void>(() => {})
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const connectionName = currentConn?.name ?? ''
  const isReadonly = currentConn?.readonly ?? false
  const sqlHistory = useSqlHistory()

  useEffect(() => {
    if (!connId) return
    api.resources(connId)
      .then(nodes => setDatabases(nodes.filter(n => n.type === 'sql_database').map(n => n.name)))
      .catch(e => showToast(e.message))
  }, [connId, showToast])

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
    const queryId = makeQueryId()
    setActiveQueryId(queryId)
    const startTime = Date.now()
    try {
      const r = await api.executeQuery(connId, selectedDb, sql, { confirmDangerous, queryId })
      setResult(r)
      setActiveTab('result')
      // Record to localStorage history
      sqlHistory.add({
        sql: sql.trim(),
        database: selectedDb,
        connectionName,
        executionTimeMs: Date.now() - startTime,
        success: true,
      })
    } catch (e) {
      if (e instanceof ApiError && e.details?.confirm_required) {
        setConfirmState({ reason: String(e.details.reason ?? 'dangerous operation') })
      } else {
        const msg = e instanceof Error ? e.message : 'Query failed'
        setError(msg)
        // Record failed query to history
        sqlHistory.add({
          sql: sql.trim(),
          database: selectedDb,
          connectionName,
          executionTimeMs: Date.now() - startTime,
          success: false,
        })
      }
    } finally {
      setRunning(false)
      setActiveQueryId(null)
    }
  }, [connId, selectedDb, sql, sqlHistory, connectionName])

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

  function handleDeleteHistory(entryId: string) {
    sqlHistory.remove(entryId)
  }

  function handleClearHistory() {
    sqlHistory.clear()
  }

  async function handleCancelQuery() {
    if (!activeQueryId) return
    try {
      await api.cancelQuery(activeQueryId)
      showToast(t('query.cancelled'), 'success')
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Cancel failed', 'error')
    }
  }

  const tabLabel = result
    ? `(${result.type === 'result_set' ? result.rows.length : result.rowsAffected} ${
        result.type === 'result_set' ? 'rows' : 'affected'
      })`
    : ''

  return (
    <div className="flex flex-col h-full">
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={[t('query.title')]}
        datasourceType={currentConn?.driver ?? 'mysql'}
        readonly={isReadonly}
      />
      <div className="flex flex-col flex-1 overflow-hidden p-4 gap-3">
      {/* Toolbar */}
      <div className="toolbar rounded-md">
        <select
          value={selectedDb}
          onChange={e => setSelectedDb(e.target.value)}
          className="select max-w-60"
        >
          <option value="">{t('query.db_placeholder')}</option>
          {databases.map(db => <option key={db} value={db}>{db}</option>)}
        </select>
        <button
          onClick={() => handleRun(false)}
          disabled={running}
          className="btn btn-primary disabled:opacity-50"
        >
          {running ? t('query.running') : t('query.run')}
        </button>
        {running && activeQueryId && (
          <button
            onClick={handleCancelQuery}
            className="btn btn-danger"
          >
            {t('query.cancel')}
          </button>
        )}
        <button
          onClick={handleCopy}
          className="btn btn-ghost"
        >{t('query.copy')}</button>
        <button
          onClick={() => setSql('')}
          className="btn btn-ghost"
        >{t('query.clear')}</button>
      </div>

      {/* Editor */}
      <div className="overflow-hidden rounded-md shrink-0" style={{ height: 240, border: '1px solid var(--border-subtle)' }}>
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
        <ErrorState
          title="Query Error"
          message={error}
          onRetry={() => handleRun(false)}
        />
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
            {t('history.tab')} ({sqlHistory.entries.length})
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
            <div
              className="flex items-center gap-2 rounded-md border p-4 text-sm"
              style={{ background: 'var(--success-soft)', borderColor: 'var(--success-border)', color: 'var(--success)' }}
            >
              <span className="status-dot" style={{ background: 'var(--success)' }} />
              {result.rowsAffected} {t('query.rows_affected')} · {result.executionTimeMs}ms
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
            {sqlHistory.entries.length > 0 && (
              <div className="flex justify-end mb-2">
                <button
                  onClick={handleClearHistory}
                  className="text-xs px-2 py-1"
                  style={{ color: 'var(--text-subtle)' }}
                >{t('history.clear_all')}</button>
              </div>
            )}
            <div className="space-y-1">
              {sqlHistory.entries.map(h => (
                <div
                  key={h.id}
                  className="group relative rounded-md border p-3 transition-colors hover:bg-[var(--bg-hover)]"
                  style={{ borderColor: 'var(--border-subtle)' }}
                >
                  <div onClick={() => setSql(h.sql)} className="cursor-pointer">
                    <div className="flex items-center justify-between mb-0.5">
                      <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
                        {new Date(h.executedAt).toLocaleString()}
                        {h.database ? ` · ${h.database}` : ''}
                      </span>
                      <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>{h.executionTimeMs}ms</span>
                    </div>
                    <pre className="text-xs font-mono truncate" style={{ color: 'var(--text-default)' }}>{h.sql}</pre>
                    {!h.success && (
                      <div className="mt-1 flex items-center gap-1.5 text-xs" style={{ color: 'var(--danger)' }}>
                        <span className="status-dot" style={{ background: 'var(--danger)' }} />
                        Failed
                      </div>
                    )}
                  </div>
                  <button
                    onClick={() => handleDeleteHistory(h.id)}
                    className="icon-btn absolute right-2 top-2 opacity-0 group-hover:opacity-100"
                    title={t('history.delete')}
                    aria-label={t('history.delete')}
                  >
                    <XIcon className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
              {sqlHistory.entries.length === 0 && (
                <div className="text-center py-6 text-sm" style={{ color: 'var(--text-subtle)' }}>
                  {t('history.empty')}
                </div>
              )}
            </div>
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
