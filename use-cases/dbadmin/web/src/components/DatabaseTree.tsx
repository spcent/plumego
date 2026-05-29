import { useState, useCallback } from 'react'
import { Link, useParams, useLocation } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useI18n } from '../i18n'

interface TableEntry {
  name: string
}

interface Props {
  connections: Connection[]
  onRefresh: () => void
}

function DriverBadge({ driver }: { driver: string }) {
  return (
    <span
      className="shrink-0 text-[10px] font-mono px-1 py-px rounded leading-none"
      style={{
        background: 'var(--sb-surface)',
        color: 'var(--sb-muted)',
        border: '1px solid var(--sb-border)',
      }}
    >
      {driver === 'mysql' ? 'MY' : 'SQ'}
    </span>
  )
}

function TreeItem({
  depth,
  icon,
  label,
  active,
  expanded,
  expandable,
  onClick,
  to,
  extra,
}: {
  depth: number
  icon?: React.ReactNode
  label: string
  active?: boolean
  expanded?: boolean
  expandable?: boolean
  onClick?: () => void
  to?: string
  extra?: React.ReactNode
}) {
  const paddingLeft = 12 + depth * 16
  const cls = [
    'flex items-center gap-1.5 h-[28px] w-full text-[13px] rounded-sm mx-1 pr-2 cursor-pointer select-none truncate',
    'transition-colors duration-75',
    active
      ? 'font-medium'
      : '',
  ].join(' ')

  const style: React.CSSProperties = {
    paddingLeft,
    width: 'calc(100% - 8px)',
    color: active ? '#fff' : 'var(--sb-text)',
    background: active ? 'var(--sb-active)' : 'transparent',
  }

  const inner = (
    <>
      {expandable && (
        <span className="shrink-0 text-[10px] w-3 text-center" style={{ color: 'var(--sb-muted)' }}>
          {expanded ? '▾' : '▸'}
        </span>
      )}
      {icon && <span className="shrink-0">{icon}</span>}
      <span className="truncate flex-1 min-w-0">{label}</span>
      {extra}
    </>
  )

  if (to) {
    return (
      <Link to={to} className={cls} style={style} title={label}>
        {inner}
      </Link>
    )
  }

  return (
    <button onClick={onClick} className={cls} style={style} title={label}>
      {inner}
    </button>
  )
}

export default function DatabaseTree({ connections, onRefresh: _onRefresh }: Props) {
  const params = useParams<{ connId?: string; dbName?: string; tableName?: string }>()
  const location = useLocation()
  const { t } = useI18n()

  const [expandedConns, setExpandedConns] = useState<Record<string, boolean>>({})
  const [expandedDbs, setExpandedDbs] = useState<Record<string, boolean>>({})
  const [databases, setDatabases] = useState<Record<string, string[]>>({})
  const [tables, setTables] = useState<Record<string, TableEntry[]>>({})

  const toggleConn = useCallback(async (connId: string) => {
    const next = !expandedConns[connId]
    setExpandedConns(p => ({ ...p, [connId]: next }))
    if (next && !databases[connId]) {
      try {
        const dbs = await api.databases(connId)
        setDatabases(p => ({ ...p, [connId]: dbs }))
      } catch {}
    }
  }, [expandedConns, databases])

  const toggleDb = useCallback(async (connId: string, dbName: string) => {
    const key = `${connId}:${dbName}`
    const next = !expandedDbs[key]
    setExpandedDbs(p => ({ ...p, [key]: next }))
    if (next && !tables[key]) {
      try {
        const result = await api.tables(connId, dbName)
        setTables(p => ({ ...p, [key]: result.map((r: { name: string }) => ({ name: r.name })) }))
      } catch {}
    }
  }, [expandedDbs, tables])

  return (
    <nav className="flex-1 overflow-y-auto py-1 overflow-x-hidden">
      {connections.length === 0 && (
        <div className="px-4 py-3 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
          No connections
        </div>
      )}
      {connections.map(conn => {
        const connExpanded = expandedConns[conn.id]
        const isQueryPage = location.pathname.includes('/query')
        const connActive = params.connId === conn.id && isQueryPage

        return (
          <div key={conn.id}>
            <TreeItem
              depth={0}
              expandable
              expanded={connExpanded}
              label={conn.name}
              icon={<DriverBadge driver={conn.driver} />}
              onClick={() => toggleConn(conn.id)}
              extra={
                conn.readonly ? (
                  <span
                    className="shrink-0 text-[10px] font-mono px-1 py-px rounded leading-none"
                    style={{ background: '#92400e44', color: '#fbbf24' }}
                  >
                    RO
                  </span>
                ) : undefined
              }
            />

            {connExpanded && (
              <>
                {/* SQL Console shortcut */}
                <TreeItem
                  depth={1}
                  icon={<span style={{ color: 'var(--sb-muted)', fontSize: 11 }}>▶</span>}
                  label={t('nav.sql_console')}
                  active={connActive}
                  to={`/conn/${conn.id}/query`}
                />

                {(databases[conn.id] || []).map(db => {
                  const dbKey = `${conn.id}:${db}`
                  const dbExpanded = expandedDbs[dbKey]
                  const dbActive = params.connId === conn.id && params.dbName === db && !isQueryPage

                  return (
                    <div key={db}>
                      <TreeItem
                        depth={1}
                        expandable
                        expanded={dbExpanded}
                        active={dbActive}
                        label={db}
                        icon={<span style={{ color: 'var(--sb-muted)', fontSize: 12 }}>🗄</span>}
                        onClick={() => toggleDb(conn.id, db)}
                      />

                      {dbExpanded && (
                        <>
                          {/* Tables page link */}
                          <TreeItem
                            depth={2}
                            icon={<span style={{ color: 'var(--sb-muted)', fontSize: 10 }}>⊞</span>}
                            label="Tables"
                            active={params.connId === conn.id && params.dbName === db && !params.tableName && !isQueryPage}
                            to={`/conn/${conn.id}/db/${db}/tables`}
                          />

                          {(tables[dbKey] || []).map(tbl => {
                            const tableActive = params.connId === conn.id && params.dbName === db && params.tableName === tbl.name
                            return (
                              <TreeItem
                                key={tbl.name}
                                depth={2}
                                icon={<span style={{ color: 'var(--sb-muted)', fontSize: 10 }}>▤</span>}
                                label={tbl.name}
                                active={tableActive}
                                to={`/conn/${conn.id}/db/${db}/tables/${tbl.name}/data`}
                              />
                            )
                          })}
                        </>
                      )}
                    </div>
                  )
                })}
              </>
            )}
          </div>
        )
      })}
    </nav>
  )
}
