import { client } from './client'

export interface ImportJob {
  id: string
  name: string
  source_path: string
  status: string // pending | running | paused | done | failed | cancelled
  total_count: number
  processed_count: number
  success_count: number
  failed_count: number
  skipped_count: number
  error_message?: string
  started_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface ImportJobListResult {
  items: ImportJob[]
  total: number
  limit: number
  offset: number
}

export interface ImportJobItem {
  id: string
  job_id: string
  file_path: string
  document_id?: string
  status: string // pending | success | skipped | failed
  error_message?: string
  created_at: string
  updated_at: string
}

export interface ImportJobItemListResult {
  items: ImportJobItem[]
  total: number
  limit: number
  offset: number
}

const BASE = '/api/v1'

export const importsAPI = {
  list(limit = 20, offset = 0): Promise<ImportJobListResult> {
    return client.get<ImportJobListResult>(`${BASE}/imports?limit=${limit}&offset=${offset}`)
  },

  get(id: string): Promise<ImportJob> {
    return client.get<ImportJob>(`${BASE}/imports/${id}`)
  },

  create(name: string, sourcePath: string): Promise<ImportJob> {
    return client.post<ImportJob>(`${BASE}/imports`, { name, source_path: sourcePath })
  },

  start(id: string): Promise<void> {
    return client.post<void>(`${BASE}/imports/${id}/start`, {})
  },

  pause(id: string): Promise<void> {
    return client.post<void>(`${BASE}/imports/${id}/pause`, {})
  },

  cancel(id: string): Promise<void> {
    return client.post<void>(`${BASE}/imports/${id}/cancel`, {})
  },

  retry(id: string): Promise<void> {
    return client.post<void>(`${BASE}/imports/${id}/retry`, {})
  },

  listItems(
    id: string,
    params?: { status?: string; limit?: number; offset?: number },
  ): Promise<ImportJobItemListResult> {
    const qs = new URLSearchParams()
    if (params?.status) qs.set('status', params.status)
    if (params?.limit != null) qs.set('limit', String(params.limit))
    if (params?.offset != null) qs.set('offset', String(params.offset))
    const query = qs.toString()
    return client.get<ImportJobItemListResult>(
      `${BASE}/imports/${id}/items${query ? '?' + query : ''}`,
    )
  },
}
