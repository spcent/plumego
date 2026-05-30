import { useState, useCallback, useEffect, useRef } from 'react'
import { Link, useNavigate, useParams, useLocation } from 'react-router-dom'
import { api, type Connection, type ResourceNode, type ResourceNodeType } from '../api'
import { useI18n } from '../i18n'
import { useToast } from './Toast'

// ── Context Menu ───────────────────────────────────────────────────────────

interface ContextMenuItem {
  label: string
  onClick: () => void
}

interface ContextMenuState {
  x: number
  y: number
  items: ContextMenuItem[]
}

function ContextMenu({ state, onClose }: { state: ContextMenuState; onClose: () => void }) {
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [onClose])

  return (
    <div
      ref={menuRef}
      className="fixed z-50 bg-[var(--bg-surface)] border border-[var(--border-subtle)] rounded shadow-lg py-1 min-w-[160px]"
      style={{ left: state.x, top: state.y }}
    >
      {state.items.map((item, i) => (
        <button
          key={i}
          onClick={() => {
            item.onClick()
            onClose()
          }}
          className="w-full text-left px-3 py-1.5 text-sm hover:bg-[var(--bg-hover)] text-[var(--text-default)]"
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}

// ── Icon dispatch ──────────────────────────────────────────────────────────

function nodeIcon(type: ResourceNodeType): React.ReactNode {
  const s: React.CSSProperties = { color: 'var(--sb-muted)', fontSize: 12 }
  switch (type) {
    case 'sql_database':    return <span style={s}>🗄</span>
    case 'sql_table':       return <span style={{ ...s, fontSize: 10 }}>▤</span>
    case 'sql_view':        return <span style={{ ...s, fontSize: 10 }}>◧</span>
    case 'redis_db':        return <span style={{ ...s, color: '#f87171' }}>⬡</span>
    case 'redis_key':       return <span style={{ ...s, fontSize: 10, color: '#fb923c' }}>⬡</span>
    case 'mongo_database':  return <span style={{ ...s, color: '#4ade80', fontWeight: 700 }}>M</span>
    case 'mongo_collection':return <span style={{ ...s, fontSize: 10, color: '#22d3ee' }}>C</span>
    case 'es_index':        return <span style={{ ...s, color: '#fbbf24', fontWeight: 700 }}>E</span>
    case 'es_alias':        return <span style={{ ...s, fontSize: 10, color: '#a78bfa' }}>~</span>
    case 'es_data_stream':  return <span style={{ ...s, fontSize: 10, color: '#34d399' }}>↓</span>
    default:                return <span style={{ ...s, fontSize: 10 }}>•</span>
  }
}

// ── Navigation URL dispatch ────────────────────────────────────────────────

function nodeUrl(connId: string, node: ResourceNode): string | null {
  switch (node.type) {
    case 'sql_database':
      return `/conn/${connId}/db/${encodeURIComponent(node.path)}/tables`
    case 'sql_table':
    case 'sql_view': {
      const slash = node.path.indexOf('/')
      if (slash < 0) return null
      const db = node.path.slice(0, slash)
      const table = node.path.slice(slash + 1)
      return `/conn/${connId}/db/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/data`
    }
    case 'redis_db':
      return `/conn/${connId}/redis/${encodeURIComponent(node.path)}`
    case 'redis_key': {
      const slash = node.path.indexOf('/')
      if (slash < 0) return null
      const db = node.path.slice(0, slash)
      const key = node.path.slice(slash + 1)
      return `/conn/${connId}/redis/${encodeURIComponent(db)}?key=${encodeURIComponent(key)}`
    }
    case 'mongo_database':
      return `/conn/${connId}/mongo/${encodeURIComponent(node.path)}/collections`
    case 'mongo_collection': {
      const slash = node.path.indexOf('/')
      if (slash < 0) return null
      const db = node.path.slice(0, slash)
      const coll = node.path.slice(slash + 1)
      return `/conn/${connId}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents`
    }
    case 'es_index':
      return `/conn/${connId}/es/${encodeURIComponent(node.path)}`
    case 'es_alias':
      return `/conn/${connId}/es/alias/${encodeURIComponent(node.path)}`
    case 'es_data_stream':
      return `/conn/${connId}/es/data-stream/${encodeURIComponent(node.path)}`
    default:
      return null
  }
}

// ── TreeItem ───────────────────────────────────────────────────────────────

function DriverBadge({ driver }: { driver: string }) {
  const label =
    driver === 'mysql'  ? 'MY' :
    driver === 'sqlite' ? 'SQ' :
    driver === 'redis'  ? 'RD' :
    driver === 'mongodb'? 'MG' :
    driver === 'elasticsearch' ? 'ES' :
    driver.slice(0, 2).toUpperCase()
  return (
    <span
      className="shrink-0 text-[10px] font-mono px-1 py-px rounded leading-none"
      style={{
        background: 'var(--sb-surface)',
        color: 'var(--sb-muted)',
        border: '1px solid var(--sb-border)',
      }}
    >
      {label}
    </span>
  )
}

function TreeItem({
  depth, icon, label, active, expanded, expandable, onClick, to, extra, onContextMenu,
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
  onContextMenu?: (e: React.MouseEvent) => void
}) {
  const paddingLeft = 12 + depth * 16
  const cls = [
    'flex items-center gap-1.5 h-[28px] w-full text-[13px] rounded-sm mx-1 pr-2 cursor-pointer select-none truncate',
    'transition-colors duration-75',
    active ? 'font-medium' : '',
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
  if (to) return <Link to={to} className={cls} style={style} title={label} onContextMenu={onContextMenu}>{inner}</Link>
  return <button onClick={onClick} onContextMenu={onContextMenu} className={cls} style={style} title={label}>{inner}</button>
}

// ── ResourceExplorer ───────────────────────────────────────────────────────

interface Props {
  connections: Connection[]
  onRefresh: () => void
}

export default function ResourceExplorer({ connections, onRefresh: _onRefresh }: Props) {
  const params = useParams<{ connId?: string; dbName?: string; tableName?: string; redisDb?: string; mongoDb?: string; mongoColl?: string; esIndex?: string; esAlias?: string; esDataStream?: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const { t } = useI18n()
  const { showToast } = useToast()

  // Expanded state: connection IDs and node paths
  const [expandedConns, setExpandedConns] = useState<Record<string, boolean>>({})
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({})

  // Cached resource nodes: key = connId (top-level) or "connId:dbPath" (children)
  const [resourceCache, setResourceCache] = useState<Record<string, ResourceNode[]>>({})

  // Loading and error states
  const [loadingConns, setLoadingConns] = useState<Set<string>>(new Set())
  const [connErrors, setConnErrors] = useState<Record<string, string>>({})
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null)

  const isQueryPage = location.pathname.includes('/query')

  // Fetch top-level nodes (databases) for a connection
  const fetchTopLevel = useCallback(async (connId: string, forceRefresh = false) => {
    if (!forceRefresh && resourceCache[connId] !== undefined) return

    setLoadingConns(prev => new Set(prev).add(connId))
    setConnErrors(prev => {
      const next = { ...prev }
      delete next[connId]
      return next
    })

    try {
      const nodes = await api.resources(connId)
      setResourceCache(c => ({ ...c, [connId]: nodes }))
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      setConnErrors(prev => ({ ...prev, [connId]: msg }))
    } finally {
      setLoadingConns(prev => {
        const next = new Set(prev)
        next.delete(connId)
        return next
      })
    }
  }, [resourceCache])

  // Fetch child nodes (tables/views) under a database node
  const fetchChildren = useCallback(async (connId: string, node: ResourceNode, forceRefresh = false) => {
    const key = `${connId}:${node.path}`
    if (!forceRefresh && resourceCache[key] !== undefined) return
    try {
      const nodes = await api.resources(connId, node.id)
      setResourceCache(c => ({ ...c, [key]: nodes }))
    } catch {}
  }, [resourceCache])

  const toggleConn = useCallback((connId: string) => {
    const next = !expandedConns[connId]
    setExpandedConns(p => ({ ...p, [connId]: next }))
    if (next) fetchTopLevel(connId)
  }, [expandedConns, fetchTopLevel])

  const toggleNode = useCallback((connId: string, node: ResourceNode) => {
    const key = `${connId}:${node.path}`
    const next = !expandedNodes[key]
    setExpandedNodes(p => ({ ...p, [key]: next }))
    if (next) fetchChildren(connId, node)
  }, [expandedNodes, fetchChildren])

  const handleRefreshConn = useCallback((connId: string) => {
    setResourceCache(c => {
      const next = { ...c }
      delete next[connId]
      Object.keys(next).forEach(k => {
        if (k.startsWith(`${connId}:`)) delete next[k]
      })
      return next
    })
    fetchTopLevel(connId, true)
    showToast(t('resource.refresh'), 'success')
  }, [fetchTopLevel, showToast, t])

  const handleRefreshNode = useCallback((connId: string, node: ResourceNode) => {
    const key = `${connId}:${node.path}`
    setResourceCache(c => {
      const next = { ...c }
      delete next[key]
      return next
    })
    fetchChildren(connId, node, true)
    showToast(t('resource.refresh'), 'success')
  }, [fetchChildren, showToast, t])

  const handleContextMenu = useCallback((e: React.MouseEvent, items: ContextMenuItem[]) => {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, items })
  }, [])

  // Dispatch a click on a non-expandable node (future: redis_key, etc.)
  const handleNodeClick = useCallback((connId: string, node: ResourceNode) => {
    const url = nodeUrl(connId, node)
    if (url) navigate(url)
    // Future drivers: add custom click actions here based on node.datasource_type
  }, [navigate])

  return (
    <nav className="flex-1 overflow-y-auto py-1 overflow-x-hidden" onClick={() => setContextMenu(null)}>
      {connections.length === 0 && (
        <div className="px-4 py-3 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
          {t('connections.empty')}
        </div>
      )}

      {connections.map(conn => {
        const connExpanded = expandedConns[conn.id]
        const connActive = params.connId === conn.id && isQueryPage
        const topNodes: ResourceNode[] = resourceCache[conn.id] ?? []
        const isLoading = loadingConns.has(conn.id)
        const error = connErrors[conn.id]

        return (
          <div key={conn.id}>
            {/* Connection row */}
            <TreeItem
              depth={0}
              expandable
              expanded={connExpanded}
              label={conn.name}
              icon={<DriverBadge driver={conn.driver} />}
              onClick={() => toggleConn(conn.id)}
              onContextMenu={(e) => {
                handleContextMenu(e, [
                  { label: t('resource.refresh'), onClick: () => handleRefreshConn(conn.id) },
                ])
              }}
              extra={
                <>
                  {isLoading && (
                    <span className="shrink-0 text-[10px] animate-spin" style={{ color: 'var(--sb-muted)' }}>
                      ⟳
                    </span>
                  )}
                  {conn.readonly && (
                    <span
                      className="shrink-0 text-[10px] font-mono px-1 py-px rounded leading-none"
                      style={{ background: '#92400e44', color: '#fbbf24' }}
                    >
                      RO
                    </span>
                  )}
                </>
              }
            />

            {connExpanded && (
              <>
                {/* SQL Console shortcut — SQL drivers only */}
                {(conn.driver === 'mysql' || conn.driver === 'sqlite') && (
                  <TreeItem
                    depth={1}
                    icon={<span style={{ color: 'var(--sb-muted)', fontSize: 11 }}>▶</span>}
                    label={t('nav.sql_console')}
                    active={connActive}
                    to={`/conn/${conn.id}/query`}
                  />
                )}

                {/* Error state */}
                {error && (
                  <div className="px-4 py-2 text-[12px]" style={{ color: 'var(--danger)' }}>
                    {t('resource.error')}: {error}
                    <button
                      className="ml-2 underline"
                      onClick={() => fetchTopLevel(conn.id, true)}
                    >
                      {t('resource.retry')}
                    </button>
                  </div>
                )}

                {/* Empty state */}
                {!isLoading && !error && topNodes.length === 0 && (
                  <div className="px-4 py-2 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
                    {t('resource.empty')}
                  </div>
                )}

                {/* Resource tree nodes */}
                {topNodes.map(node => (
                  <ResourceNodeRow
                    key={node.id}
                    connId={conn.id}
                    node={node}
                    depth={1}
                    params={params}
                    isQueryPage={isQueryPage}
                    expandedNodes={expandedNodes}
                    resourceCache={resourceCache}
                    onToggle={toggleNode}
                    onLeafClick={handleNodeClick}
                    onContextMenu={handleContextMenu}
                    onRefresh={handleRefreshNode}
                    driver={conn.driver}
                  />
                ))}
              </>
            )}
          </div>
        )
      })}

      {/* Context menu */}
      {contextMenu && (
        <ContextMenu
          state={contextMenu}
          onClose={() => setContextMenu(null)}
        />
      )}
    </nav>
  )
}

// ── ResourceNodeRow ────────────────────────────────────────────────────────

function ResourceNodeRow({
  connId, node, depth, params, isQueryPage,
  expandedNodes, resourceCache, onToggle, onLeafClick, onContextMenu, onRefresh, driver,
}: {
  connId: string
  node: ResourceNode
  depth: number
  params: Record<string, string | undefined>
  isQueryPage: boolean
  expandedNodes: Record<string, boolean>
  resourceCache: Record<string, ResourceNode[]>
  onToggle: (connId: string, node: ResourceNode) => void
  onLeafClick: (connId: string, node: ResourceNode) => void
  onContextMenu: (e: React.MouseEvent, items: ContextMenuItem[]) => void
  onRefresh: (connId: string, node: ResourceNode) => void
  driver: string
}) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const nodeKey = `${connId}:${node.path}`
  const expanded = expandedNodes[nodeKey] ?? false
  const children: ResourceNode[] = resourceCache[nodeKey] ?? []

  const isActive = nodeIsActive(node, connId, params, isQueryPage)
  const expandable = node.type === 'sql_database' || node.type === 'mongo_database'

  const to = expandable ? undefined : (nodeUrl(connId, node) ?? undefined)

  // Build context menu items based on node type
  const handleContextMenu = (e: React.MouseEvent) => {
    const items: ContextMenuItem[] = []

    if (node.type === 'sql_database') {
      items.push({
        label: t('resource.open_console'),
        onClick: () => navigate(`/conn/${connId}/db/${encodeURIComponent(node.path)}/query`),
      })
    } else if (node.type === 'sql_table' || node.type === 'sql_view') {
      const slash = node.path.indexOf('/')
      if (slash >= 0) {
        const db = node.path.slice(0, slash)
        const table = node.path.slice(slash + 1)
        items.push({
          label: t('resource.view_data'),
          onClick: () => navigate(`/conn/${connId}/db/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/data`),
        })
        items.push({
          label: t('resource.view_structure'),
          onClick: () => navigate(`/conn/${connId}/db/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/structure`),
        })
      }
    } else if (node.type === 'mongo_collection') {
      const slash = node.path.indexOf('/')
      if (slash >= 0) {
        const db = node.path.slice(0, slash)
        const coll = node.path.slice(slash + 1)
        items.push({
          label: t('resource.open'),
          onClick: () => navigate(`/conn/${connId}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents`),
        })
      }
    } else if (node.type === 'es_index') {
      items.push({
        label: t('resource.open'),
        onClick: () => navigate(`/conn/${connId}/es/${encodeURIComponent(node.path)}`),
      })
    }

    if (items.length > 0) {
      items.push({
        label: t('resource.refresh'),
        onClick: () => onRefresh(connId, node),
      })
      onContextMenu(e, items)
    }
  }

  return (
    <>
      <TreeItem
        depth={depth}
        icon={nodeIcon(node.type)}
        label={node.name}
        active={isActive}
        expandable={expandable}
        expanded={expanded}
        to={to}
        onClick={expandable ? () => onToggle(connId, node) : () => onLeafClick(connId, node)}
        onContextMenu={handleContextMenu}
      />

      {expandable && expanded && (
        <>
          {/* Tables overview link — SQL databases only */}
          {node.type === 'sql_database' && (
            <TreeItem
              depth={depth + 1}
              icon={<span style={{ color: 'var(--sb-muted)', fontSize: 10 }}>⊞</span>}
              label="Tables"
              active={
                params.connId === connId &&
                params.dbName === node.path &&
                !params.tableName &&
                !isQueryPage
              }
              to={`/conn/${connId}/db/${encodeURIComponent(node.path)}/tables`}
            />
          )}
          {/* Collections overview link — MongoDB databases only */}
          {node.type === 'mongo_database' && (
            <TreeItem
              depth={depth + 1}
              icon={<span style={{ color: '#22d3ee', fontSize: 10 }}>⊞</span>}
              label="Collections"
              active={
                params.connId === connId &&
                params.mongoDb === node.path &&
                !params.mongoColl
              }
              to={`/conn/${connId}/mongo/${encodeURIComponent(node.path)}/collections`}
            />
          )}
          {children.map(child => (
            <ResourceNodeRow
              key={child.id}
              connId={connId}
              node={child}
              depth={depth + 1}
              params={params}
              isQueryPage={isQueryPage}
              expandedNodes={expandedNodes}
              resourceCache={resourceCache}
              onToggle={onToggle}
              onLeafClick={onLeafClick}
              onContextMenu={onContextMenu}
              onRefresh={onRefresh}
              driver={driver}
            />
          ))}
        </>
      )}
    </>
  )
}

function nodeIsActive(
  node: ResourceNode,
  connId: string,
  params: Record<string, string | undefined>,
  isQueryPage: boolean,
): boolean {
  if (params.connId !== connId) return false
  switch (node.type) {
    case 'sql_database':
      return params.dbName === node.path && !isQueryPage && !params.tableName
    case 'sql_table':
    case 'sql_view': {
      const slash = node.path.indexOf('/')
      if (slash < 0) return false
      const db = node.path.slice(0, slash)
      const table = node.path.slice(slash + 1)
      return params.dbName === db && params.tableName === table
    }
    case 'redis_db':
      return params.redisDb === node.path
    case 'mongo_database':
      return params.mongoDb === node.path && !params.mongoColl
    case 'mongo_collection': {
      const slash = node.path.indexOf('/')
      if (slash < 0) return false
      const db = node.path.slice(0, slash)
      const coll = node.path.slice(slash + 1)
      return params.mongoDb === db && params.mongoColl === coll
    }
    case 'es_index':
      return params.esIndex === node.path
    case 'es_alias':
      return params.esAlias === node.path
    case 'es_data_stream':
      return params.esDataStream === node.path
    default:
      return false
  }
}
