import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { api, ApiError, errorMessage, type QueryResult } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useSqlHistory } from '../hooks/useSqlHistory'
import { XIcon } from '../components/Icons'
import { CodePanel, EmptyStatePanel, ErrorStatePanel, PageBody, PageShell, PageStatusBar, PageToolbar, StatusBanner } from '../components/workbench'

interface ConfirmState {
  reason: string
}

interface QueryRun {
  id: string
  sql: string
  database: string
  result: QueryResult
  executedAt: string
}

const HISTORY_FAVORITES_KEY = 'dbadmin_sql_history_favorites'

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

function loadFavoriteHistoryIds(): Set<string> {
  try {
    const raw = localStorage.getItem(HISTORY_FAVORITES_KEY)
    const parsed = raw ? JSON.parse(raw) as unknown : []
    return new Set(Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === 'string') : [])
  } catch {
    return new Set()
  }
}

function saveFavoriteHistoryIds(ids: Set<string>) {
  localStorage.setItem(HISTORY_FAVORITES_KEY, JSON.stringify([...ids]))
}

export default function QueryPage() {
  const { connId } = useParams<{ connId: string }>()
  const [databases, setDatabases] = useState<string[]>([])
  const [selectedDb, setSelectedDb] = useState('')
  const [sql, setSql] = useState('SELECT 1')
  const [results, setResults] = useState<QueryRun[]>([])
  const [activeResultId, setActiveResultId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [running, setRunning] = useState(false)
  const [activeQueryId, setActiveQueryId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'result' | 'history'>('result')
  const [historySearch, setHistorySearch] = useState('')
  const [historyFavoritesOnly, setHistoryFavoritesOnly] = useState(false)
  const [favoriteHistoryIds, setFavoriteHistoryIds] = useState<Set<string>>(loadFavoriteHistoryIds)
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)
  const editorRef = useRef<unknown>(null)
  const runRef = useRef<() => void>(() => {})
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const connectionName = currentConn?.name ?? ''
  const isReadonly = currentConn?.readonly ?? false
  const sqlHistory = useSqlHistory()
  const activeResult = results.find(item => item.id === activeResultId) ?? results[0] ?? null

  useEffect(() => {
    if (!connId) return
    api.resources(connId)
      .then(nodes => setDatabases(nodes.filter(n => n.type === 'sql_database').map(n => n.name)))
      .catch(e => showToast(errorMessage(e)))
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
      const run: QueryRun = {
        id: queryId,
        sql: sql.trim(),
        database: selectedDb,
        result: r,
        executedAt: new Date().toISOString(),
      }
      setResults(prev => [run, ...prev].slice(0, 8))
      setActiveResultId(run.id)
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
        const msg = errorMessage(e, 'Query failed')
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
    setFavoriteHistoryIds(new Set())
    saveFavoriteHistoryIds(new Set())
  }

  async function handleCancelQuery() {
    if (!activeQueryId) return
    try {
      await api.cancelQuery(activeQueryId)
      showToast(t('query.cancelled'), 'success')
    } catch (e) {
      showToast(errorMessage(e, 'Cancel failed'), 'error')
    }
  }

  const activeTabLabel = activeResult
    ? `(${activeResult.result.type === 'result_set' ? activeResult.result.rows.length : activeResult.result.rowsAffected} ${
        activeResult.result.type === 'result_set' ? 'rows' : 'affected'
      })`
    : ''

  const filteredHistory = useMemo(() => {
    const query = historySearch.trim().toLowerCase()
    return sqlHistory.entries.filter(entry => {
      if (historyFavoritesOnly && !favoriteHistoryIds.has(entry.id)) return false
      if (!query) return true
      return (
        entry.sql.toLowerCase().includes(query) ||
        entry.database.toLowerCase().includes(query) ||
        entry.connectionName.toLowerCase().includes(query)
      )
    })
  }, [favoriteHistoryIds, historyFavoritesOnly, historySearch, sqlHistory.entries])

  function toggleFavoriteHistory(entryId: string) {
    setFavoriteHistoryIds(prev => {
      const next = new Set(prev)
      if (next.has(entryId)) next.delete(entryId)
      else next.add(entryId)
      saveFavoriteHistoryIds(next)
      return next
    })
  }

  function closeResult(runId: string) {
    setResults(prev => {
      const next = prev.filter(item => item.id !== runId)
      if (activeResultId === runId) setActiveResultId(next[0]?.id ?? null)
      return next
    })
  }

  return (
    <PageShell>
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={[t('query.title')]}
        datasourceType={currentConn?.driver ?? 'mysql'}
        readonly={isReadonly}
      />
      <PageToolbar
        leading={
          <select
            value={selectedDb}
            onChange={e => setSelectedDb(e.target.value)}
            className="select max-w-60"
          >
            <option value="">{t('query.db_placeholder')}</option>
            {databases.map(db => <option key={db} value={db}>{db}</option>)}
          </select>
        }
        trailing={
          <>
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
          </>
        }
      />
      <PageBody>
        <div className="flex h-full flex-col gap-3 p-4">
          <CodePanel height={240}>
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
          </CodePanel>

          {error && (
            <ErrorStatePanel
              compact
              title="Query Error"
              message={error}
              action={
                <button type="button" onClick={() => handleRun(false)} className="btn btn-ghost h-8 px-3 text-xs">
                  {t('resource.retry')}
                </button>
              }
            />
          )}

          <div className="flex min-h-0 flex-1 flex-col">
            <div className="mb-3 flex gap-1 border-b" style={{ borderColor: 'var(--border-subtle)' }}>
              <button
                onClick={() => setActiveTab('result')}
                className="tab-btn rounded-b-none"
                data-active={activeTab === 'result'}
              >
                {t('query.tab.result')} {activeTabLabel}
              </button>
              <button
                onClick={() => setActiveTab('history')}
                className="tab-btn rounded-b-none"
                data-active={activeTab === 'history'}
              >
                {t('history.tab')} ({sqlHistory.entries.length})
              </button>
            </div>

            <div className="min-h-0 flex-1 overflow-auto">
              {activeTab === 'result' && (
                activeResult ? (
                  <div className="flex h-full flex-col gap-3">
                    <div className="flex shrink-0 gap-1 overflow-x-auto border-b pb-2" style={{ borderColor: 'var(--border-subtle)' }}>
                      {results.map((run, idx) => (
                        <button
                          key={run.id}
                          type="button"
                          onClick={() => setActiveResultId(run.id)}
                          className="group inline-flex h-8 shrink-0 items-center gap-2 rounded-md border px-2 text-xs"
                          style={{
                            borderColor: run.id === activeResult.id ? 'var(--accent)' : 'var(--border-subtle)',
                            color: run.id === activeResult.id ? 'var(--accent)' : 'var(--text-muted)',
                            background: run.id === activeResult.id ? 'var(--bg-selected)' : 'transparent',
                          }}
                        >
                          <span>{t('query.result_tab', { n: results.length - idx })}</span>
                          <span className="max-w-32 truncate font-mono">{run.database || t('query.db_placeholder')}</span>
                          <span
                            role="button"
                            tabIndex={0}
                            title={t('query.result.close')}
                            aria-label={t('query.result.close')}
                            className="grid h-4 w-4 place-items-center rounded opacity-60 hover:bg-[var(--bg-hover)] hover:opacity-100"
                            onClick={(event) => {
                              event.stopPropagation()
                              closeResult(run.id)
                            }}
                            onKeyDown={(event) => {
                              if (event.key !== 'Enter' && event.key !== ' ') return
                              event.preventDefault()
                              event.stopPropagation()
                              closeResult(run.id)
                            }}
                          >
                            <XIcon className="h-3 w-3" />
                          </span>
                        </button>
                      ))}
                    </div>
                    <div className="min-h-0 flex-1 overflow-auto">
                      {activeResult.result.type === 'result_set' ? (
                        <ResultTable
                          columns={activeResult.result.columns}
                          rows={activeResult.result.rows}
                          duration={activeResult.result.executionTimeMs}
                          truncated={activeResult.result.truncated}
                        />
                      ) : (
                        <StatusBanner tone="success">
                          {activeResult.result.rowsAffected} {t('query.rows_affected')} · {activeResult.result.executionTimeMs}ms
                          {activeResult.result.lastInsertId > 0 && (
                            <span className="ml-2">
                              · {t('query.last_insert_id')}: {activeResult.result.lastInsertId}
                            </span>
                          )}
                        </StatusBanner>
                      )}
                    </div>
                  </div>
                ) : (
                  <EmptyStatePanel
                    title={t('query.results.empty')}
                    detail={t('query.run')}
                    compact
                  />
                )
              )}

              {activeTab === 'history' && (
                <div className="space-y-3">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <input
                      value={historySearch}
                      onChange={event => setHistorySearch(event.target.value)}
                      className="input h-8 max-w-sm text-xs"
                      placeholder={t('history.search_placeholder')}
                    />
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => setHistoryFavoritesOnly(v => !v)}
                        className={`btn h-8 px-3 text-xs ${historyFavoritesOnly ? 'btn-primary' : 'btn-ghost'}`}
                      >
                        {t('history.favorites_only')}
                      </button>
                      {sqlHistory.entries.length > 0 && (
                        <button
                          onClick={handleClearHistory}
                          className="btn btn-ghost h-8 px-3 text-xs"
                        >{t('history.clear_all')}</button>
                      )}
                    </div>
                  </div>
                  <div className="space-y-1">
                    {filteredHistory.map(h => (
                      <div
                        key={h.id}
                        className="group relative rounded-md border p-3 transition-colors hover:bg-[var(--bg-hover)]"
                        style={{ borderColor: 'var(--border-subtle)' }}
                      >
                        <div onClick={() => setSql(h.sql)} className="cursor-pointer pr-20">
                          <div className="mb-0.5 flex items-center justify-between gap-2">
                            <span className="truncate text-xs" style={{ color: 'var(--text-subtle)' }}>
                              {new Date(h.executedAt).toLocaleString()}
                              {h.database ? ` · ${h.database}` : ''}
                            </span>
                            <span className="shrink-0 text-xs" style={{ color: 'var(--text-subtle)' }}>{h.executionTimeMs}ms</span>
                          </div>
                          <pre className="truncate font-mono text-xs" style={{ color: 'var(--text-default)' }}>{h.sql}</pre>
                          {!h.success && (
                            <div className="mt-1 flex items-center gap-1.5 text-xs" style={{ color: 'var(--danger)' }}>
                              <span className="status-dot" style={{ background: 'var(--danger)' }} />
                              Failed
                            </div>
                          )}
                        </div>
                        <div className="absolute right-2 top-2 flex gap-1 opacity-0 group-hover:opacity-100 group-focus-within:opacity-100">
                          <button
                            onClick={() => toggleFavoriteHistory(h.id)}
                            className="btn btn-ghost h-7 px-2 text-[11px]"
                            title={favoriteHistoryIds.has(h.id) ? t('history.unfavorite') : t('history.favorite')}
                          >
                            {favoriteHistoryIds.has(h.id) ? t('history.unfavorite') : t('history.favorite')}
                          </button>
                          <button
                            onClick={() => handleDeleteHistory(h.id)}
                            className="icon-btn"
                            title={t('history.delete')}
                            aria-label={t('history.delete')}
                          >
                            <XIcon className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      </div>
                    ))}
                    {filteredHistory.length === 0 && (
                      <EmptyStatePanel
                        compact
                        title={sqlHistory.entries.length === 0 ? t('history.empty') : t('history.no_matches')}
                        detail={historySearch || undefined}
                      />
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </PageBody>
      <PageStatusBar
        left={selectedDb || t('query.db_placeholder')}
        center={running ? t('query.running') : activeResult ? `${activeTabLabel || ''} ${activeResult.result.executionTimeMs}ms` : t('history.empty')}
        right={`${sqlHistory.entries.length} ${t('history.tab').toLowerCase()}`}
      />

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
    </PageShell>
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
        <div className="mb-2 rounded px-3 py-1.5 text-xs"
          style={{ background: 'var(--bg-muted)', border: '1px solid var(--warning)', color: 'var(--warning)' }}>
          {t('query.truncated')}
        </div>
      )}
      <div className="text-xs mb-2" style={{ color: 'var(--text-subtle)' }}>{rows.length} rows · {duration}ms</div>
      <div className="data-table-wrap rounded-md">
        <table className="data-table">
          <thead>
            <tr>
              {columns.map(c => (
                <th key={c} className="text-left px-3 py-2 font-mono whitespace-nowrap"
                  style={{ fontSize: 11 }}>
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i}>
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
