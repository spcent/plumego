import { client } from './client'

export interface SearchResultItem {
  id: string
  title: string
  summary?: string
  highlights?: string[]
  score: number
  tags?: string[]
  original_path?: string
  source_type: string
  is_favorite: boolean
  updated_at: string
}

export interface SearchResult {
  items: SearchResultItem[]
  total: number
  limit: number
  offset: number
}

export interface SearchParams {
  q?: string
  tag?: string
  status?: string
  review_status?: string
  source_type?: string
  import_job_id?: string
  is_favorite?: boolean
  from?: string
  to?: string
  sort?: string
  order?: string
  limit?: number
  offset?: number
}

export interface IndexStatus {
  total_documents: number
  indexed: number
  pending: number
  failed: number
  stale: number
  last_indexed_at?: string
}

export interface ReindexRequest {
  scope: string
  document_id?: string
}

export interface HistoryItem {
  id: string
  query: string
  result_count: number
  created_at: string
}

export interface HistoryResponse {
  items: HistoryItem[]
}

const BASE = '/api/v1'

export const searchAPI = {
  search(params: SearchParams): Promise<SearchResult> {
    const qs = new URLSearchParams()
    if (params.q) qs.set('q', params.q)
    if (params.tag) qs.set('tag', params.tag)
    if (params.status) qs.set('status', params.status)
    if (params.review_status) qs.set('review_status', params.review_status)
    if (params.source_type) qs.set('source_type', params.source_type)
    if (params.import_job_id) qs.set('import_job_id', params.import_job_id)
    if (params.is_favorite != null) qs.set('is_favorite', params.is_favorite ? '1' : '0')
    if (params.from) qs.set('from', params.from)
    if (params.to) qs.set('to', params.to)
    if (params.sort) qs.set('sort', params.sort)
    if (params.order) qs.set('order', params.order)
    if (params.limit != null) qs.set('limit', String(params.limit))
    if (params.offset != null) qs.set('offset', String(params.offset))
    const query = qs.toString()
    return client.get<SearchResult>(`${BASE}/search${query ? '?' + query : ''}`)
  },

  getIndexStatus(): Promise<IndexStatus> {
    return client.get<IndexStatus>(`${BASE}/search/index-status`)
  },

  reindex(req: ReindexRequest): Promise<{ status: string }> {
    return client.post<{ status: string }>(`${BASE}/search/reindex`, req)
  },

  getHistory(): Promise<HistoryResponse> {
    return client.get<HistoryResponse>(`${BASE}/search/history`)
  },

  clearHistory(): Promise<void> {
    return client.delete(`${BASE}/search/history`)
  },
}
