import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { api, ApiError, type QueryResult } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'
import { useSqlHistory } from '../hooks/useSqlHistory'

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
  const editorRef = useRef<unknown>(null)
  const runRef = useRef<() => void>(() => {})
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const isReadonly = currentConn?.readonly ?? false
  const { entries: historyEntries, add: addHistory, remove: removeHistory, clear: clearHistory, enabled: historyEnabled } = useSqlHistory()

  useEffect(() => {
    if (!connId) return
    api.databases(connId).then(setDatabases).catch(e => showToast(e.message))
  }, [connId])

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
    const startMs = Date.now()
    try {
      const r = await api.executeQuery(connId, selectedDb, sql, { confirmDangerous })
      setResult(r)
      setActiveTab('result')
      addHistory({
        sql,
        database: selectedDb,
        connectionName: currentConn?.name ?? connId ?? '',
        success: true,
        executionTimeMs: r.executionTimeMs,
      })
    } catch (e) {
      if (e instanceof ApiError && e.details?.confirm_required) {
        setConfirmState({ reason: String(e.details.reason ?? 'dangerous operation') })
      } else {
        const msg = e instanceof Error ? e.message : 'Query failed'
        setError(msg)
        addHistory({
          sql,
          database: selectedDb,
          connectionName: currentConn?.name ?? connId ?? '',
          success: false,
          executionTimeMs: Date.now() - startMs,
        })
      }
    } finally {
      setRunning(false)
    }
  }, [connId, selectedDb, sql, currentConn, addHistory])

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

  const tabLabel = result
    ? `(${result.type === 'result_set' ? result.rows.length : result.rowsAffected} ${
        result.type === 'result_set' ? 'rows' : 'affected'
      })`
    : ''

  return (
    <div className="flex flex-col h-full p-4 gap-3">
      {/* Toolbar */}
      <div className="flex items-center gap-2 flex-wrap">
        <h1 className="text-lg font-bold text-gray-800 dark:text-gray-100 mr-1 flex items-center gap-2">
          {t('query.title')}
          {isReadonly && (
            <span className="text-xs bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 px-1.5 py-0.5 rounded font-mono font-normal">
              {t('readonly.badge')}
            </span>
          )}
        </h1>
        <select
          value={selectedDb}
          onChange={e => setSelectedDb(e.target.value)}
          className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="">{t('query.db_placeholder')}</option>
          {databases.map(db => <option key={db} value={db}>{db}</option>)}
        </select>
        <button
          onClick={() => handleRun(false)}
          disabled={running}
          className="bg-blue-600 disabled:opacity-50 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700"
        >
          {running ? t('query.running') : t('query.run')}
        </button>
        <button
          onClick={handleCopy}
          className="bg-gray-100 dark:bg-gray-700 dark:text-gray-300 text-sm px-3 py-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
        >{t('query.copy')}</button>
        <button
          onClick={() => setSql('')}
          className="bg-gray-100 dark:bg-gray-700 dark:text-gray-300 text-sm px-3 py-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
        >{t('query.clear')}</button>
      </div>

      {/* Editor */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden shrink-0" style={{ height: 240 }}>
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
        <div className="px-3 py-2 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded text-sm text-red-700 dark:text-red-400 shrink-0">
          {error}
        </div>
      )}

      {/* Result / History tabs */}
      <div className="flex-1 overflow-auto min-h-0">
        <div className="flex gap-3 border-b border-gray-200 dark:border-gray-700 mb-3">
          <button
            onClick={() => setActiveTab('result')}
            className={`px-3 py-1.5 text-sm font-medium border-b-2 -mb-px ${
              activeTab === 'result'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400'
            }`}
          >
            {t('query.tab.result')} {tabLabel}
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`px-3 py-1.5 text-sm font-medium border-b-2 -mb-px ${
              activeTab === 'history'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400'
            }`}
          >
            {t('history.tab')} ({historyEnabled ? historyEntries.length : '—'})
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
            <div className="p-4 bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-700 rounded text-sm text-green-700 dark:text-green-400">
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
            {!historyEnabled ? (
              <div className="text-center text-gray-400 py-6 text-sm">{t('history.disabled')}</div>
            ) : (
              <>
                {historyEntries.length > 0 && (
                  <div className="flex justify-end mb-2">
                    <button
                      onClick={clearHistory}
                      className="text-xs text-gray-400 hover:text-red-500 px-2 py-1"
                    >{t('history.clear_all')}</button>
                  </div>
                )}
                <div className="space-y-1">
                  {historyEntries.map(h => (
                    <div
                      key={h.id}
                      className="group relative p-3 border border-gray-100 dark:border-gray-700 rounded hover:bg-gray-50 dark:hover:bg-gray-800"
                    >
                      <div
                        onClick={() => setSql(h.sql)}
                        className="cursor-pointer"
                      >
                        <div className="flex items-center justify-between mb-0.5">
                          <span className="text-xs text-gray-400">{new Date(h.executedAt).toLocaleString()}</span>
                          <span className="text-xs text-gray-400">{h.executionTimeMs}ms</span>
                        </div>
                        <pre className="text-xs font-mono truncate text-gray-700 dark:text-gray-300">{h.sql}</pre>
                        {!h.success && <div className="text-xs text-red-500 mt-0.5">✗ failed</div>}
                      </div>
                      <button
                        onClick={() => removeHistory(h.id)}
                        className="absolute top-2 right-2 text-xs text-gray-300 dark:text-gray-600 hover:text-red-500 opacity-0 group-hover:opacity-100 px-1"
                        title={t('history.delete')}
                      >×</button>
                    </div>
                  ))}
                  {historyEntries.length === 0 && (
                    <div className="text-center text-gray-400 py-6 text-sm">{t('history.empty')}</div>
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
        <div className="mb-2 px-3 py-1.5 bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-700 text-amber-700 dark:text-amber-400 text-xs rounded">
          {t('query.truncated')}
        </div>
      )}
      <div className="text-xs text-gray-400 mb-2">{rows.length} rows · {duration}ms</div>
      <div className="overflow-auto border border-gray-200 dark:border-gray-700 rounded-lg">
        <table className="w-full text-xs">
          <thead className="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              {columns.map(c => (
                <th key={c} className="text-left px-3 py-2 font-medium text-gray-600 dark:text-gray-400 whitespace-nowrap font-mono">
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {rows.map((row, i) => (
              <tr key={i} className="hover:bg-blue-50 dark:hover:bg-blue-900/20">
                {columns.map(c => (
                  <td key={c} className="px-3 py-1.5 max-w-xs truncate text-gray-700 dark:text-gray-300">
                    {row[c] === null || row[c] === undefined
                      ? <span className="text-gray-300 dark:text-gray-600 italic">NULL</span>
                      : String(row[c])}
                  </td>
                ))}
              </tr>
            ))}
            {rows.length === 0 && (
              <tr>
                <td colSpan={columns.length || 1} className="px-3 py-6 text-center text-gray-400">
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
