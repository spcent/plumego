import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { useI18n } from '../i18nContext'
import { sqlHistorySettings } from '../hooks/useSqlHistory'
import { api, type ActiveOperation, type AuditEvent, type Connection, type PoolStats } from '../api'
import { useToast } from '../components/toastContext'
import { PageBody, PageShell, PageStatusBar, PageToolbar } from '../components/workbench'

export default function SettingsPage() {
  const { t } = useI18n()
  const toast = useToast()
  const [historyEnabled, setHistoryEnabled] = useState(() => sqlHistorySettings.get().sqlHistoryEnabled)
  const [operations, setOperations] = useState<ActiveOperation[]>([])
  const [operationsLoading, setOperationsLoading] = useState(false)
  const [connections, setConnections] = useState<Connection[]>([])
  const [poolStats, setPoolStats] = useState<PoolStats | null>(null)
  const [runtimeLoading, setRuntimeLoading] = useState(false)
  const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([])
  const [me, setMe] = useState<{ user: string; role: string } | null>(null)

  function toggleHistory(val: boolean) {
    setHistoryEnabled(val)
    sqlHistorySettings.setEnabled(val)
  }

  const loadOperations = useCallback(async () => {
    setOperationsLoading(true)
    try {
      setOperations(await api.listActiveOperations())
    } catch (err) {
      toast.showToast(err instanceof Error ? err.message : 'Failed to load operations', 'error')
    } finally {
      setOperationsLoading(false)
    }
  }, [toast])

  const loadRuntime = useCallback(async () => {
    setRuntimeLoading(true)
    try {
      const [conns, stats] = await Promise.all([
        api.listConnections(),
        api.poolStats(),
      ])
      setConnections(conns)
      setPoolStats(stats)
    } catch (err) {
      toast.showToast(err instanceof Error ? err.message : 'Failed to load runtime state', 'error')
    } finally {
      setRuntimeLoading(false)
    }
  }, [toast])

  const loadAudit = useCallback(async () => {
    try {
      const [profile, events] = await Promise.all([
        api.me(),
        api.listAuditEvents(),
      ])
      setMe(profile)
      setAuditEvents(events)
    } catch (err) {
      toast.showToast(err instanceof Error ? err.message : 'Failed to load audit events', 'error')
    }
  }, [toast])

  useEffect(() => {
    const timer = window.setInterval(() => {
      void loadOperations()
    }, 5000)
    queueMicrotask(() => {
      void loadOperations()
      void loadRuntime()
      void loadAudit()
    })
    return () => window.clearInterval(timer)
  }, [loadAudit, loadOperations, loadRuntime])

  async function cancelOperation(operationId: string) {
    try {
      await api.cancelOperation(operationId)
      await loadOperations()
    } catch (err) {
      toast.showToast(err instanceof Error ? err.message : 'Failed to cancel operation', 'error')
    }
  }

  async function closeRuntime(connId: string) {
    try {
      await api.closeConnectionRuntime(connId)
      await loadRuntime()
    } catch (err) {
      toast.showToast(err instanceof Error ? err.message : 'Failed to close connection runtime', 'error')
    }
  }

  return (
    <PageShell>
      <PageToolbar
        leading={
          <div className="min-w-0">
            <h1 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>
              {t('settings.title')}
            </h1>
            <div className="mt-0.5 text-xs" style={{ color: 'var(--text-muted)' }}>
              {me ? `${me.user} · ${me.role}` : t('workbench.connection')}
            </div>
          </div>
        }
        trailing={
          <button
            type="button"
            onClick={() => {
              void loadOperations()
              void loadRuntime()
              void loadAudit()
            }}
            className="btn btn-ghost h-8 px-3 text-xs"
          >
            {t('settings.operations_refresh')}
          </button>
        }
      />

      <PageBody scroll>
        <div className="mx-auto grid w-full max-w-5xl gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(320px,0.7fr)]">
          <div className="space-y-4">
            <SectionPanel
              title={t('settings.connection_lifecycle')}
              action={
                <button
                  type="button"
                  onClick={() => void loadRuntime()}
                  disabled={runtimeLoading}
                  className="btn btn-ghost h-8 px-3 text-xs disabled:opacity-50"
                >
                  {runtimeLoading ? t('settings.operations_loading') : t('settings.operations_refresh')}
                </button>
              }
            >
              {poolStats && (
                <div className="mb-3 grid grid-cols-2 gap-2 text-xs md:grid-cols-4">
                  <Metric label="SQL" value={String(poolStats.sql_connections?.length ?? 0)} />
                  <Metric label="Redis" value={String(poolStats.redis_connections)} />
                  <Metric label="MongoDB" value={String(poolStats.mongodb_connections)} />
                  <Metric label="ES" value={String(poolStats.es_connections)} />
                </div>
              )}
              <div className="space-y-2">
                {connections.map(conn => {
                  const sqlStats = poolStats?.sql_connections?.find(s => s.connection_id === conn.id)
                  return (
                    <div
                      key={conn.id}
                      className="flex items-center gap-3 rounded-md border p-3"
                      style={{ borderColor: 'var(--border-subtle)' }}
                    >
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm font-medium" style={{ color: 'var(--text-default)' }}>
                          {conn.name}
                        </div>
                        <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
                          {conn.driver}
                          {sqlStats ? ` · open ${sqlStats.open} · in use ${sqlStats.in_use} · idle ${sqlStats.idle}` : ''}
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => void closeRuntime(conn.id)}
                        className="btn btn-ghost h-8 px-3 text-xs"
                      >
                        {t('settings.close_runtime')}
                      </button>
                    </div>
                  )
                })}
              </div>
            </SectionPanel>

            <SectionPanel
              title={t('settings.active_operations')}
              action={
                <button
                  type="button"
                  onClick={() => void loadOperations()}
                  disabled={operationsLoading}
                  className="btn btn-ghost h-8 px-3 text-xs disabled:opacity-50"
                >
                  {operationsLoading ? t('settings.operations_loading') : t('settings.operations_refresh')}
                </button>
              }
            >
              {operations.length === 0 ? (
                <EmptyLine>{t('settings.operations_empty')}</EmptyLine>
              ) : (
                <div className="space-y-2">
                  {operations.map(op => (
                    <div
                      key={op.operationId}
                      className="flex items-center gap-3 rounded-md border p-3"
                      style={{ borderColor: 'var(--border-subtle)' }}
                    >
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                          <span className="badge h-5 min-h-0 px-1.5 text-[10px]">{op.driver}</span>
                          <span>{op.kind}</span>
                          <span>{op.duration}</span>
                        </div>
                        <div className="mt-1 truncate text-sm font-medium" style={{ color: 'var(--text-default)' }}>
                          {op.resource || op.connId}
                        </div>
                        <div className="truncate text-xs" style={{ color: 'var(--text-muted)' }}>
                          {op.summary}
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => void cancelOperation(op.operationId)}
                        className="btn btn-danger h-8 px-3 text-xs"
                      >
                        {t('settings.operations_cancel')}
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </SectionPanel>
          </div>

          <div className="space-y-4">
            <SectionPanel title={t('settings.sql_history')}>
              <label className="flex cursor-pointer items-start gap-3">
                <input
                  type="checkbox"
                  className="mt-1 shrink-0"
                  checked={historyEnabled}
                  onChange={e => toggleHistory(e.target.checked)}
                />
                <div>
                  <div className="text-sm" style={{ color: 'var(--text-default)' }}>
                    {t('settings.sql_history_enabled')}
                  </div>
                  <p className="mt-1 text-xs leading-relaxed" style={{ color: 'var(--text-muted)' }}>
                    {t('settings.sql_history_hint')}
                  </p>
                </div>
              </label>
            </SectionPanel>

            <SectionPanel
              title={t('settings.audit_rbac')}
              action={
                <button
                  type="button"
                  onClick={() => void loadAudit()}
                  className="btn btn-ghost h-8 px-3 text-xs"
                >
                  {t('settings.operations_refresh')}
                </button>
              }
            >
              {auditEvents.length === 0 ? (
                <EmptyLine>{t('settings.audit_empty')}</EmptyLine>
              ) : (
                <div className="max-h-[420px] space-y-2 overflow-auto pr-1">
                  {auditEvents.map(event => (
                    <div
                      key={event.id}
                      className="rounded-md border p-3"
                      style={{ borderColor: 'var(--border-subtle)' }}
                    >
                      <div className="flex items-center justify-between gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                        <span>{new Date(event.created_at).toLocaleString()}</span>
                        <span className="badge h-5 min-h-0 px-1.5 text-[10px]">{event.status}</span>
                      </div>
                      <div className="mt-1 truncate font-mono text-sm" style={{ color: 'var(--text-default)' }}>
                        {event.action}
                      </div>
                      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
                        {event.user}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </SectionPanel>
          </div>
        </div>
      </PageBody>

      <PageStatusBar
        left={`${operations.length} ${t('settings.active_operations')}`}
        right={poolStats ? `${connections.length} ${t('connections.title')}` : undefined}
      />
    </PageShell>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border p-3" style={{ borderColor: 'var(--border-subtle)', background: 'var(--bg-muted)' }}>
      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>{label}</div>
      <div className="text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>{value}</div>
    </div>
  )
}

function SectionPanel({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="panel-flat p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold" style={{ color: 'var(--text-default)' }}>
          {title}
        </h2>
        {action}
      </div>
      {children}
    </section>
  )
}

function EmptyLine({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-md border border-dashed px-3 py-6 text-center text-sm" style={{ borderColor: 'var(--border-subtle)', color: 'var(--text-muted)' }}>
      {children}
    </div>
  )
}
