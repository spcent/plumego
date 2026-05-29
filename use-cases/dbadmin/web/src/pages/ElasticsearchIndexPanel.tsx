import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'

export default function ElasticsearchIndexPanel() {
  const { connId, esIndex } = useParams<{ connId: string; esIndex: string }>()
  const { t } = useI18n()
  const conn = useCurrentConn(connId)

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-4 py-3 border-b flex items-center gap-3" style={{ borderColor: 'var(--border-subtle)' }}>
        <span className="text-lg font-mono" style={{ color: '#fbbf24', fontWeight: 700 }}>E</span>
        <h2 className="text-base font-semibold" style={{ color: 'var(--text-strong)' }}>
          {esIndex}
        </h2>
        {conn && (
          <span className="text-xs font-mono px-2 py-0.5 rounded"
            style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {conn.name}
          </span>
        )}
        {conn?.readonly && (
          <span className="text-xs px-2 py-0.5 rounded"
            style={{ background: 'var(--warning)18', color: 'var(--warning)', border: '1px solid var(--warning)44' }}>
            {t('elasticsearch.readonly')}
          </span>
        )}
      </div>

      {/* Placeholder content */}
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-3">
          <div className="text-4xl" style={{ color: 'var(--text-muted)' }}>🚧</div>
          <div className="text-lg font-medium" style={{ color: 'var(--text-strong)' }}>
            {t('elasticsearch.coming_soon')}
          </div>
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            {t('elasticsearch.coming_soon_hint')}
          </div>
          <div className="text-xs px-4 py-2 rounded" style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {t('elasticsearch.p0_features')}
          </div>
        </div>
      </div>
    </div>
  )
}
