import { useI18n } from '../i18nContext'
import type { DataSourceType } from '../api'
import { RefreshIcon } from './Icons'

interface WorkbenchHeaderProps {
  connectionName?: string
  resourcePath: string[]
  datasourceType: DataSourceType
  readonly?: boolean
  onRefresh?: () => void
  meta?: Record<string, unknown>
}

const DRIVER_BADGE: Record<DataSourceType, { label: string; color: string }> = {
  mysql: { label: 'MY', color: '#3b82f6' },
  sqlite: { label: 'SQ', color: '#8b5cf6' },
  redis: { label: 'RD', color: '#ef4444' },
  mongodb: { label: 'MG', color: '#22c55e' },
  elasticsearch: { label: 'ES', color: '#f59e0b' },
}

export default function WorkbenchHeader({
  connectionName,
  resourcePath,
  datasourceType,
  readonly,
  onRefresh,
  meta,
}: WorkbenchHeaderProps) {
  const { t } = useI18n()
  const badge = DRIVER_BADGE[datasourceType]

  return (
    <div
      className="px-5 py-3 border-b flex items-center gap-3 shrink-0"
      style={{
        borderColor: 'var(--border-subtle)',
        background: 'color-mix(in srgb, var(--bg-surface) 92%, transparent)',
        backdropFilter: 'blur(10px)',
      }}
    >
      <span
        className="text-xs font-mono font-bold px-2 py-1 rounded-md"
        style={{ background: badge.color + '1f', color: badge.color, border: `1px solid ${badge.color}33` }}
      >
        {badge.label}
      </span>

      {connectionName && (
        <span
          className="badge font-mono"
          style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
        >
          {connectionName}
        </span>
      )}

      <div className="flex items-center gap-1.5 text-sm min-w-0 flex-1">
        {resourcePath.map((seg, i) => (
          <span key={i} className="flex items-center gap-1.5 min-w-0">
            {i > 0 && <span style={{ color: 'var(--text-subtle)' }}>/</span>}
            <span
              className="font-mono truncate"
              style={{ color: 'var(--text-strong)' }}
              title={seg}
            >
              {seg}
            </span>
          </span>
        ))}
      </div>

      {readonly && (
        <span
          className="text-xs px-2 py-0.5 rounded shrink-0"
          style={{
            background: 'var(--warning)18',
            color: 'var(--warning)',
            border: '1px solid var(--warning)44',
          }}
        >
          {t('workbench.readonly')}
        </span>
      )}

      {meta && Object.keys(meta).length > 0 && (
        <div className="flex items-center gap-2 shrink-0">
          {meta.rowCount !== undefined && (
            <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
              {String(meta.rowCount)} rows
            </span>
          )}
          {meta.docsCount !== undefined && (
            <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
              {String(meta.docsCount)} docs
            </span>
          )}
          {meta.keyCount !== undefined && (
            <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
              {String(meta.keyCount)} keys
            </span>
          )}
        </div>
      )}

      {onRefresh && (
        <button
          onClick={onRefresh}
          className="btn btn-ghost shrink-0"
          title={t('workbench.refresh')}
        >
          <RefreshIcon className="h-3.5 w-3.5" />
          {t('workbench.refresh')}
        </button>
      )}
    </div>
  )
}
