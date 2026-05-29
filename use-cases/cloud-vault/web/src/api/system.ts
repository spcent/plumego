import { client } from './client'

export interface HealthResult {
  status: string
  database: string
  storage: string
  search: string
  ai: string
}

export interface StatsResult {
  documents: number
  versions: number
  collections: number
  tags: number
  import_jobs: number
  indexed_documents: number
  pending_indexes: number
  failed_indexes: number
  ai_tasks: number
  prompts: number
}

export interface IssueItem {
  document_id?: string
  entity_id?: string
  issue: string
  detail?: string
}

export interface CheckResult {
  name: string
  status: string
  total: number
  failed: number
  items: IssueItem[]
}

export interface DoctorResult {
  status: string
  checks: CheckResult[]
}

export async function getHealth(): Promise<HealthResult> {
  return client.get('/api/v1/system/health')
}

export async function getStats(): Promise<StatsResult> {
  return client.get('/api/v1/system/stats')
}

export async function runDoctor(checks?: string[], sampleSize?: number): Promise<DoctorResult> {
  const body: { checks?: string[]; sample_size?: number } = {}
  if (checks && checks.length > 0) {
    body.checks = checks
  }
  if (sampleSize && sampleSize > 0) {
    body.sample_size = sampleSize
  }
  return client.post('/api/v1/system/doctor', body)
}

export const ALL_CHECKS = [
  'storage_objects',
  'document_versions',
  'document_hash',
  'fts_index',
  'tags',
  'collections',
  'sources',
  'imports',
  'ai',
]
