const BASE = '/api'

export class ApiError extends Error {
  readonly details?: Record<string, unknown>
  constructor(message: string, details?: Record<string, unknown>) {
    super(message)
    this.details = details
  }
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    credentials: 'include',
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const errBody = await res.json().catch(() => ({ error: { message: res.statusText } }))
    throw new ApiError(
      errBody?.error?.message || res.statusText,
      errBody?.error?.details,
    )
  }
  if (res.status === 204) return undefined as T
  const json = await res.json()
  return json.data !== undefined ? json.data : json
}

export const api = {
  login: (username: string, password: string) =>
    req<{ user: string }>('POST', '/auth/login', { username, password }),
  logout: () => req<void>('POST', '/auth/logout'),
  me: () => req<{ user: string }>('GET', '/auth/me'),

  listConnections: () => req<Connection[]>('GET', '/connections'),
  getConnection: (id: string) => req<Connection>('GET', `/connections/${id}`),
  createConnection: (c: Partial<Connection>) => req<Connection>('POST', '/connections', c),
  updateConnection: (id: string, c: Partial<Connection>) => req<Connection>('PUT', `/connections/${id}`, c),
  deleteConnection: (id: string, deleteFile = false) =>
    req<void>('DELETE', `/connections/${id}${deleteFile ? '?deleteFile=true' : ''}`),
  testConnection: (id: string) => req<{ ok: boolean; error?: string }>('POST', `/connections/${id}/test`),

  // Unified resource tree — works for all datasource types.
  // parentId is ResourceNode.path; omit to list top-level nodes (databases for SQL).
  resources: (id: string, parentId?: string) =>
    req<ResourceNode[]>('GET', `/connections/${id}/resources${parentId ? `?parentId=${encodeURIComponent(parentId)}` : ''}`),

  tableStructure: (id: string, db: string, table: string) =>
    req<TableStructure>('GET', `/conn/${id}/db/${db}/tables/${table}/structure`),

  listRows: (id: string, db: string, table: string, params: RowParams) => {
    const q = new URLSearchParams()
    if (params.page) q.set('page', String(params.page))
    if (params.pageSize) q.set('pageSize', String(params.pageSize))
    if (params.sortColumn) q.set('sortColumn', params.sortColumn)
    if (params.sortDirection) q.set('sortDirection', params.sortDirection)
    if (params.filters?.length) q.set('filters', JSON.stringify(params.filters))
    if (params.selectedColumns?.length) q.set('selectedColumns', params.selectedColumns.join(','))
    return req<RowsResponse>('GET', `/conn/${id}/db/${db}/tables/${table}/rows?${q}`)
  },
  createRow: (id: string, db: string, table: string, values: Record<string, unknown>) =>
    req<Record<string, unknown>>('POST', `/conn/${id}/db/${db}/tables/${table}/rows`, { values }),
  updateRow: (id: string, db: string, table: string, primaryKey: Record<string, unknown>, values: Record<string, unknown>) =>
    req<Record<string, unknown>>('PATCH', `/conn/${id}/db/${db}/tables/${table}/rows`, { primaryKey, values, confirm: true }),
  deleteRow: (id: string, db: string, table: string, primaryKey: Record<string, unknown>) =>
    req<void>('DELETE', `/conn/${id}/db/${db}/tables/${table}/rows`, { primaryKey, confirm: true }),

  executeQuery: (id: string, db: string, sql: string,
                opts?: { readonly?: boolean; confirmDangerous?: boolean }) =>
    req<QueryResult>('POST', `/conn/${id}/db/${db}/query`, { sql, database: db, ...opts }),
  listHistory: (id: string) => req<HistoryEntry[]>('GET', `/conn/${id}/history`),
  deleteHistory: (id: string, entryId: string) => req<void>('DELETE', `/conn/${id}/history/${entryId}`),
  clearHistory: (id: string) => req<void>('DELETE', `/conn/${id}/history`),

  createTable: (id: string, db: string, body: unknown) =>
    req<{ table: string }>('POST', `/conn/${id}/db/${db}/tables`, body),
  dropTable: (id: string, db: string, table: string) =>
    req<void>('DELETE', `/conn/${id}/db/${db}/tables/${table}?confirm=true`),

  schemaDoc: (id: string, db: string) =>
    req<{ markdown: string }>('GET', `/conn/${id}/db/${db}/schema-doc`),

  exportURL: (id: string, db: string, table: string, format: 'csv' | 'sql',
             opts?: { includeSchema?: boolean; includeData?: boolean }) => {
    const q = new URLSearchParams({ format })
    if (opts?.includeSchema === false) q.set('includeSchema', 'false')
    if (opts?.includeData === false) q.set('includeData', 'false')
    return `${BASE}/conn/${id}/db/${db}/tables/${table}/export?${q}`
  },

  importSQL: (id: string, db: string, sql: string, confirmDangerous = false) =>
    req<ImportResult>('POST', `/conn/${id}/db/${db}/import`, { sql, confirmDangerous }),

  uploadSQLite: (
    file: File,
    onProgress?: (pct: number) => void,
  ): Promise<{ file_path: string; size: number; original_name: string }> => {
    return new Promise((resolve, reject) => {
      const form = new FormData()
      form.append('file', file)
      const xhr = new XMLHttpRequest()
      xhr.open('POST', `${BASE}/sqlite/upload`)
      xhr.withCredentials = true
      if (onProgress) {
        xhr.upload.onprogress = (e) => {
          if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
        }
      }
      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const json = JSON.parse(xhr.responseText)
            resolve(json.data !== undefined ? json.data : json)
          } catch {
            reject(new ApiError('Invalid response'))
          }
        } else {
          try {
            const errBody = JSON.parse(xhr.responseText)
            reject(new ApiError(errBody?.error?.message || xhr.statusText, errBody?.error?.details))
          } catch {
            reject(new ApiError(xhr.statusText))
          }
        }
      }
      xhr.onerror = () => reject(new ApiError('Upload failed'))
      xhr.send(form)
    })
  },

  sqliteDownloadURL: (id: string) => `${BASE}/conn/${id}/sqlite/download`,

  // ── Redis API ──────────────────────────────────────────────────────────────
  redisListDBs: (id: string) => req<{ databases: RedisDB[] }>('GET', `/conn/${id}/redis/databases`),
  redisListKeys: (id: string, dbIndex: number, params?: RedisKeyParams) => {
    const q = new URLSearchParams()
    if (params?.pattern) q.set('pattern', params.pattern)
    if (params?.cursor != null) q.set('cursor', String(params.cursor))
    if (params?.count) q.set('count', String(params.count))
    return req<RedisKeysResponse>('GET', `/conn/${id}/redis/${dbIndex}/keys?${q}`)
  },
  redisGetKey: (id: string, dbIndex: number, key: string) =>
    req<RedisKeyDetail>('GET', `/conn/${id}/redis/${dbIndex}/key?key=${encodeURIComponent(key)}`),
  redisSetTTL: (id: string, dbIndex: number, key: string, ttl: number) =>
    req<{ ok: boolean }>('PATCH', `/conn/${id}/redis/${dbIndex}/key/ttl`, { key, ttl }),
  redisDeleteKey: (id: string, dbIndex: number, key: string) =>
    req<{ ok: boolean }>('DELETE', `/conn/${id}/redis/${dbIndex}/key`, { key, confirm: true }),
  redisCommand: (id: string, dbIndex: number, command: string) =>
    req<RedisCommandResult>('POST', `/conn/${id}/redis/${dbIndex}/command`, { command }),
  redisBatchPreview: (id: string, dbIndex: number, pattern: string, maxKeys: number) =>
    req<RedisBatchPreviewResponse>('POST', `/conn/${id}/redis/${dbIndex}/batch-preview`, {
      pattern,
      maxKeys,
    }),
  redisBatchDelete: (id: string, dbIndex: number, keys: string[], confirm: boolean) =>
    req<{ deleted: number }>('POST', `/conn/${id}/redis/${dbIndex}/batch-delete`, {
      keys,
      confirm,
    }),

  // ── MongoDB API (placeholder — P0 not yet implemented) ─────────────────────
  // These endpoints will be implemented in the MongoDB P0 phase.
  // Listed here for type completeness and to drive the placeholder UI.
  mongoListDatabases: (id: string) =>
    req<{ databases: MongoDatabaseInfo[] }>('GET', `/conn/${id}/mongo/databases`),
  mongoListCollections: (id: string, db: string) =>
    req<{ collections: MongoCollectionInfo[] }>('GET', `/conn/${id}/mongo/${encodeURIComponent(db)}/collections`),
  mongoListDocuments: (id: string, db: string, coll: string, params?: MongoDocParams) => {
    const q = new URLSearchParams()
    if (params?.page) q.set('page', String(params.page))
    if (params?.pageSize) q.set('pageSize', String(params.pageSize))
    if (params?.filter) q.set('filter', params.filter)
    return req<MongoDocsResponse>('GET', `/conn/${id}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents?${q}`)
  },
  mongoInsertDocument: (id: string, db: string, coll: string, document: Record<string, unknown>) =>
    req<{ inserted_id: string }>('POST', `/conn/${id}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents`, { document }),
  mongoUpdateDocument: (id: string, db: string, coll: string, filter: Record<string, unknown>, update: Record<string, unknown>) =>
    req<{ modified: number }>('PATCH', `/conn/${id}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents`, { filter, update, confirm: true }),
  mongoDeleteDocument: (id: string, db: string, coll: string, filter: Record<string, unknown>) =>
    req<{ deleted: number }>('DELETE', `/conn/${id}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/documents`, { filter, confirm: true }),
  mongoListIndexes: (id: string, db: string, coll: string) =>
    req<{ indexes: MongoIndexInfo[] }>('GET', `/conn/${id}/mongo/${encodeURIComponent(db)}/${encodeURIComponent(coll)}/indexes`),
}

