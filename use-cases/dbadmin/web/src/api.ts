const BASE = '/api'

export interface ApiErrorOptions {
  status?: number
  code?: string
  category?: string
  type?: string
  details?: Record<string, unknown>
  requestId?: string
  retryable?: boolean
}

export class ApiError extends Error {
  readonly details?: Record<string, unknown>
  readonly status?: number
  readonly code?: string
  readonly category?: string
  readonly type?: string
  readonly requestId?: string
  readonly retryable: boolean

  constructor(message: string, options: ApiErrorOptions = {}) {
    super(redactSensitive(message))
    this.name = 'ApiError'
    this.details = sanitizeDetails(options.details)
    this.status = options.status
    this.code = options.code
    this.category = options.category
    this.type = options.type
    this.requestId = options.requestId
    this.retryable = options.retryable ?? isRetryableStatus(options.status, options.category)
  }
}

export interface ApiErrorInfo {
  title: string
  message: string
  detail?: string
  status?: number
  code?: string
  requestId?: string
  retryable: boolean
}

function isRetryableStatus(status?: number, category?: string): boolean {
  if (category === 'timeout_error' || category === 'rate_limit_error') return true
  if (!status) return false
  return status === 408 || status === 409 || status === 429 || status >= 500
}

export function redactSensitive(input: unknown): string {
  const raw = input == null ? '' : String(input)
  return raw
    .replace(/:\/\/([^:/\s]+):([^@\s]+)@/g, '://***:***@')
    .replace(/\b(password|passwd|pwd|token|api[_-]?key|secret|authorization)=([^&\s]+)/gi, '$1=***')
    .replace(/\b(Bearer|Basic)\s+[A-Za-z0-9._~+/=-]+/gi, '$1 ***')
    .slice(0, 800)
}

function sanitizeDetails(details?: Record<string, unknown>): Record<string, unknown> | undefined {
  if (!details) return undefined
  const clean: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(details)) {
    if (/password|passwd|pwd|token|api[_-]?key|secret|authorization/i.test(key)) {
      clean[key] = '***'
    } else if (typeof value === 'string') {
      clean[key] = redactSensitive(value)
    } else {
      clean[key] = value
    }
  }
  return clean
}

function parseErrorBody(body: unknown, status: number, statusText: string): ApiError {
  const envelope = body as {
    error?: {
      message?: unknown
      details?: Record<string, unknown>
      code?: unknown
      category?: unknown
      type?: unknown
    }
    request_id?: unknown
  }
  const err = envelope?.error ?? {}
  return new ApiError(
    typeof err.message === 'string' && err.message ? err.message : statusText,
    {
      status,
      code: typeof err.code === 'string' ? err.code : undefined,
      category: typeof err.category === 'string' ? err.category : undefined,
      type: typeof err.type === 'string' ? err.type : undefined,
      details: err.details,
      requestId: typeof envelope?.request_id === 'string' ? envelope.request_id : undefined,
    },
  )
}

function networkError(err: unknown): ApiError {
  return new ApiError(err instanceof Error && err.message ? err.message : 'Network request failed', {
    code: 'NETWORK_ERROR',
    category: 'network_error',
    retryable: true,
  })
}

export function getApiErrorInfo(err: unknown, fallback = 'Operation failed'): ApiErrorInfo {
  if (err instanceof ApiError) {
    let title = 'Operation failed'
    if (err.category === 'network_error') title = 'Network error'
    else if (err.status === 401) title = 'Authentication required'
    else if (err.status === 403) title = 'Permission denied'
    else if (err.status === 404) title = 'Not found'
    else if (err.status === 408 || err.category === 'timeout_error') title = 'Request timed out'
    else if (err.status === 413) title = 'Payload too large'
    else if (err.status === 429) title = 'Too many requests'
    else if (err.status && err.status >= 500) title = 'Service error'
    else if (err.category === 'validation_error' || err.type === 'bad_request') title = 'Check request'

    const detailParts = [
      err.code ? `code: ${err.code}` : '',
      err.requestId ? `request: ${err.requestId}` : '',
    ].filter(Boolean)
    return {
      title,
      message: err.message || fallback,
      detail: detailParts.length ? detailParts.join(' · ') : undefined,
      status: err.status,
      code: err.code,
      requestId: err.requestId,
      retryable: err.retryable,
    }
  }
  if (err instanceof Error) {
    return { title: 'Operation failed', message: redactSensitive(err.message || fallback), retryable: false }
  }
  return { title: 'Operation failed', message: fallback, retryable: false }
}

