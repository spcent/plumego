import { client } from './client'

export interface DuplicateDoc {
  id: string
  title: string
  status: string
  is_favorite: boolean
  created_at: string
  updated_at: string
  import_job_id: string
  original_path: string
}

export interface DuplicateGroup {
  content_hash: string
  documents: DuplicateDoc[]
}

export interface SimilarDoc {
  similarity_id: string
  document_id: string
  title: string
  similarity_type: string
  score: number
  status: string
}

export interface TagSuggestion {
  id: string
  document_id: string
  tag_id: string
  tag_name: string
  source: string
  confidence: number
  status: string
  created_at: string
}

export interface Topic {
  id: string
  name: string
  description: string
  source: string
  status: string
  doc_count: number
  created_at: string
}

export interface OrganizeJob {
  id: string
  job_type: string
  status: string
  total_items: number
  processed_items: number
  failed_items: number
  error_message: string
  started_at: string | null
  finished_at: string | null
  created_at: string
}

export interface ReviewItem {
  type: string
  document_id: string
  title: string
  score: number
  extra?: Record<string, string>
}

export interface ReviewQueueResult {
  items: ReviewItem[]
  total: number
  limit: number
  offset: number
}

export const detectDuplicates = () => client.post<OrganizeJob>('/api/v1/organize/detect-duplicates', {})

export const listDuplicates = () =>
  client.get<{ groups: DuplicateGroup[]; total: number }>('/api/v1/organize/duplicates')

export const resolveDuplicates = (req: {
  keep_document_id: string
  duplicate_document_ids: string[]
  action: string
}) => client.post<void>('/api/v1/organize/duplicates/resolve', req)

export const detectSimilarity = () => client.post<OrganizeJob>('/api/v1/organize/detect-similarity', {})

export const getSimilarDocuments = (docId: string) =>
  client.get<{ items: SimilarDoc[] }>(`/api/v1/documents/${docId}/similar`)

export const ignoreSimilarity = (id: string) =>
  client.post<void>(`/api/v1/organize/similarity/${id}/ignore`, {})

export const confirmSimilarity = (id: string) =>
  client.post<void>(`/api/v1/organize/similarity/${id}/confirm`, {})

export const suggestTags = () => client.post<OrganizeJob>('/api/v1/organize/suggest-tags', {})

export const getTagSuggestions = (docId: string) =>
  client.get<{ items: TagSuggestion[] }>(`/api/v1/documents/${docId}/tag-suggestions`)

export const acceptTagSuggestion = (id: string) =>
  client.post<void>(`/api/v1/tag-suggestions/${id}/accept`, {})

export const rejectTagSuggestion = (id: string) =>
  client.post<void>(`/api/v1/tag-suggestions/${id}/reject`, {})

export const batchAcceptTagSuggestions = (ids: string[]) =>
  client.post<{ accepted: number }>('/api/v1/tag-suggestions/batch/accept', { ids })

export const buildTopics = () => client.post<OrganizeJob>('/api/v1/organize/build-topics', {})

export const listTopics = () => client.get<{ items: Topic[]; total: number }>('/api/v1/topics')

export const getTopic = (id: string) =>
  client.get<{ topic: Topic; documents: DuplicateDoc[]; total: number }>(`/api/v1/topics/${id}`)

export const getReviewQueue = (params?: { type?: string; limit?: number; offset?: number }) => {
  const q = new URLSearchParams()
  if (params?.type) q.set('type', params.type)
  if (params?.limit) q.set('limit', String(params.limit))
  if (params?.offset) q.set('offset', String(params.offset))
  return client.get<ReviewQueueResult>(`/api/v1/review/queue?${q}`)
}

export const runAll = () => client.post<{ status: string }>('/api/v1/organize/run-all', {})

export const scoreQuality = () => client.post<OrganizeJob>('/api/v1/organize/score-quality', {})

export const detectPromptCandidates = () =>
  client.post<OrganizeJob>('/api/v1/organize/detect-prompt-candidates', {})

export const listOrganizeJobs = () => client.get<{ items: OrganizeJob[] }>('/api/v1/organize/jobs')