export interface Connection {
  id: string
  name: string
  driver: 'mysql' | 'sqlite' | 'redis' | 'mongodb'
  host?: string
  port?: number
  database?: string
  username?: string
  password?: string
  file_path?: string
  options?: string
  // Redis-specific
  redis_db_index?: number  // 0-15, default 0
  // MongoDB-specific
  mongo_uri?: string           // mongodb://host:port or mongodb+srv://...
  mongo_auth_db?: string       // authentication database (default: admin)
  mongo_tls_enabled?: boolean  // use TLS/SSL
  mongo_replica_set?: string   // replica set name (optional)
  // Common flags
  tls_enabled?: boolean
  readonly?: boolean
  save_password?: boolean
  uploaded_file?: boolean
  original_filename?: string
}

export interface ColumnInfo {
  name: string
  position: number
  data_type: string
  full_type: string
  nullable: boolean
  default?: string
  primary_key?: boolean
  auto_increment?: boolean
}

export interface IndexInfo {
  name: string
  unique: boolean
  columns: string[]
  type?: string
}

export interface ForeignKeyInfo {
  name: string
  column: string
  ref_table: string
  ref_column: string
  on_delete?: string
  on_update?: string
}

export interface TableStructure {
  columns: ColumnInfo[]
  indexes: IndexInfo[]
  foreign_keys: ForeignKeyInfo[]
  ddl?: string
}

