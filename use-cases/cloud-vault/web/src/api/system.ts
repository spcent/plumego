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
  'config_check',
  'data_dir_check',
  'storage_writable_check',
  'auth_security_check',
  'cookie_security_check',
  'qiniu_config_check',
  'backup_check',
  'migration_check',
]

// V1.0: Version and update interfaces
export interface VersionInfo {
  version: string
  commit: string
  build_time: string
  channel: string
}

export interface UpdateStatus {
  current_version: VersionInfo
  latest_release?: {
    version: string
    commit: string
    build_time: string
    channel: string
    release_date: string
    release_url: string
    download_url: string
    changelog: string
  }
  update_available: boolean
  last_check?: string
  next_check?: string
  check_enabled: boolean
}

export interface DiagnosticBundle {
  id: string
  filename: string
  size: number
  created_at: string
  download_url: string
}

export interface Backup {
  id: string
  filename: string
  size: number
  created_at: string
  storage_path: string
}

export interface SystemSettings {
  version: string
  storage_provider: string
  auth_enabled: boolean
  search_enabled: boolean
  ai_enabled: boolean
  database_path: string
  storage_root?: string
}

export async function getSettings(): Promise<SystemSettings> {
  return client.get('/api/v1/system/settings')
}

export async function createBackup(): Promise<{ backup: Backup }> {
  return client.post('/api/v1/system/backup', {})
}

export async function listBackups(): Promise<{ backups: Backup[] }> {
  return client.get('/api/v1/system/backups')
}

export async function deleteBackup(name: string): Promise<void> {
  return client.delete(`/api/v1/system/backups/${encodeURIComponent(name)}`)
}

export function getBackupDownloadUrl(name: string): string {
  return `/api/v1/system/backups/${encodeURIComponent(name)}/download`
}

// V1.0: Version, update, and diagnostics API methods
export async function getVersion(): Promise<VersionInfo> {
  return client.get('/api/v1/system/version')
}

export async function getUpdateStatus(): Promise<UpdateStatus> {
  return client.get('/api/v1/system/update/status')
}

export async function checkForUpdates(): Promise<UpdateStatus> {
  return client.post('/api/v1/system/update/check', {})
}

export async function generateDiagnostics(): Promise<DiagnosticBundle> {
  return client.post('/api/v1/system/diagnostics/export', {})
}

export async function listDiagnostics(): Promise<{ bundles: DiagnosticBundle[] }> {
  return client.get('/api/v1/system/diagnostics')
}

export function getDiagnosticDownloadUrl(filename: string): string {
  return `/api/v1/system/diagnostics/${encodeURIComponent(filename)}/download`
}
