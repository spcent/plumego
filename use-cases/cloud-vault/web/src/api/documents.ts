import { client } from './client'

export interface DocumentSummary {
  id: string
  title: string
  version: number
  size_bytes: number
  word_count: number
  line_count: number
  is_favorite: boolean
  source_type: string
  review_status: string
  summary?: string
  updated_at: string
}

export interface DocumentDetail extends DocumentSummary {
  content: string
  status?: string
  import_job_id?: string
  original_path?: string
  heading_text?: string
  created_at: string
  imported_at?: string
}

export interface DocumentListResult {
  items: DocumentSummary[]
  total: number
  limit: number
  offset: number
}

export interface SaveResult {
  id: string
  title: string
  version: number
  changed: boolean
  updated_at: string
}

export interface VersionSummary {
  version: number
  size_bytes: number
  created_at: string
}

export interface VersionListResult {
  items: VersionSummary[]
}

export interface VersionDetail {
  id: string
  version: number
  content: string
  created_at: string
}

export interface CreateDocumentRequest {
  title: string
  content: string
}

export interface UpdateDocumentRequest {
  title: string
  content: string
  base_version: number
}

export interface ListDocumentsParams {
  q?: string
  tag_id?: string
  status?: string        // active | archived | all
  source_type?: string   // manual | imported
  import_job_id?: string
  is_favorite?: boolean
  review_status?: string // pending | reviewed
  sort_by?: string       // updated_at | created_at | title | size_bytes
  order?: string         // asc | desc
  limit?: number
  offset?: number
}

const BASE = '/api/v1'

export const documentsAPI = {
  list(params?: ListDocumentsParams): Promise<DocumentListResult> {
    const qs = new URLSearchParams()
    if (params?.q) qs.set('q', params.q)
    if (params?.tag_id) qs.set('tag_id', params.tag_id)
    if (params?.status) qs.set('status', params.status)
    if (params?.source_type) qs.set('source_type', params.source_type)
    if (params?.import_job_id) qs.set('import_job_id', params.import_job_id)
    if (params?.is_favorite != null) qs.set('is_favorite', params.is_favorite ? '1' : '0')
    if (params?.review_status) qs.set('review_status', params.review_status)
    if (params?.sort_by) qs.set('sort_by', params.sort_by)
    if (params?.order) qs.set('order', params.order)
    if (params?.limit != null) qs.set('limit', String(params.limit))
    if (params?.offset != null) qs.set('offset', String(params.offset))
    const query = qs.toString()
    return client.get<DocumentListResult>(`${BASE}/documents${query ? '?' + query : ''}`)
  },

  get(id: string): Promise<DocumentDetail> {
    return client.get<DocumentDetail>(`${BASE}/documents/${id}`)
  },

  create(req: CreateDocumentRequest): Promise<SaveResult> {
    return client.post<SaveResult>(`${BASE}/documents`, req)
  },

  update(id: string, req: UpdateDocumentRequest): Promise<SaveResult> {
    return client.put<SaveResult>(`${BASE}/documents/${id}`, req)
  },

  delete(id: string): Promise<void> {
    return client.delete(`${BASE}/documents/${id}`)
  },

  updateFavorite(id: string, isFavorite: boolean): Promise<void> {
    return client.put<void>(`${BASE}/documents/${id}/favorite`, { is_favorite: isFavorite })
  },

  updateStatus(id: string, status: string): Promise<void> {
    return client.put<void>(`${BASE}/documents/${id}/status`, { status })
  },

  updateReviewStatus(id: string, reviewStatus: string): Promise<void> {
    return client.put<void>(`${BASE}/documents/${id}/review-status`, { review_status: reviewStatus })
  },

  batchUpdateStatus(ids: string[], status: string): Promise<void> {
    return client.put<void>(`${BASE}/documents/batch-status`, { ids, status })
  },

  listVersions(id: string): Promise<VersionListResult> {
    return client.get<VersionListResult>(`${BASE}/documents/${id}/versions`)
  },

  getVersion(id: string, version: number): Promise<VersionDetail> {
    return client.get<VersionDetail>(`${BASE}/documents/${id}/versions/${version}`)
  },
}