export type FilterOperator =
  'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'not_like' | 'is_null' | 'is_not_null'

export interface FilterCondition {
  column: string
  operator: FilterOperator
  value?: string
}

export interface RowParams {
  page?: number
  pageSize?: number
  sortColumn?: string
  sortDirection?: 'asc' | 'desc'
  filters?: FilterCondition[]
  selectedColumns?: string[]
}

export interface RowsResponse {
  rows: Record<string, unknown>[]
  total: number
  page: number
  pageSize: number
  columns: string[]
  executionTimeMs: number
}

export interface SelectResult {
  type: 'result_set'
  columns: string[]
  rows: Record<string, unknown>[]
  executionTimeMs: number
  truncated: boolean
}

export interface ExecResult {
  type: 'exec_result'
  rowsAffected: number
  lastInsertId: number
  executionTimeMs: number
}

export type QueryResult = SelectResult | ExecResult

export interface ImportErrorDetail { index: number; snippet: string; error: string }
export interface ImportResult {
  statements_executed: number
  errors: number
  errors_detail: ImportErrorDetail[]
}
export interface DangerousStatement { index: number; snippet: string; reason: string }

// ── Data Workbench resource model ──────────────────────────────────────────

export type DataSourceType =
  | 'mysql'
  | 'sqlite'
  | 'redis'         // config accepted; full driver in progress
  | 'mongodb'       // reserved — not yet implemented
  | 'elasticsearch' // reserved — not yet implemented

