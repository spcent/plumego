import { client } from './client'

export interface AITask {
  id: string
  task_type: string
  status: string
  input_json: string
  output_document_id?: string
  output_json?: string
  provider?: string
  model?: string
  error_message?: string
  retry_count: number
  started_at?: string
  finished_at?: string
  created_at: string
  updated_at: string
}

export interface DocumentAISummary {
  id: string
  document_id: string
  summary: string
  key_points_json?: string
  actions_json?: string
  code_refs_json?: string
  provider?: string
  model?: string
  created_at: string
}

export interface AIPrompt {
  id: string
  title: string
  content: string
  source_document_id?: string
  model_hint?: string
  scenario?: string
  tags_json?: string
  quality_score: number
  created_at: string
  updated_at: string
}

export async function enqueueSummary(documentId: string): Promise<{ task: AITask }> {
  return client.post('/api/v1/ai/tasks/summary', { document_id: documentId })
}

export async function enqueueQA(question: string, documentIds: string[]): Promise<{ task: AITask }> {
  return client.post('/api/v1/ai/tasks/qa', { question, document_ids: documentIds })
}

export async function enqueuePromptExtract(documentId: string): Promise<{ task: AITask }> {
  return client.post('/api/v1/ai/tasks/prompt-extract', { document_id: documentId })
}

export async function listTasks(status?: string, limit = 20, offset = 0): Promise<{ items: AITask[]; total: number; limit: number; offset: number }> {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  return client.get(`/api/v1/ai/tasks?${params}`)
}

export async function getTask(id: string): Promise<{ task: AITask }> {
  return client.get(`/api/v1/ai/tasks/${id}`)
}

export async function cancelTask(id: string): Promise<{ ok: boolean }> {
  return client.post(`/api/v1/ai/tasks/${id}/cancel`, {})
}

export async function getDocumentSummary(documentId: string): Promise<{ summary: DocumentAISummary }> {
  return client.get(`/api/v1/ai/documents/${documentId}/summary`)
}

export async function listPrompts(scenario?: string, limit = 20, offset = 0): Promise<{ items: AIPrompt[]; total: number; limit: number; offset: number }> {
  const params = new URLSearchParams()
  if (scenario) params.set('scenario', scenario)
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  return client.get(`/api/v1/ai/prompts?${params}`)
}

export async function getPrompt(id: string): Promise<{ prompt: AIPrompt }> {
  return client.get(`/api/v1/ai/prompts/${id}`)
}

export async function deletePrompt(id: string): Promise<void> {
  return client.delete(`/api/v1/ai/prompts/${id}`)
}
