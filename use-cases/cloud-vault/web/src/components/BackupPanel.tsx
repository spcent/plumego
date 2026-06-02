import { useEffect, useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import {
  createBackup,
  deleteBackup,
  getBackupDownloadUrl,
  listBackups,
  type Backup,
} from '../api/system'
import { Button, EmptyState, Panel, SkeletonRows, StatusBanner } from './ui'

export default function BackupPanel() {
  const { t } = useI18n()
  const [backups, setBackups] = useState<Backup[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    loadBackups()
  }, [])

  async function loadBackups() {
    setLoading(true)
    setError('')
    try {
      const result = await listBackups()
      setBackups(result.backups || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load backups')
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate() {
    setCreating(true)
    setError('')
    try {
      await createBackup()
      await loadBackups()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create backup')
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(name: string) {
    if (!window.confirm(t.backup.confirmDelete)) return
    setDeletingId(name)
    setError('')
    try {
      await deleteBackup(name)
      await loadBackups()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete backup')
    } finally {
      setDeletingId(null)
    }
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  function formatDate(dateStr: string): string {
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  return (
    <Panel
      title={t.backup.title}
      description="Create local archive snapshots before maintenance or recovery work."
      action={<Button size="sm" variant="primary" icon="archive" onClick={handleCreate} disabled={creating}>{creating ? t.backup.creating : t.backup.create}</Button>}
    >
      <div className="p-4">
        {error && <StatusBanner tone="danger">{error}</StatusBanner>}
        {loading && <SkeletonRows count={3} />}
        {!loading && backups.length === 0 && (
          <EmptyState compact icon="archive" title={t.backup.noBackups} description="Create a backup to keep a restorable snapshot." />
        )}
        {!loading && backups.length > 0 && (
          <div className="divide-y divide-border overflow-hidden rounded-lg border border-border">
            {backups.map(backup => (
              <div key={backup.id} className="flex items-center justify-between gap-3 bg-background/45 px-3 py-3">
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium text-foreground">{backup.filename}</div>
                  <div className="mt-0.5 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
                    <span>{t.backup.size}: {formatSize(backup.size)}</span>
                    <span>{t.backup.createdAt}: {formatDate(backup.created_at)}</span>
                  </div>
                </div>
                <div className="flex shrink-0 gap-2">
                  <a
                    href={getBackupDownloadUrl(backup.filename)}
                    download
                    className="inline-flex h-8 items-center justify-center rounded-md border border-border bg-surface px-2.5 text-xs font-medium text-foreground transition-colors hover:bg-accent"
                  >
                    {t.backup.download}
                  </a>
                  <Button
                    size="sm"
                    variant="danger"
                    onClick={() => handleDelete(backup.filename)}
                    disabled={deletingId === backup.filename}
                  >
                    {deletingId === backup.filename ? t.backup.deleting : t.backup.delete}
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Panel>
  )
}
