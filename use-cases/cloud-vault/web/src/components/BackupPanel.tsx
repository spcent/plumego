import { useState, useEffect } from 'react'
import { useI18n } from '../i18n/I18nContext'
import {
  createBackup,
  listBackups,
  deleteBackup,
  getBackupDownloadUrl,
  type Backup,
} from '../api/system'

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
    if (!window.confirm(t.backup.confirmDelete)) {
      return
    }
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
      const date = new Date(dateStr)
      return date.toLocaleString()
    } catch {
      return dateStr
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">{t.backup.title}</h3>
        <button
          onClick={handleCreate}
          disabled={creating}
          className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded hover:opacity-90 disabled:opacity-50"
        >
          {creating ? t.backup.creating : t.backup.create}
        </button>
      </div>

      {error && (
        <div className="text-sm text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      {loading && (
        <div className="text-sm text-muted-foreground">{t.common.loading}</div>
      )}

      {!loading && backups.length === 0 && (
        <div className="text-sm text-muted-foreground border border-border rounded-lg p-4 text-center">
          {t.backup.noBackups}
        </div>
      )}

      {!loading && backups.length > 0 && (
        <div className="space-y-2">
          {backups.map(backup => (
            <div
              key={backup.id}
              className="border border-border rounded-lg p-3 bg-background flex items-center justify-between gap-3"
            >
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate">{backup.filename}</div>
                <div className="text-xs text-muted-foreground mt-0.5 flex gap-3">
                  <span>{t.backup.size}: {formatSize(backup.size)}</span>
                  <span>{t.backup.createdAt}: {formatDate(backup.created_at)}</span>
                </div>
              </div>
              <div className="flex gap-2">
                <a
                  href={getBackupDownloadUrl(backup.filename)}
                  download
                  className="px-3 py-1.5 text-xs border border-border rounded hover:bg-accent"
                >
                  {t.backup.download}
                </a>
                <button
                  onClick={() => handleDelete(backup.filename)}
                  disabled={deletingId === backup.filename}
                  className="px-3 py-1.5 text-xs border border-border rounded hover:bg-destructive/10 disabled:opacity-50"
                >
                  {deletingId === backup.filename ? t.backup.deleting : t.backup.delete}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
