import { useState, useCallback, useEffect, useMemo, useRef } from 'react'
import { useNavigate, useParams, useLocation } from 'react-router-dom'
import { api, errorMessage, type Connection, type ResourceNode, type ResourceNodeType } from '../api'
import { useI18n } from '../i18nContext'
import { useToast } from './toastContext'
import {
  ChevronDownIcon,
  ChevronRightIcon,
  DatabaseIcon,
  ElasticsearchIcon,
  MongoDBIcon,
  MySQLIcon,
  RedisIcon,
  RefreshIcon,
  SQLiteIcon,
  TableIcon,
  XIcon,
} from './Icons'

const EXPANDED_CONNS_KEY = 'dbadmin.resourceExplorer.expandedConnections.v1'
const EXPANDED_NODES_KEY = 'dbadmin.resourceExplorer.expandedNodes.v1'

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

function readExpandedState(key: string): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}

    return Object.fromEntries(
      Object.entries(parsed)
        .filter((entry): entry is [string, boolean] => typeof entry[1] === 'boolean')
        .filter(([, expanded]) => expanded),
    )
  } catch {
    return {}
  }
}

function writeExpandedState(key: string, state: Record<string, boolean>) {
  try {
    const expandedOnly = Object.fromEntries(Object.entries(state).filter(([, expanded]) => expanded))
    localStorage.setItem(key, JSON.stringify(expandedOnly))
  } catch {
    // Best-effort preference persistence; the tree remains usable without storage.
  }
}

