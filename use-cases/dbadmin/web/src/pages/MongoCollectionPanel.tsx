import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'

export default function MongoCollectionPanel() {
  const { connId, mongoDb, mongoColl } = useParams<{ connId: string; mongoDb: string; mongoColl?: string }>()
  const { t } = useI18n()
  const conn = useCurrentConn(connId)
  const isReadonly = conn?.readonly ?? false

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-4 py-3 border-b flex items-center gap-3" style={{ borderColor: 'var(--border-subtle)' }}>
        <span className="text-lg font-mono" style={{ color: '#4ade80' }}>M</span>
        <h2 className="text-base font-semibold font-mono" style={{ color: 'var(--text-strong)' }}>
          {mongoDb}{mongoColl ? ` / ${mongoColl}` : ''}
        </h2>
        {conn && (
          <span className="text-xs font-mono px-2 py-0.5 rounded"
            style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {conn.name}
          </span>
        )}
        {isReadonly && (
          <span className="text-xs px-2 py-0.5 rounded"
            style={{ background: 'var(--warning)18', color: 'var(--warning)', border: '1px solid var(--warning)44' }}>
            {t('mongodb.readonly')}
          </span>
        )}
      </div>

      {/* Placeholder content */}
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center space-y-3 max-w-md">
          <div className="text-5xl">🚧</div>
          <div className="text-lg font-medium" style={{ color: 'var(--text-strong)' }}>
            {t('mongodb.coming_soon')}
          </div>
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            {t('mongodb.coming_soon_hint')}
          </div>
          <div className="text-xs px-4 py-2 rounded" style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {t('mongodb.p0_features')}
          </div>
        </div>
      </div>
    </div>
  )
}
