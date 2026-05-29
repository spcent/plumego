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

  databases: (id: string) => req<string[]>('GET', `/conn/${id}/databases`),
  tables: (id: string, db: string) => req<TableInfo[]>('GET', `/conn/${id}/db/${db}/tables`),
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

  exportURL: (id: string, db: string, table: string, format: 'csv' | 'sql',
             opts?: { includeSchema?: boolean; includeData?: boolean }) => {
    const q = new URLSearchParams({ format })
    if (opts?.includeSchema === false) q.set('includeSchema', 'false')
    if (opts?.includeData === false) q.set('includeData', 'false')
    return `${BASE}/conn/${id}/db/${db}/tables/${table}/export?${q}`
  },

  importSQL: (id: string, db: string, sql: string, confirmDangerous = false) =>
    req<ImportResult>('POST', `/conn/${id}/db/${db}/import`, { sql, confirmDangerous }),

  uploadSQLite: async (file: File): Promise<{ file_path: string; size: number; original_name: string }> => {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${BASE}/sqlite/upload`, {
      method: 'POST',
      credentials: 'include',
      body: form,
    })
    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: { message: res.statusText } }))
      throw new ApiError(errBody?.error?.message || res.statusText, errBody?.error?.details)
    }
    const json = await res.json()
    return json.data !== undefined ? json.data : json
  },

  sqliteDownloadURL: (id: string) => `${BASE}/conn/${id}/sqlite/download`,
}

export interface Connection {
  id: string
  name: string
  driver: 'mysql' | 'sqlite'
  host?: string
  port?: number
  database?: string
  username?: string
  password?: string
  file_path?: string
  options?: string
  readonly?: boolean
  save_password?: boolean
  uploaded_file?: boolean
  original_filename?: string
}

export interface TableInfo {
  name: string
  type: string
  comment?: string
  engine?: string
  rows?: number
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

export interface HistoryEntry {
  id: string
  conn_id: string
  database: string
  sql: string
  duration_ms: number
  error?: string
  created_at: string
}
