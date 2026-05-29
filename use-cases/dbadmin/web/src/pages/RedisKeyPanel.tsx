import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'

export default function RedisKeyPanel() {
  const { connId, redisDb } = useParams<{ connId: string; redisDb: string }>()
  const { t } = useI18n()
  const conn = useCurrentConn(connId)

  return (
    <div className="flex flex-col items-center justify-center h-full gap-5 p-8">
      <div className="text-5xl select-none" style={{ color: '#f87171' }}>⚡</div>
      <h2 className="text-xl font-bold" style={{ color: 'var(--text-strong)' }}>
        {t('redis.panel.title')}
      </h2>
      {conn && (
        <div className="text-sm font-mono px-3 py-1.5 rounded"
          style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
          {conn.name} · DB {redisDb ?? '0'}
        </div>
      )}
      <p className="text-sm text-center max-w-sm" style={{ color: 'var(--text-muted)' }}>
        {t('redis.panel.hint')}
      </p>
      <div className="text-xs px-4 py-2 rounded"
        style={{ background: 'var(--warning)18', color: 'var(--warning)', border: '1px solid var(--warning)44' }}>
        {t('redis.driver_coming_soon')}
      </div>
    </div>
  )
}