export function errorMessage(err: unknown, fallback = 'Operation failed'): string {
  const info = getApiErrorInfo(err, fallback)
  return info.detail ? `${info.message} (${info.detail})` : info.message
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE}${path}`, {
      method,
      credentials: 'include',
      headers: body ? { 'Content-Type': 'application/json' } : {},
      body: body ? JSON.stringify(body) : undefined,
    })
  } catch (err) {
    throw networkError(err)
  }
  if (!res.ok) {
    const errBody = await res.json().catch(() => ({ error: { message: res.statusText } }))
    throw parseErrorBody(errBody, res.status, res.statusText)
  }
  if (res.status === 204) return undefined as T
  const json = await res.json()
  return json.data !== undefined ? json.data : json
}

export const api = {
  login: (username: string, password: string) =>
    req<{ user: string; role: string }>('POST', '/auth/login', { username, password }),
  logout: () => req<void>('POST', '/auth/logout'),
  me: () => req<{ user: string; role: string }>('GET', '/auth/me'),
  listAuditEvents: () => req<AuditEvent[]>('GET', '/audit/events'),

  listConnections: () => req<Connection[]>('GET', '/connections'),
  getConnection: (id: string) => req<Connection>('GET', `/connections/${id}`),
  createConnection: (c: Partial<Connection>) => req<Connection>('POST', '/connections', c),
  updateConnection: (id: string, c: Partial<Connection>) => req<Connection>('PUT', `/connections/${id}`, c),
  deleteConnection: (id: string, deleteFile = false) =>
    req<void>('DELETE', `/connections/${id}${deleteFile ? '?deleteFile=true' : ''}`),
  testConnection: (id: string) => req<{ ok: boolean; error?: string }>('POST', `/connections/${id}/test`),
  closeConnectionRuntime: (id: string) => req<void>('DELETE', `/connections/${id}/runtime`),
  poolStats: () => req<PoolStats>('GET', '/pool-stats'),

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
                opts?: { readonly?: boolean; confirmDangerous?: boolean; queryId?: string }) =>
    req<QueryResult>('POST', `/conn/${id}/db/${db}/query`, { sql, database: db, ...opts }),
  cancelQuery: (queryId: string) =>
    req<{ status: string; queryId: string }>('POST', '/queries/cancel', { queryId }),
  listActiveQueries: () =>
    req<ActiveQuery[]>('GET', '/queries/active'),
  listActiveOperations: () =>
    req<ActiveOperation[]>('GET', '/operations/active'),
  cancelOperation: (operationId: string) =>
    req<{ status: string; operationId: string }>('POST', '/operations/cancel', { operationId }),
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
             opts?: { includeSchema?: boolean; includeData?: boolean; limit?: number }) => {
    const q = new URLSearchParams({ format })
    if (opts?.includeSchema === false) q.set('includeSchema', 'false')
    if (opts?.includeData === false) q.set('includeData', 'false')
    if (opts?.limit) q.set('limit', String(opts.limit))
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
            reject(new ApiError('Invalid response', { code: 'INVALID_RESPONSE' }))
          }
        } else {
          try {
            const errBody = JSON.parse(xhr.responseText)
            reject(parseErrorBody(errBody, xhr.status, xhr.statusText))
          } catch {
            reject(new ApiError(xhr.statusText, { status: xhr.status }))
          }
        }
      }
      xhr.onerror = () => reject(new ApiError('Upload failed', { code: 'NETWORK_ERROR', category: 'network_error', retryable: true }))
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
  redisListHistory: (id: string) =>
    req<RedisHistoryEntry[]>('GET', `/conn/${id}/redis/history`),
  redisDeleteHistoryEntry: (id: string, entryId: string) =>
    req<void>('DELETE', `/conn/${id}/redis/history/${encodeURIComponent(entryId)}`),
  redisClearHistory: (id: string) =>
    req<void>('DELETE', `/conn/${id}/redis/history`),
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

  // ── MongoDB API ─────────────────────────────────────────────────────────────
  mongoListDatabases: (connId: string) =>
    req<{ databases: MongoDatabaseInfo[] }>('GET', `/connections/${connId}/mongo/databases`),

  mongoListCollections: (connId: string, database: string) =>
    req<{ collections: MongoCollectionInfo[] }>('GET', `/connections/${connId}/mongo/collections?database=${encodeURIComponent(database)}`),

  mongoQueryDocuments: (connId: string, query: MongoDocQuery) =>
    req<MongoDocsResponse>('POST', `/connections/${connId}/mongo/documents/query`, query),

  mongoInsertDocument: (connId: string, database: string, collection: string, document: string) =>
    req<{ inserted_id: string }>('POST', `/connections/${connId}/mongo/documents`, { database, collection, document }),

  mongoUpdateDocument: (connId: string, database: string, collection: string, id: string, document: string) =>
    req<{ modified: number }>('PATCH', `/connections/${connId}/mongo/documents`, { database, collection, id, document }),

  mongoDeleteDocument: (connId: string, database: string, collection: string, id: string) =>
    req<{ deleted: number }>('DELETE', `/connections/${connId}/mongo/documents`, { database, collection, id, confirm: true }),

  mongoListIndexes: (connId: string, database: string, collection: string) =>
    req<{ indexes: MongoIndexInfo[] }>('GET', `/connections/${connId}/mongo/indexes?database=${encodeURIComponent(database)}&collection=${encodeURIComponent(collection)}`),

  // MongoDB P1 - Advanced features
  mongoAggregate: (connId: string, database: string, collection: string, pipeline: string) =>
    req<MongoAggregateResponse>('POST', `/connections/${connId}/mongo/aggregate`, { database, collection, pipeline }),

  mongoExplain: (connId: string, database: string, collection: string, filter?: string) =>
    req<MongoExplainResponse>('POST', `/connections/${connId}/mongo/explain`, { database, collection, filter }),

  mongoSchema: (connId: string, database: string, collection: string, sampleSize = 100) =>
    req<MongoSchemaResponse>('GET', `/connections/${connId}/mongo/schema?database=${encodeURIComponent(database)}&collection=${encodeURIComponent(collection)}&sample=${sampleSize}`),

  mongoStats: (connId: string, database: string, collection: string) =>
    req<MongoStatsResponse>('GET', `/connections/${connId}/mongo/stats?database=${encodeURIComponent(database)}&collection=${encodeURIComponent(collection)}`),

  mongoExport: (connId: string, database: string, collection: string, format: 'json' | 'ndjson' | 'csv' = 'json', filter?: string, limit?: number) => {
    const params = new URLSearchParams({ database, collection, format })
    if (filter) params.append('filter', filter)
    if (limit) params.set('limit', String(limit))
    return `${BASE}/connections/${connId}/mongo/export?${params.toString()}`
  },

  mongoImport: (connId: string, database: string, collection: string, data: string, format: 'json' | 'ndjson' = 'json') =>
    req<MongoImportResponse>('POST', `/connections/${connId}/mongo/import`, { database, collection, data, format }),

  mongoParseObjectId: (objectId: string) =>
    req<MongoObjectIdInfo>('GET', `/mongo/objectid/${objectId}/parse`),

  // MongoDB P1 - Pipeline history
  mongoListHistory: (connId: string) =>
    req<MongoPipelineEntry[]>('GET', `/connections/${connId}/mongo/history`),

  mongoDeleteHistoryEntry: (connId: string, entryId: string) =>
    req<void>('DELETE', `/connections/${connId}/mongo/history/${entryId}`),

  mongoClearHistory: (connId: string) =>
    req<void>('DELETE', `/connections/${connId}/mongo/history`),

  // ── Elasticsearch API ─────────────────────────────────────────────────────
  esInfo: (connId: string) =>
    req<ESClusterInfo>('GET', `/connections/${connId}/es/info`),

  esListIndices: (connId: string) =>
    req<{ indices: ESIndexInfo[] }>('GET', `/connections/${connId}/es/indices`),

  esGetMapping: (connId: string, index: string) =>
    req<Record<string, unknown>>('GET', `/connections/${connId}/es/index/mapping?index=${encodeURIComponent(index)}`),

  esGetSettings: (connId: string, index: string) =>
    req<Record<string, unknown>>('GET', `/connections/${connId}/es/index/settings?index=${encodeURIComponent(index)}`),

  esSearch: (connId: string, index: string, dsl: Record<string, unknown>) =>
    req<ESSearchResponse>('POST', `/connections/${connId}/es/search`, { index, dsl }),

  esGetDocument: (connId: string, index: string, id: string) =>
    req<ESDocument>('GET', `/connections/${connId}/es/document?index=${encodeURIComponent(index)}&id=${encodeURIComponent(id)}`),

  esDeleteDocument: (connId: string, index: string, id: string, confirm: boolean) =>
    req<void>('DELETE', `/connections/${connId}/es/document`, { index, id, confirm }),

  esExport: (connId: string, index: string, format: 'json' | 'ndjson' = 'json', query?: Record<string, unknown>, limit?: number) => {
    const params = new URLSearchParams({ index, format })
    if (query) params.set('query', JSON.stringify(query))
    if (limit) params.set('limit', String(limit))
    return `${BASE}/connections/${connId}/es/export?${params.toString()}`
  },

  esImport: (connId: string, index: string, documents: Record<string, unknown>[], confirm: boolean) =>
    req<ESImportResponse>('POST', `/connections/${connId}/es/import`, { index, documents, confirm }),

  esListHistory: (connId: string) =>
    req<ESHistoryEntry[]>('GET', `/connections/${connId}/es/history`),

  esDeleteHistoryEntry: (connId: string, entryId: string) =>
    req<void>('DELETE', `/connections/${connId}/es/history/${encodeURIComponent(entryId)}`),

  esClearHistory: (connId: string) =>
    req<void>('DELETE', `/connections/${connId}/es/history`),
}

export interface Connection {
  id: string
  name: string
  driver: 'mysql' | 'sqlite' | 'redis' | 'mongodb' | 'elasticsearch'
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
  // Elasticsearch-specific
  es_nodes?: string[]            // ["http://host:9200", ...]
  es_username?: string           // basic auth username
  es_password?: string           // basic auth password
  es_api_key?: string            // API key (alternative to basic auth)
  es_ca_cert?: string            // CA cert path (optional)
  es_insecure_skip_tls?: boolean // skip TLS verification
  // Common flags
  tls_enabled?: boolean
  readonly?: boolean
  save_password?: boolean
  uploaded_file?: boolean
  original_filename?: string
}

export interface SQLPoolStats {
  connection_id: string
  driver: string
  max_open: number
  open: number
  in_use: number
  idle: number
  wait_count: number
  wait_duration: string
  max_idle_closed: number
  max_lifetime_closed: number
}

export interface PoolStats {
  sql_connections?: SQLPoolStats[]
  redis_connections: number
  mongodb_connections: number
  es_connections: number
  timestamp: string
}

export interface AuditEvent {
  id: string
  user: string
  action: string
  method: string
  path: string
  status: number
  created_at: string
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
  queryId?: string
  columns: string[]
  rows: Record<string, unknown>[]
  executionTimeMs: number
  truncated: boolean
}

export interface ExecResult {
  type: 'exec_result'
  queryId?: string
  rowsAffected: number
  lastInsertId: number
  executionTimeMs: number
}

export type QueryResult = SelectResult | ExecResult

export interface ActiveQuery {
  queryId: string
  connId: string
  database: string
  sql: string
  startTime: string
  duration: string
}

export interface ActiveOperation {
  operationId: string
  driver: string
  kind: string
  connId: string
  resource: string
  summary: string
  startTime: string
  duration: string
}

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

export interface RedisHistoryEntry {
  id: string
  conn_id: string
  db_index: number
  command: string
  duration_ms: number
  success: boolean
  error?: string
  created_at: string
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

export interface MongoDocQuery {
  database: string
  collection: string
  filter?: string
  projection?: string
  sort?: string
  limit?: number
  skip?: number
}

export interface MongoDocsResponse {
  documents: Record<string, unknown>[]
  total: number
  limit: number
  skip: number
}

export interface MongoIndexInfo {
  v: number
  key: Record<string, unknown>
  name: string
  unique?: boolean
  sparse?: boolean
  [key: string]: unknown
}

// MongoDB P1 - Advanced types
export interface MongoAggregateResponse {
  documents: Record<string, unknown>[]
  count: number
  duration_ms: number
}

export interface MongoExplainResponse {
  explain: Record<string, unknown>
}

export interface MongoSchemaResponse {
  fields: MongoSchemaField[]
  sampled_count: number
}

export interface MongoSchemaField {
  name: string
  types: Array<{ type: string; count: number }>
  sample_values: unknown[]
}

export interface MongoStatsResponse {
  count: number
  size: number
  avgObjSize: number
  storageSize: number
  totalIndexSize: number
  indexes: Array<{ name: string; size: number }>
}

export interface MongoImportResponse {
  inserted_count: number
}

export interface MongoObjectIdInfo {
  hex: string
  timestamp: string
  counter: number
}

export interface MongoPipelineEntry {
  id: string
  conn_id: string
  database: string
  collection: string
  pipeline: string
  duration_ms: number
  result_count: number
  error?: string
  created_at: string
}

// ── Elasticsearch Types ─────────────────────────────────────────────────────

export interface ESClusterInfo {
  cluster_name: string
  cluster_uuid: string
  version: {
    number: string
    build_flavor: string
    lucene_version: string
  }
  tagline: string
}

export interface ESIndexInfo {
  name: string
  health: 'green' | 'yellow' | 'red'
  status: 'open' | 'close'
  docs_count: string
  store_size: string
  pri: string
  rep: string
}

export interface ESSearchResponse {
  took: number
  timed_out: boolean
  total: { value: number; relation: string }
  hits: Array<{
    _index: string
    _id: string
    _score: number | null
    _source: Record<string, unknown>
  }>
  aggregations?: Record<string, unknown> | null
}

export interface ESDocument {
  _index: string
  _id: string
  _version: number
  _source: Record<string, unknown>
}

export interface ESImportResponse {
  imported_count: number
  errors: number
}

export interface ESHistoryEntry {
  id: string
  conn_id: string
  index: string
  dsl: string
  duration_ms: number
  result_count: number
  error?: string
  created_at: string
}
