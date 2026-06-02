import { useCallback, useEffect, useState } from 'react'
import { useI18n } from '../i18nContext'
import { sqlHistorySettings } from '../hooks/useSqlHistory'
import { api, type ActiveOperation, type AuditEvent, type Connection, type PoolStats } from '../api'
import { useToast } from '../components/toastContext'

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
    <div className="p-6 max-w-3xl space-y-4">
      <h1 className="text-xl font-bold mb-6" style={{ color: 'var(--text-strong)' }}>
        {t('settings.title')}
      </h1>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <h2 className="text-sm font-semibold mb-3" style={{ color: 'var(--text-default)' }}>
          {t('settings.sql_history')}
        </h2>
        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="checkbox"
            className="mt-0.5 shrink-0"
            checked={historyEnabled}
            onChange={e => toggleHistory(e.target.checked)}
          />
          <div>
            <div className="text-sm" style={{ color: 'var(--text-default)' }}>
              {t('settings.sql_history_enabled')}
            </div>
            <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>
              {t('settings.sql_history_hint')}
            </p>
          </div>
        </label>
      </section>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <div className="flex items-center justify-between gap-3 mb-3">
          <div>
            <h2 className="text-sm font-semibold" style={{ color: 'var(--text-default)' }}>
              {t('settings.audit_rbac')}
            </h2>
            <div className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
              {me ? `${me.user} · ${me.role}` : ''}
            </div>
          </div>
          <button
            type="button"
            onClick={() => void loadAudit()}
            className="px-2 py-1 text-xs rounded border"
            style={{ color: 'var(--text-default)', borderColor: 'var(--border-subtle)' }}
          >
            {t('settings.operations_refresh')}
          </button>
        </div>
        {auditEvents.length === 0 ? (
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            {t('settings.audit_empty')}
          </div>
        ) : (
          <div className="space-y-2 max-h-72 overflow-auto">
            {auditEvents.map(event => (
              <div
                key={event.id}
                className="rounded border p-3"
                style={{ borderColor: 'var(--border-subtle)' }}
              >
                <div className="flex items-center justify-between gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                  <span>{new Date(event.created_at).toLocaleString()}</span>
                  <span>{event.status}</span>
                </div>
                <div className="text-sm font-mono truncate mt-1" style={{ color: 'var(--text-default)' }}>
                  {event.action}
                </div>
                <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
                  {event.user}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <div className="flex items-center justify-between gap-3 mb-3">
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text-default)' }}>
            {t('settings.connection_lifecycle')}
          </h2>
          <button
            type="button"
            onClick={() => void loadRuntime()}
            disabled={runtimeLoading}
            className="px-2 py-1 text-xs rounded border"
            style={{
              color: 'var(--text-default)',
              borderColor: 'var(--border-subtle)',
              opacity: runtimeLoading ? 0.6 : 1,
            }}
          >
            {runtimeLoading ? t('settings.operations_loading') : t('settings.operations_refresh')}
          </button>
        </div>
        {poolStats && (
          <div className="grid grid-cols-4 gap-2 mb-3 text-xs">
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
                className="flex items-center gap-3 rounded border p-3"
                style={{ borderColor: 'var(--border-subtle)' }}
              >
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate" style={{ color: 'var(--text-default)' }}>
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
                  className="px-2 py-1 text-xs rounded border"
                  style={{ color: 'var(--text-default)', borderColor: 'var(--border-subtle)' }}
                >
                  {t('settings.close_runtime')}
                </button>
              </div>
            )
          })}
        </div>
      </section>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <div className="flex items-center justify-between gap-3 mb-3">
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text-default)' }}>
            {t('settings.active_operations')}
          </h2>
          <button
            type="button"
            onClick={() => void loadOperations()}
            disabled={operationsLoading}
            className="px-2 py-1 text-xs rounded border"
            style={{
              color: 'var(--text-default)',
              borderColor: 'var(--border-subtle)',
              opacity: operationsLoading ? 0.6 : 1,
            }}
          >
            {operationsLoading ? t('settings.operations_loading') : t('settings.operations_refresh')}
          </button>
        </div>
        {operations.length === 0 ? (
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            {t('settings.operations_empty')}
          </div>
        ) : (
          <div className="space-y-2">
            {operations.map(op => (
              <div
                key={op.operationId}
                className="flex items-center gap-3 rounded border p-3"
                style={{ borderColor: 'var(--border-subtle)' }}
              >
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                    <span>{op.driver}</span>
                    <span>{op.kind}</span>
                    <span>{op.duration}</span>
                  </div>
                  <div className="text-sm font-medium truncate mt-1" style={{ color: 'var(--text-default)' }}>
                    {op.resource || op.connId}
                  </div>
                  <div className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>
                    {op.summary}
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => void cancelOperation(op.operationId)}
                  className="px-2 py-1 text-xs rounded"
                  style={{ color: 'white', background: 'var(--danger)' }}
                >
                  {t('settings.operations_cancel')}
                </button>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded border p-2" style={{ borderColor: 'var(--border-subtle)' }}>
      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>{label}</div>
      <div className="text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>{value}</div>
    </div>
  )
}