export type ResourceNodeType =
  // SQL types (available now)
  | 'sql_database'
  | 'sql_table'
  | 'sql_view'
  // Redis types (reserved)
  | 'redis_db'
  | 'redis_key'
  // MongoDB types (reserved)
  | 'mongo_database'
  | 'mongo_collection'
  // Elasticsearch types (reserved)
  | 'es_index'
  | 'es_alias'
  | 'es_data_stream'

export interface ResourceNode {
  id: string
  name: string
  type: ResourceNodeType
  parent_id?: string
  datasource_type: DataSourceType
  /** Slash-separated identifier: "mydb" for a database, "mydb/users" for a table. */
  path: string
  meta?: Record<string, unknown>
}

export interface HistoryEntry {
  id: string
  conn_id: string
  database: string
  sql: string
  duration_ms: number
  error?: string
  created_at: string
}

// ── Redis types ────────────────────────────────────────────────────────────

export interface RedisDB {
  index: number
  keys: number
}

export interface RedisKeyEntry {
  key: string
  type: string
  ttl: number // seconds; -1 = no expiry, -2 = not found
  memory?: number // bytes
  isBig?: boolean // true if > 1MB
}

export interface RedisKeyParams {
  pattern?: string
  cursor?: number
  count?: number
}

export interface RedisKeysResponse {
  keys: RedisKeyEntry[]
  nextCursor: number
  done: boolean
}

export interface RedisZSetMember {
  member: string
  score: number
}

export interface RedisStreamMessage {
  id: string
  values: Record<string, string>
}

export interface RedisStreamInfo {
  length: number
  groups: number
  messages?: RedisStreamMessage[]
}

export interface RedisKeyDetail {
  key: string
  type: string
  ttl: number
  encoding?: string
  string?: string
  hash?: Record<string, string>
  list?: string[]
  set?: string[]
  zset?: RedisZSetMember[]
  stream?: RedisStreamInfo
}

export interface RedisCommandResult {
  result: unknown
  error?: string
  timeMs: number
}

export interface RedisBatchPreviewResponse {
  keys: RedisKeyEntry[]
  totalKeys: number
  truncated: boolean
}

// ── MongoDB types (placeholder — P0 not yet implemented) ─────────────────

export interface MongoDatabaseInfo {
  name: string
  size_on_disk: number
  empty: boolean
}

export interface MongoCollectionInfo {
  name: string
  type: 'collection' | 'view'  // "collection" or "view"
}

export interface MongoDocParams {
  page?: number
  pageSize?: number
  filter?: string  // JSON query string, e.g. '{ "age": { "$gt": 30 } }'
}

export interface MongoDocsResponse {
  documents: Record<string, unknown>[]
  total: number
  page: number
  pageSize: number
}

export interface MongoIndexInfo {
  name: string
  unique: boolean
  keys: Record<string, number | string>  // e.g. { "field": 1 } or { "field": "text" }
  sparse?: boolean
}