async function copyText(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

function textMatches(value: string, query: string): boolean {
  return value.toLowerCase().includes(query)
}

function connectionMatches(conn: Connection, query: string): boolean {
  return (
    textMatches(conn.name, query) ||
    textMatches(conn.driver, query) ||
    textMatches(conn.host ?? '', query) ||
    textMatches(conn.database ?? '', query)
  )
}

function resourceNodeMatches(node: ResourceNode, query: string): boolean {
  return (
    textMatches(node.name, query) ||
    textMatches(node.path, query) ||
    textMatches(node.type, query) ||
    textMatches(node.datasource_type, query)
  )
}

function resourceNodeHasMatch(
  connId: string,
  node: ResourceNode,
  resourceCache: Record<string, ResourceNode[]>,
  query: string,
  visited = new Set<string>(),
): boolean {
  if (resourceNodeMatches(node, query)) return true

  const key = `${connId}:${node.path}`
  if (visited.has(key)) return false
  visited.add(key)

  return (resourceCache[key] ?? []).some(child =>
    resourceNodeHasMatch(connId, child, resourceCache, query, visited),
  )
}

// ── Icon dispatch ──────────────────────────────────────────────────────────

function nodeIcon(type: ResourceNodeType): React.ReactNode {
  const iconClass = 'h-3.5 w-3.5'
  switch (type) {
    case 'sql_database':    return <DatabaseIcon className={iconClass} style={{ color: 'var(--sb-muted)' }} />
    case 'sql_table':       return <TableIcon className={iconClass} style={{ color: 'var(--sb-muted)' }} />
    case 'sql_view':        return <TableIcon className={iconClass} style={{ color: 'var(--sb-muted)', opacity: 0.72 }} />
    case 'redis_db':        return <RedisIcon className={iconClass} style={{ color: '#d94f4f' }} />
    case 'redis_key':       return <span className="h-1.5 w-1.5 rounded-full" style={{ background: '#fb923c' }} />
    case 'mongo_database':  return <MongoDBIcon className={iconClass} style={{ color: '#35a869' }} />
    case 'mongo_collection':return <TableIcon className={iconClass} style={{ color: '#35a869' }} />
    case 'es_index':        return <ElasticsearchIcon className={iconClass} style={{ color: '#d79a2b' }} />
    case 'es_alias':        return <ChevronRightIcon className={iconClass} style={{ color: '#9b7bd8' }} />
    case 'es_data_stream':  return <ElasticsearchIcon className={iconClass} style={{ color: '#2eaa91' }} />
    default:                return <span className="h-1.5 w-1.5 rounded-full" style={{ background: 'var(--sb-muted)' }} />
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
  const iconClass = 'h-3.5 w-3.5'
  const icon =
    driver === 'mysql'  ? <MySQLIcon className={iconClass} /> :
    driver === 'sqlite' ? <SQLiteIcon className={iconClass} /> :
    driver === 'redis'  ? <RedisIcon className={iconClass} /> :
    driver === 'mongodb'? <MongoDBIcon className={iconClass} /> :
    driver === 'elasticsearch' ? <ElasticsearchIcon className={iconClass} /> :
    <DatabaseIcon className={iconClass} />

  const color =
    driver === 'mysql'  ? '#2f7dbd' :
    driver === 'sqlite' ? '#4f86c6' :
    driver === 'redis'  ? '#d94f4f' :
    driver === 'mongodb'? '#35a869' :
    driver === 'elasticsearch' ? '#d79a2b' :
    'var(--sb-muted)'

  return (
    <span
      className="grid h-5 w-5 shrink-0 place-items-center rounded-md"
      style={{
        background: 'var(--sb-surface)',
        color,
        border: '1px solid var(--sb-border)',
      }}
      title={driver}
      aria-label={driver}
    >
      {icon}
    </span>
  )
}

function TreeItem({
  depth, icon, label, active, expanded, expandable, onClick, to, extra, actions, onContextMenu,
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
  actions?: React.ReactNode
  onContextMenu?: (e: React.MouseEvent) => void
}) {
  const navigate = useNavigate()
  const paddingLeft = 12 + depth * 16
  const cls = [
    'group/treeitem grid h-[30px] w-full items-center rounded-md mx-1 pr-1 cursor-pointer select-none truncate',
    'transition-colors duration-100',
    'hover:bg-[var(--sb-hover)] focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--focus-ring)]',
    active ? 'font-medium' : '',
  ].join(' ')
  const style: React.CSSProperties = {
    paddingLeft,
    width: 'calc(100% - 8px)',
    gridTemplateColumns: '14px 22px minmax(0, 1fr) auto',
    columnGap: 6,
    color: active ? '#fff' : 'var(--sb-text)',
    background: active ? 'var(--sb-active)' : 'transparent',
  }

  const activate = () => {
    if (onClick) {
      onClick()
      return
    }
    if (to) navigate(to)
  }

  const onKeyDown = (event: React.KeyboardEvent) => {
    if (event.key !== 'Enter' && event.key !== ' ') return
    event.preventDefault()
    activate()
  }

  return (
    <div
      role="button"
      tabIndex={0}
      aria-expanded={expandable ? expanded : undefined}
      className={cls}
      style={style}
      title={label}
      onClick={activate}
      onKeyDown={onKeyDown}
      onContextMenu={onContextMenu}
    >
      <span className="grid h-4 w-[14px] place-items-center" style={{ color: active ? 'rgba(255,255,255,0.76)' : 'var(--sb-muted)' }}>
        {expandable ? (
          expanded ? <ChevronDownIcon className="h-3.5 w-3.5" /> : <ChevronRightIcon className="h-3.5 w-3.5" />
        ) : null}
      </span>
      <span className="grid h-[22px] w-[22px] place-items-center">{icon}</span>
      <span className="min-w-0 truncate text-left">{label}</span>
      <span className="flex min-w-0 shrink-0 items-center justify-end gap-1">
        {extra}
        {actions && (
          <span className="ml-1 flex items-center gap-0.5 opacity-0 transition-opacity group-hover/treeitem:opacity-100 group-focus-within/treeitem:opacity-100">
            {actions}
          </span>
        )}
      </span>
    </div>
  )
}

function TreeActionButton({
  title,
  children,
  onClick,
}: {
  title: string
  children: React.ReactNode
  onClick: () => void
}) {
  return (
    <button
      type="button"
      title={title}
      aria-label={title}
      className="grid h-5 min-w-5 place-items-center rounded px-1 text-[10px] font-medium leading-none transition-colors hover:bg-[var(--sb-surface)] active:translate-y-px"
      style={{ color: 'var(--sb-muted)', border: '1px solid transparent' }}
      onClick={(event) => {
        event.stopPropagation()
        onClick()
      }}
      onContextMenu={(event) => event.stopPropagation()}
    >
      {children}
    </button>
  )
}

function ResourceSkeletonRows({ depth }: { depth: number }) {
  return (
    <div className="py-1">
      {[0, 1, 2].map(row => (
        <div key={row} className="flex h-[30px] items-center gap-2 px-3" style={{ paddingLeft: 12 + depth * 16 }}>
          <span className="skeleton h-4 w-4 rounded" style={{ background: 'var(--sb-surface)' }} />
          <span
            className="skeleton h-3 rounded"
            style={{ width: `${68 - row * 12}%`, background: 'var(--sb-surface)' }}
          />
        </div>
      ))}
    </div>
  )
}

// ── ResourceExplorer ───────────────────────────────────────────────────────

interface Props {
  connections: Connection[]
  onRefresh: () => void
}

export default function ResourceExplorer({ connections, onRefresh }: Props) {
  const params = useParams<{ connId?: string; dbName?: string; tableName?: string; redisDb?: string; mongoDb?: string; mongoColl?: string; esIndex?: string; esAlias?: string; esDataStream?: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const { t } = useI18n()
  const { showToast } = useToast()

  // Expanded state: connection IDs and node paths
  const [expandedConns, setExpandedConns] = useState<Record<string, boolean>>(() => readExpandedState(EXPANDED_CONNS_KEY))
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>(() => readExpandedState(EXPANDED_NODES_KEY))

  // Cached resource nodes: key = connId (top-level) or "connId:dbPath" (children)
  const [resourceCache, setResourceCache] = useState<Record<string, ResourceNode[]>>({})

  // Loading and error states
  const [loadingConns, setLoadingConns] = useState<Set<string>>(new Set())
  const [connErrors, setConnErrors] = useState<Record<string, string>>({})
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null)
  const [filterText, setFilterText] = useState('')

  const isQueryPage = location.pathname.includes('/query')
  const filterQuery = filterText.trim().toLowerCase()

  useEffect(() => {
    writeExpandedState(EXPANDED_CONNS_KEY, expandedConns)
  }, [expandedConns])

  useEffect(() => {
    writeExpandedState(EXPANDED_NODES_KEY, expandedNodes)
  }, [expandedNodes])

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
      const msg = errorMessage(err, 'Failed to load resources')
      setConnErrors(prev => ({ ...prev, [connId]: msg }))
    } finally {
      setLoadingConns(prev => {
        const next = new Set(prev)
        next.delete(connId)
        return next
      })
    }
  }, [resourceCache])

  useEffect(() => {
    connections.forEach(conn => {
      if (expandedConns[conn.id]) void fetchTopLevel(conn.id)
    })
  }, [connections, expandedConns, fetchTopLevel])

  // Fetch child nodes (tables/views) under a database node
  const fetchChildren = useCallback(async (connId: string, node: ResourceNode, forceRefresh = false) => {
    const key = `${connId}:${node.path}`
    if (!forceRefresh && resourceCache[key] !== undefined) return
    try {
      const nodes = await api.resources(connId, node.id)
      setResourceCache(c => ({ ...c, [key]: nodes }))
    } catch {
      // Leave this branch unloaded; expanding or refreshing can retry.
    }
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

  const handleCopyResource = useCallback(async (value: string) => {
    try {
      await copyText(value)
      showToast(t('resource.copy_success'), 'success')
    } catch {
      showToast(t('data.copy_failed'), 'error')
    }
  }, [showToast, t])

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

  const filteredConnections = useMemo(() => {
    if (!filterQuery) return connections
    return connections.filter(conn => {
      if (connectionMatches(conn, filterQuery)) return true
      return (resourceCache[conn.id] ?? []).some(node =>
        resourceNodeHasMatch(conn.id, node, resourceCache, filterQuery),
      )
    })
  }, [connections, filterQuery, resourceCache])

  return (
    <nav className="flex-1 overflow-y-auto py-1 overflow-x-hidden" onClick={() => setContextMenu(null)}>
      <div className="flex items-center justify-between px-3 py-2">
        <div className="text-[11px] font-semibold uppercase tracking-[0.14em]" style={{ color: 'var(--sb-muted)' }}>
          Connections
        </div>
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation()
            onRefresh()
          }}
          className="grid h-6 w-6 place-items-center rounded-md transition-colors hover:bg-white/10"
          style={{ color: 'var(--sb-muted)' }}
          title={t('Refresh')}
          aria-label={t('Refresh')}
        >
          <RefreshIcon className="h-3.5 w-3.5" />
        </button>
      </div>
      <div className="px-3 pb-2">
        <div className="relative">
          <input
            value={filterText}
            onChange={(event) => setFilterText(event.target.value)}
            onClick={(event) => event.stopPropagation()}
            placeholder={t('resource.search_placeholder')}
            className="h-7 w-full rounded-md border px-2.5 pr-7 text-[12px] outline-none transition-colors"
            style={{
              background: 'var(--sb-surface)',
              borderColor: 'var(--sb-border)',
              color: 'var(--sb-text)',
            }}
          />
          {filterText && (
            <button
              type="button"
              onClick={(event) => {
                event.stopPropagation()
                setFilterText('')
              }}
              className="absolute right-1 top-1 grid h-5 w-5 place-items-center rounded transition-colors hover:bg-[var(--sb-hover)]"
              style={{ color: 'var(--sb-muted)' }}
              title={t('resource.clear_search')}
              aria-label={t('resource.clear_search')}
            >
              <XIcon className="h-3 w-3" />
            </button>
          )}
        </div>
      </div>
      {connections.length === 0 && (
        <div className="px-4 py-3 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
          {t('connections.empty')}
        </div>
      )}

      {connections.length > 0 && filteredConnections.length === 0 && (
        <div className="px-4 py-3 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
          {t('resource.no_matches')}
        </div>
      )}

      {filteredConnections.map(conn => {
        const connExpanded = expandedConns[conn.id]
        const connActive = params.connId === conn.id && isQueryPage
        const topNodes: ResourceNode[] = resourceCache[conn.id] ?? []
        const isLoading = loadingConns.has(conn.id)
        const error = connErrors[conn.id]
        const connSelfMatches = filterQuery ? connectionMatches(conn, filterQuery) : true
        const visibleTopNodes = filterQuery
          ? topNodes.filter(node => resourceNodeHasMatch(conn.id, node, resourceCache, filterQuery))
          : topNodes
        const showConnBody = connExpanded || Boolean(filterQuery)

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
                  ...(conn.driver === 'mysql' || conn.driver === 'sqlite'
                    ? [{ label: t('resource.open_console'), onClick: () => navigate(`/conn/${conn.id}/query`) }]
                    : []),
                  { label: t('resource.copy_name'), onClick: () => void handleCopyResource(conn.name) },
                  { label: t('resource.refresh'), onClick: () => handleRefreshConn(conn.id) },
                  { label: t('resource.manage_connection'), onClick: () => navigate('/connections') },
                ])
              }}
              extra={
                <>
                  {isLoading && (
                    <RefreshIcon className="h-3 w-3 shrink-0 animate-spin" style={{ color: 'var(--sb-muted)' }} />
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
              actions={
                <>
                  {(conn.driver === 'mysql' || conn.driver === 'sqlite') && (
                    <TreeActionButton title={t('resource.open_console')} onClick={() => navigate(`/conn/${conn.id}/query`)}>
                      SQL
                    </TreeActionButton>
                  )}
                  <TreeActionButton title={t('resource.refresh')} onClick={() => handleRefreshConn(conn.id)}>
                    <RefreshIcon className="h-3 w-3" />
                  </TreeActionButton>
                </>
              }
            />

            {showConnBody && (
              <>
                {/* SQL Console shortcut — SQL drivers only */}
                {(conn.driver === 'mysql' || conn.driver === 'sqlite') && (!filterQuery || connSelfMatches || textMatches(t('nav.sql_console'), filterQuery)) && (
                  <TreeItem
                    depth={1}
                    icon={<ChevronRightIcon className="h-3.5 w-3.5" style={{ color: 'var(--sb-muted)' }} />}
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

                {isLoading && topNodes.length === 0 && (
                  <ResourceSkeletonRows depth={1} />
                )}

                {/* Empty state */}
                {!filterQuery && !isLoading && !error && topNodes.length === 0 && (
                  <div className="px-4 py-2 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
                    {t('resource.empty')}
                  </div>
                )}

                {/* Resource tree nodes */}
                {filterQuery && !isLoading && !error && topNodes.length > 0 && visibleTopNodes.length === 0 && connSelfMatches && (
                  <div className="px-4 py-2 text-[12px]" style={{ color: 'var(--sb-muted)' }}>
                    {t('resource.no_matches')}
                  </div>
                )}

                {visibleTopNodes.map(node => (
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
                    onCopy={handleCopyResource}
                    filterQuery={filterQuery}
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
  expandedNodes, resourceCache, onToggle, onLeafClick, onContextMenu, onRefresh, onCopy, filterQuery,
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
  onCopy: (value: string) => void
  filterQuery: string
}) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const nodeKey = `${connId}:${node.path}`
  const expanded = expandedNodes[nodeKey] ?? false
  const children: ResourceNode[] = resourceCache[nodeKey] ?? []
  const hasFilter = Boolean(filterQuery)

  const isActive = nodeIsActive(node, connId, params, isQueryPage)
  const expandable = node.type === 'sql_database' || node.type === 'mongo_database'
  const selfMatches = !hasFilter || resourceNodeMatches(node, filterQuery)
  const visibleChildren = hasFilter && !selfMatches
    ? children.filter(child => resourceNodeHasMatch(connId, child, resourceCache, filterQuery))
    : children
  const to = expandable ? undefined : (nodeUrl(connId, node) ?? undefined)

  if (hasFilter && !selfMatches && visibleChildren.length === 0) {
    return null
  }

  const openNode = () => {
    const url = nodeUrl(connId, node)
    if (url) navigate(url)
  }

  const actionButtons = (
    <>
      {node.type === 'sql_database' && (
        <TreeActionButton title={t('resource.open_console')} onClick={() => navigate(`/conn/${connId}/query`)}>
          SQL
        </TreeActionButton>
      )}
      {(node.type === 'sql_table' || node.type === 'sql_view') && (
        <>
          <TreeActionButton title={t('resource.view_data')} onClick={openNode}>
            Data
          </TreeActionButton>
          <TreeActionButton
            title={t('resource.view_structure')}
            onClick={() => {
              const slash = node.path.indexOf('/')
              if (slash < 0) return
              const db = node.path.slice(0, slash)
              const table = node.path.slice(slash + 1)
              navigate(`/conn/${connId}/db/${encodeURIComponent(db)}/tables/${encodeURIComponent(table)}/structure`)
            }}
          >
            DDL
          </TreeActionButton>
        </>
      )}
      {(node.type === 'mongo_database' ||
        node.type === 'mongo_collection' ||
        node.type === 'redis_db' ||
        node.type === 'redis_key' ||
        node.type === 'es_index' ||
        node.type === 'es_alias' ||
        node.type === 'es_data_stream') && (
        <TreeActionButton title={t('resource.open')} onClick={openNode}>
          Open
        </TreeActionButton>
      )}
      {expandable && (
        <TreeActionButton title={t('resource.refresh')} onClick={() => onRefresh(connId, node)}>
          <RefreshIcon className="h-3 w-3" />
        </TreeActionButton>
      )}
      <TreeActionButton title={t('resource.copy_path')} onClick={() => onCopy(node.path || node.name)}>
        Copy
      </TreeActionButton>
    </>
  )

  // Build context menu items based on node type
  const handleContextMenu = (e: React.MouseEvent) => {
    const items: ContextMenuItem[] = []

    if (node.type === 'sql_database') {
      items.push({
        label: t('resource.open'),
        onClick: openNode,
      })
      items.push({
        label: t('resource.open_console'),
        onClick: () => navigate(`/conn/${connId}/query`),
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
    } else if (nodeUrl(connId, node)) {
      items.push({
        label: t('resource.open'),
        onClick: openNode,
      })
    }

    items.push({
      label: t('resource.copy_path'),
      onClick: () => onCopy(node.path || node.name),
    })

    if (expandable) {
      items.push({
        label: t('resource.refresh'),
        onClick: () => onRefresh(connId, node),
      })
    }

    if (items.length > 0) {
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
        actions={actionButtons}
      />

      {expandable && (expanded || hasFilter) && (
        <>
          {/* Tables overview link — SQL databases only */}
          {node.type === 'sql_database' && (!hasFilter || selfMatches || textMatches(t('resource.tables'), filterQuery)) && (
            <TreeItem
              depth={depth + 1}
              icon={<TableIcon className="h-3.5 w-3.5" style={{ color: 'var(--sb-muted)' }} />}
              label={t('resource.tables')}
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
          {node.type === 'mongo_database' && (!hasFilter || selfMatches || textMatches(t('resource.collections'), filterQuery)) && (
            <TreeItem
              depth={depth + 1}
              icon={<TableIcon className="h-3.5 w-3.5" style={{ color: '#22d3ee' }} />}
              label={t('resource.collections')}
              active={
                params.connId === connId &&
                params.mongoDb === node.path &&
                !params.mongoColl
              }
              to={`/conn/${connId}/mongo/${encodeURIComponent(node.path)}/collections`}
            />
          )}
          {visibleChildren.map(child => (
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
              onCopy={onCopy}
              filterQuery={filterQuery}
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
