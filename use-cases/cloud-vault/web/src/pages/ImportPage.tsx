import { useEffect, useState } from 'react'
import { importsAPI, type ImportJob, type ImportJobItem } from '../api/imports'

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending:   'bg-gray-100 text-gray-600',
    running:   'bg-blue-100 text-blue-700',
    paused:    'bg-amber-100 text-amber-700',
    done:      'bg-green-100 text-green-700',
    failed:    'bg-red-100 text-red-600',
    cancelled: 'bg-gray-100 text-gray-500',
    success:   'bg-green-100 text-green-700',
    skipped:   'bg-gray-100 text-gray-500',
  }
  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${colors[status] ?? 'bg-gray-100 text-gray-600'}`}>
      {status}
    </span>
  )
}

function ProgressBar({ total, processed }: { total: number; processed: number }) {
  const pct = total > 0 ? Math.round((processed / total) * 100) : 0
  return (
    <div className="w-full bg-gray-100 rounded-full h-1.5 overflow-hidden">
      <div
        className="h-full bg-primary transition-all"
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}

export default function ImportPage() {
  const [jobs, setJobs] = useState<ImportJob[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedJob, setSelectedJob] = useState<ImportJob | null>(null)
  const [items, setItems] = useState<ImportJobItem[]>([])
  const [itemsLoading, setItemsLoading] = useState(false)
  const [itemFilter, setItemFilter] = useState('')

  // Create form
  const [showForm, setShowForm] = useState(false)
  const [formName, setFormName] = useState('')
  const [formPath, setFormPath] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  useEffect(() => {
    loadJobs()
  }, [])

  // Poll running jobs.
  useEffect(() => {
    const running = jobs.some(j => j.status === 'running')
    if (!running) return
    const timer = setInterval(() => loadJobs(), 2000)
    return () => clearInterval(timer)
  }, [jobs])

  async function loadJobs() {
    try {
      const result = await importsAPI.list(50)
      setJobs(result.items)
      // Refresh selected job if still active.
      if (selectedJob) {
        const updated = result.items.find(j => j.id === selectedJob.id)
        if (updated) setSelectedJob(updated)
      }
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  async function loadItems(jobId: string, status = '') {
    setItemsLoading(true)
    try {
      const result = await importsAPI.listItems(jobId, { status, limit: 100 })
      setItems(result.items)
    } catch {
      setItems([])
    } finally {
      setItemsLoading(false)
    }
  }

  async function handleSelectJob(job: ImportJob) {
    setSelectedJob(job)
    setItemFilter('')
    await loadItems(job.id)
  }

  async function handleAction(action: 'start' | 'pause' | 'cancel' | 'retry', jobId: string) {
    try {
      await importsAPI[action](jobId)
      await loadJobs()
      if (selectedJob?.id === jobId) {
        const updated = await importsAPI.get(jobId)
        setSelectedJob(updated)
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed')
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!formPath.trim()) return
    setCreating(true)
    setCreateError('')
    try {
      const job = await importsAPI.create(formName || formPath, formPath.trim())
      setJobs(prev => [job, ...prev])
      setFormName('')
      setFormPath('')
      setShowForm(false)
      await handleSelectJob(job)
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create import job')
    } finally {
      setCreating(false)
    }
  }

  async function handleItemFilterChange(filter: string) {
    setItemFilter(filter)
    if (selectedJob) await loadItems(selectedJob.id, filter)
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' })
  }

  return (
    <div className="h-full flex flex-col overflow-hidden bg-background">
      {/* Header */}
      <header className="h-12 flex items-center px-4 border-b border-border gap-3 shrink-0">
        <span className="text-sm font-semibold text-foreground">Import Jobs</span>
        <div className="flex-1" />
        <button
          onClick={() => setShowForm(v => !v)}
          className="px-3 py-1 text-xs font-medium bg-primary text-primary-foreground rounded hover:opacity-90 transition-opacity"
        >
          + New Import
        </button>
      </header>

      {/* Create form */}
      {showForm && (
        <form
          onSubmit={handleCreate}
          className="px-4 py-3 border-b border-border bg-accent/20 shrink-0"
        >
          <div className="flex flex-col gap-2 max-w-lg">
            <div className="text-sm font-medium text-foreground">New Import Job</div>
            <input
              type="text"
              value={formName}
              onChange={e => setFormName(e.target.value)}
              placeholder="Job name (optional)"
              className="px-3 py-1.5 text-sm border border-border rounded bg-background focus:outline-none"
            />
            <input
              type="text"
              value={formPath}
              onChange={e => setFormPath(e.target.value)}
              placeholder="Source directory path (e.g. /Users/me/docs)"
              required
              className="px-3 py-1.5 text-sm border border-border rounded bg-background focus:outline-none"
            />
            {createError && (
              <p className="text-xs text-red-500">{createError}</p>
            )}
            <div className="flex gap-2">
              <button
                type="submit"
                disabled={creating || !formPath.trim()}
                className="px-4 py-1.5 text-sm font-medium bg-primary text-primary-foreground rounded disabled:opacity-50"
              >
                {creating ? 'Scanning…' : 'Create & Scan'}
              </button>
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="px-4 py-1.5 text-sm border border-border rounded hover:bg-accent"
              >
                Cancel
              </button>
            </div>
          </div>
        </form>
      )}

      {/* Main split: job list + detail */}
      <div className="flex flex-1 overflow-hidden">
        {/* Job list */}
        <aside className="w-72 shrink-0 border-r border-border overflow-y-auto">
          {loading && (
            <div className="p-4 text-sm text-muted-foreground text-center">Loading…</div>
          )}
          {!loading && jobs.length === 0 && (
            <div className="p-4 text-sm text-muted-foreground text-center">
              No import jobs yet. Click "+ New Import" to start.
            </div>
          )}
          {jobs.map(job => (
            <button
              key={job.id}
              onClick={() => handleSelectJob(job)}
              className={`w-full text-left px-3 py-3 border-b border-border hover:bg-accent transition-colors ${
                selectedJob?.id === job.id ? 'bg-accent' : ''
              }`}
            >
              <div className="flex items-start justify-between gap-1 mb-1">
                <span className="text-sm font-medium truncate text-foreground flex-1 min-w-0">
                  {job.name}
                </span>
                <StatusBadge status={job.status} />
              </div>
              {job.total_count > 0 && (
                <div className="mb-1">
                  <ProgressBar total={job.total_count} processed={job.processed_count} />
                </div>
              )}
              <div className="text-xs text-muted-foreground">
                {job.total_count > 0
                  ? `${job.processed_count}/${job.total_count} · ✓${job.success_count} ⊘${job.skipped_count} ✗${job.failed_count}`
                  : 'No files scanned'}
              </div>
              <div className="text-xs text-muted-foreground mt-0.5">{formatDate(job.created_at)}</div>
            </button>
          ))}
        </aside>

        {/* Job detail */}
        <main className="flex-1 flex flex-col overflow-hidden">
          {!selectedJob ? (
            <div className="flex-1 flex items-center justify-center text-sm text-muted-foreground">
              Select an import job to see details
            </div>
          ) : (
            <>
              {/* Job detail header */}
              <div className="px-4 py-3 border-b border-border shrink-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-foreground truncate">{selectedJob.name}</div>
                    <div className="text-xs text-muted-foreground truncate">{selectedJob.source_path}</div>
                  </div>
                  <StatusBadge status={selectedJob.status} />

                  {/* Action buttons */}
                  {(selectedJob.status === 'pending' || selectedJob.status === 'paused') && (
                    <button
                      onClick={() => handleAction('start', selectedJob.id)}
                      className="px-3 py-1 text-xs font-medium bg-primary text-primary-foreground rounded hover:opacity-90"
                    >
                      {selectedJob.status === 'paused' ? 'Resume' : 'Start'}
                    </button>
                  )}
                  {selectedJob.status === 'running' && (
                    <button
                      onClick={() => handleAction('pause', selectedJob.id)}
                      className="px-3 py-1 text-xs font-medium border border-border rounded hover:bg-accent"
                    >
                      Pause
                    </button>
                  )}
                  {selectedJob.failed_count > 0 && selectedJob.status !== 'running' && (
                    <button
                      onClick={() => handleAction('retry', selectedJob.id)}
                      className="px-3 py-1 text-xs font-medium border border-amber-400 text-amber-600 rounded hover:bg-amber-50"
                    >
                      Retry Failed
                    </button>
                  )}
                  {(selectedJob.status === 'running' || selectedJob.status === 'pending' || selectedJob.status === 'paused') && (
                    <button
                      onClick={() => { if (window.confirm('Cancel this import job?')) handleAction('cancel', selectedJob.id) }}
                      className="px-3 py-1 text-xs font-medium text-red-500 border border-red-200 rounded hover:bg-red-50"
                    >
                      Cancel
                    </button>
                  )}
                </div>

                {/* Progress */}
                {selectedJob.total_count > 0 && (
                  <div className="mt-3">
                    <ProgressBar total={selectedJob.total_count} processed={selectedJob.processed_count} />
                    <div className="flex justify-between text-xs text-muted-foreground mt-1">
                      <span>{selectedJob.processed_count} / {selectedJob.total_count} processed</span>
                      <span className="flex gap-3">
                        <span className="text-green-600">✓ {selectedJob.success_count}</span>
                        <span className="text-gray-400">⊘ {selectedJob.skipped_count}</span>
                        <span className="text-red-500">✗ {selectedJob.failed_count}</span>
                      </span>
                    </div>
                  </div>
                )}
                {selectedJob.error_message && (
                  <p className="mt-2 text-xs text-red-500">{selectedJob.error_message}</p>
                )}
              </div>

              {/* Items */}
              <div className="px-4 py-2 border-b border-border shrink-0 flex gap-2 items-center">
                <span className="text-xs text-muted-foreground">Filter:</span>
                {(['', 'pending', 'success', 'skipped', 'failed'] as const).map(s => (
                  <button
                    key={s}
                    onClick={() => handleItemFilterChange(s)}
                    className={`px-2 py-0.5 text-xs rounded-full border transition-colors ${
                      itemFilter === s
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'border-border text-muted-foreground hover:bg-accent'
                    }`}
                  >
                    {s || 'All'}
                  </button>
                ))}
              </div>

              <div className="flex-1 overflow-y-auto">
                {itemsLoading && (
                  <div className="p-4 text-sm text-muted-foreground text-center">Loading…</div>
                )}
                {!itemsLoading && items.length === 0 && (
                  <div className="p-4 text-sm text-muted-foreground text-center">No items</div>
                )}
                {items.map(item => (
                  <div
                    key={item.id}
                    className="px-4 py-2 border-b border-border flex items-start justify-between gap-3"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="text-xs font-mono truncate text-foreground" title={item.file_path}>
                        {item.file_path.split('/').slice(-2).join('/')}
                      </div>
                      {item.error_message && (
                        <div className="text-xs text-red-500 mt-0.5 truncate" title={item.error_message}>
                          {item.error_message}
                        </div>
                      )}
                    </div>
                    <StatusBadge status={item.status} />
                  </div>
                ))}
              </div>
            </>
          )}
        </main>
      </div>
    </div>
  )
}
