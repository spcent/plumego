import type { Connection } from '../api'

export const RECENT_RESOURCES_KEY = 'dbadmin.recentResources.v1'

export interface RecentEntry {
  path: string
  lastVisited: number
}

export interface RecentResourceDescriptor {
  path: string
  label: string
  detail: string
  driver?: Connection['driver']
  activeKey: string
}

type Translate = (key: string, vars?: Record<string, string | number>) => string

const listeners = new Set<() => void>()
let cachedEntries: RecentEntry[] | null = null

function safeDecode(value: string): string {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

function detail(parts: string[]): string {
  return parts.filter(Boolean).join(' · ')
}

function storage(): Storage | null {
  if (typeof localStorage === 'undefined') return null
  if (typeof localStorage.getItem !== 'function') return null
  if (typeof localStorage.setItem !== 'function') return null
  return localStorage
}

function readRecentEntries(): RecentEntry[] {
  try {
    const store = storage()
    const raw = store?.getItem(RECENT_RESOURCES_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return []

    return parsed.filter((entry): entry is RecentEntry => {
      return (
        Boolean(entry) &&
        typeof entry === 'object' &&
        typeof (entry as RecentEntry).path === 'string' &&
        typeof (entry as RecentEntry).lastVisited === 'number'
      )
    })
  } catch {
    return []
  }
}

function writeRecentEntries(entries: RecentEntry[]) {
  try {
    storage()?.setItem(RECENT_RESOURCES_KEY, JSON.stringify(entries))
  } catch {
    // Recent shortcuts are a convenience feature; ignore unavailable storage.
  }
}

function emitRecentChange() {
  listeners.forEach(listener => listener())
}

function setRecentEntries(entries: RecentEntry[]) {
  cachedEntries = entries
  writeRecentEntries(entries)
  emitRecentChange()
}

export function subscribeRecentResources(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function getRecentEntriesSnapshot(): RecentEntry[] {
  if (!cachedEntries) cachedEntries = readRecentEntries()
  return cachedEntries
}

export function reloadRecentResourcesFromStorage() {
  cachedEntries = readRecentEntries()
  emitRecentChange()
}

export function trackRecentResource(path: string) {
  const prev = getRecentEntriesSnapshot()
  if (prev[0]?.path === path) return

  setRecentEntries([
    { path, lastVisited: Date.now() },
    ...prev.filter(entry => entry.path !== path),
  ].slice(0, 12))
}

export function clearRecentResources() {
  cachedEntries = []
  try {
    storage()?.removeItem(RECENT_RESOURCES_KEY)
  } catch {
    writeRecentEntries([])
  }
  emitRecentChange()
}

export function buildRecentResourceDescriptor(
  path: string,
  connections: Connection[],
  t: Translate,
): RecentResourceDescriptor | null {
  const url = new URL(path, 'http://dbadmin.local')
  const parts = url.pathname.split('/').filter(Boolean).map(safeDecode)

  if (parts[0] !== 'conn' || !parts[1]) return null

  const connId = parts[1]
  const conn = connections.find(c => c.id === connId)
  const connName = conn?.name ?? connId
  const driver = conn?.driver

  if (parts[2] === 'query') {
    return {
      path,
      label: t('query.title'),
      detail: connName,
      driver,
      activeKey: `${connId}:query`,
    }
  }

  if (parts[2] === 'db' && parts[3]) {
    const db = parts[3]
    const table = parts[5]
    const view = parts[6]

    if (parts[4] === 'tables' && table && view === 'data') {
      return {
        path,
        label: table,
        detail: detail([db, t('resource.view_data'), connName]),
        driver,
        activeKey: `${connId}:table:${db}:${table}:data`,
      }
    }

    if (parts[4] === 'tables' && table && view === 'structure') {
      return {
        path,
        label: table,
        detail: detail([db, t('resource.view_structure'), connName]),
        driver,
        activeKey: `${connId}:table:${db}:${table}:structure`,
      }
    }

    if (parts[4] === 'tables') {
      return {
        path,
        label: db,
        detail: detail([t('resource.tables'), connName]),
        driver,
        activeKey: `${connId}:db:${db}`,
      }
    }
  }

  if (parts[2] === 'redis' && parts[3]) {
    const redisDb = parts[3]
    const key = url.searchParams.get('key')
    return {
      path,
      label: key ? safeDecode(key) : t('redis.browser.db', { n: redisDb }),
      detail: detail([key ? t('recent.redis_key') : 'Redis', connName]),
      driver,
      activeKey: `${connId}:redis:${redisDb}:${key ?? ''}`,
    }
  }

  if (parts[2] === 'mongo' && parts[3]) {
    const db = parts[3]
    const collection = parts[4]

    if (collection && parts[5] === 'documents') {
      return {
        path,
        label: collection,
        detail: detail([db, t('mongodb.browser.title'), connName]),
        driver,
        activeKey: `${connId}:mongo:${db}:${collection}`,
      }
    }

    if (parts[4] === 'collections') {
      return {
        path,
        label: db,
        detail: detail([t('resource.collections'), connName]),
        driver,
        activeKey: `${connId}:mongo:${db}`,
      }
    }
  }

  if (parts[2] === 'es' && parts[3]) {
    const kind = parts[3] === 'alias' || parts[3] === 'data-stream' ? parts[3] : 'index'
    const resource = kind === 'index' ? parts[3] : parts[4]
    if (!resource) return null

    const kindLabel =
      kind === 'alias' ? t('recent.es_alias') :
      kind === 'data-stream' ? t('recent.es_data_stream') :
      t('elasticsearch.index_name')

    return {
      path,
      label: resource,
      detail: detail([kindLabel, connName]),
      driver,
      activeKey: `${connId}:es:${kind}:${resource}`,
    }
  }

  return null
}
