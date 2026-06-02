import { useEffect, useMemo, useSyncExternalStore, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import type { Connection } from '../api'
import { useI18n } from '../i18nContext'
import {
  DatabaseIcon,
  ElasticsearchIcon,
  MongoDBIcon,
  MySQLIcon,
  RedisIcon,
  SQLiteIcon,
  TableIcon,
  XIcon,
} from './Icons'
import {
  buildRecentResourceDescriptor,
  clearRecentResources,
  getRecentEntriesSnapshot,
  subscribeRecentResources,
  trackRecentResource,
  type RecentResourceDescriptor,
} from './recentResourcesModel'

function iconForDriver(driver: Connection['driver'] | undefined): ReactNode {
  const cls = 'h-3.5 w-3.5'
  switch (driver) {
    case 'mysql':
      return <MySQLIcon className={cls} />
    case 'sqlite':
      return <SQLiteIcon className={cls} />
    case 'redis':
      return <RedisIcon className={cls} />
    case 'mongodb':
      return <MongoDBIcon className={cls} />
    case 'elasticsearch':
      return <ElasticsearchIcon className={cls} />
    default:
      return <DatabaseIcon className={cls} />
  }
}

export default function RecentResources({
  connections,
  limit = 5,
}: {
  connections: Connection[]
  limit?: number
}) {
  const location = useLocation()
  const navigate = useNavigate()
  const { t } = useI18n()
  const entries = useSyncExternalStore(
    subscribeRecentResources,
    getRecentEntriesSnapshot,
    getRecentEntriesSnapshot,
  )

  const currentPath = `${location.pathname}${location.search}`
  const currentDescriptor = useMemo(
    () => buildRecentResourceDescriptor(currentPath, connections, t),
    [connections, currentPath, t],
  )
  const currentTrackablePath = currentDescriptor?.path

  useEffect(() => {
    if (!currentTrackablePath) return
    trackRecentResource(currentTrackablePath)
  }, [currentTrackablePath])

  const recent = useMemo(() => {
    return entries
      .map(entry => buildRecentResourceDescriptor(entry.path, connections, t))
      .filter((entry): entry is RecentResourceDescriptor => Boolean(entry))
      .slice(0, limit)
  }, [connections, entries, limit, t])

  if (recent.length === 0) return null

  return (
    <section className="shrink-0 px-2 py-2" style={{ borderBottom: '1px solid var(--sb-border)' }}>
      <div className="mb-1 flex items-center justify-between px-1">
        <div className="text-[10px] font-semibold uppercase tracking-[0.14em]" style={{ color: 'var(--sb-muted)' }}>
          {t('recent.title')}
        </div>
        <button
          type="button"
          className="grid h-5 w-5 place-items-center rounded transition-colors hover:bg-[var(--sb-hover)]"
          style={{ color: 'var(--sb-muted)' }}
          title={t('recent.clear')}
          aria-label={t('recent.clear')}
          onClick={() => {
            clearRecentResources()
          }}
        >
          <XIcon className="h-3 w-3" />
        </button>
      </div>

      <div className="space-y-0.5">
        {recent.map(item => {
          const active = item.path === currentPath
          return (
            <button
              key={item.activeKey}
              type="button"
              className="grid h-[34px] w-full min-w-0 items-center rounded-md px-1.5 text-left transition-colors hover:bg-[var(--sb-hover)] active:translate-y-px"
              style={{
                gridTemplateColumns: '22px minmax(0, 1fr)',
                columnGap: 7,
                color: active ? '#fff' : 'var(--sb-text)',
                background: active ? 'var(--sb-active)' : 'transparent',
              }}
              title={`${item.label} - ${item.detail}`}
              onClick={() => navigate(item.path)}
            >
              <span
                className="grid h-5 w-5 place-items-center rounded"
                style={{
                  background: active ? 'rgba(255,255,255,0.16)' : 'var(--sb-surface)',
                  border: `1px solid ${active ? 'rgba(255,255,255,0.22)' : 'var(--sb-border)'}`,
                  color: active ? '#fff' : 'var(--sb-muted)',
                }}
              >
                {item.driver ? iconForDriver(item.driver) : <TableIcon className="h-3.5 w-3.5" />}
              </span>
              <span className="min-w-0">
                <span className="block truncate text-[12px] font-medium leading-4">{item.label}</span>
                <span
                  className="block truncate text-[10px] leading-3"
                  style={{ color: active ? 'rgba(255,255,255,0.72)' : 'var(--sb-muted)' }}
                >
                  {item.detail}
                </span>
              </span>
            </button>
          )
        })}
      </div>
    </section>
  )
}
