import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { importsAPI, type ImportJob, type ImportJobItem } from '../api/imports'
import { Badge, Button, EmptyState, Field, Panel, ProgressLine, SegmentedControl, SkeletonRows, StatusBanner, TextInput, cn } from '../components/ui'

const ITEM_FILTERS = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'success', label: 'Success' },
  { value: 'skipped', label: 'Skipped' },
  { value: 'failed', label: 'Failed' },
] as const

function statusTone(status: string): 'neutral' | 'accent' | 'success' | 'warning' | 'danger' {
  if (status === 'running') return 'accent'
  if (status === 'done' || status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'paused' || status === 'pending') return 'warning'
  return 'neutral'
}

function StatusBadge({ status }: { status: string }) {
  return <Badge tone={statusTone(status)}>{status}</Badge>
}

function progressPct(total: number, processed: number) {
  return total > 0 ? Math.round((processed / total) * 100) : 0
}

export default function ImportPage() {
  const [jobs, setJobs] = useState<ImportJob[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedJob, setSelectedJob] = useState<ImportJob | null>(null)
  const [items, setItems] = useState<ImportJobItem[]>([])
  const [itemsLoading, setItemsLoading] = useState(false)
  const [itemFilter, setItemFilter] = useState('')

  const [showForm, setShowForm] = useState(false)
  const [formName, setFormName] = useState('')
  const [formPath, setFormPath] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  useEffect(() => {
    loadJobs()
  }, [])

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
      if (selectedJob) {
        const updated = result.items.find(j => j.id === selectedJob.id)
        if (updated) setSelectedJob(updated)
      }
    } catch {
      // Keep the current list visible if refresh fails.
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
      setCreateError(err instanceof Error ? err.message : 'Action failed')
    }
  }

  async function handleCreate(e: FormEvent) {
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
    <div className="h-full overflow-hidden bg-background">
      <div className="flex h-full flex-col">
        <div className="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
          <div className="min-w-0">
            <h1 className="text-xl font-semibold tracking-tight text-foreground">Import Jobs</h1>
            <p className="mt-1 text-sm text-muted-foreground">Scan local markdown folders and review every imported file.</p>
          </div>
          <div className="flex shrink-0 gap-2">
            <Button variant="secondary" size="sm" icon="refresh" onClick={loadJobs}>Refresh</Button>
            <Button variant="primary" size="sm" icon="plus" onClick={() => setShowForm(v => !v)}>
              New Import
            </Button>
          </div>
        </div>

        {showForm && (
          <form onSubmit={handleCreate} className="border-b border-border bg-surface px-5 py-4">
            <div className="grid max-w-4xl gap-3 md:grid-cols-[1fr_1.8fr_auto] md:items-end">
              <Field label="Job name" helper="Optional. Defaults to the folder path.">
                <TextInput value={formName} onChange={e => setFormName(e.target.value)} placeholder="Research notes" />
              </Field>
              <Field label="Source directory" error={createError}>
                <TextInput value={formPath} onChange={e => setFormPath(e.target.value)} placeholder="/Users/me/docs" required />
              </Field>
              <div className="flex gap-2">
                <Button type="submit" variant="primary" disabled={creating || !formPath.trim()}>
                  {creating ? 'Scanning...' : 'Create'}
                </Button>
                <Button variant="ghost" onClick={() => setShowForm(false)}>Cancel</Button>
              </div>
            </div>
          </form>
        )}

        <div className="grid min-h-0 flex-1 grid-cols-1 overflow-hidden lg:grid-cols-[20rem_1fr]">
          <aside className="min-h-0 border-b border-border bg-surface lg:border-b-0 lg:border-r">
            <div className="flex h-11 items-center justify-between border-b border-border px-3">
              <span className="text-xs font-semibold uppercase tracking-[0.08em] text-muted-foreground">Recent imports</span>
              <Badge>{jobs.length}</Badge>
            </div>
            <div className="h-[16rem] overflow-y-auto lg:h-[calc(100%-2.75rem)]">
              {loading && <SkeletonRows count={5} />}
              {!loading && jobs.length === 0 && (
                <EmptyState compact icon="import" title="No import jobs" description="Create an import to scan a markdown folder." />
              )}
              {jobs.map(job => {
                const pct = progressPct(job.total_count, job.processed_count)
                return (
                  <button
                    key={job.id}
                    onClick={() => handleSelectJob(job)}
                    className={cn(
                      'w-full border-b border-border px-3 py-3 text-left transition-colors hover:bg-accent/70',
                      selectedJob?.id === job.id && 'bg-primary/10',
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <div className="truncate text-sm font-medium text-foreground">{job.name}</div>
                        <div className="mt-0.5 truncate text-xs text-muted-foreground">{formatDate(job.created_at)}</div>
                      </div>
                      <StatusBadge status={job.status} />
                    </div>
                    {job.total_count > 0 && (
                      <div className="mt-2">
                        <ProgressLine value={pct} />
                        <div className="mt-1 flex justify-between text-xs text-muted-foreground">
                          <span>{job.processed_count}/{job.total_count}</span>
                          <span>{pct}%</span>
                        </div>
                      </div>
                    )}
                  </button>
                )
              })}
            </div>
          </aside>

          <main className="min-h-0 overflow-hidden">
            {!selectedJob ? (
              <EmptyState icon="import" title="Select an import job" description="Job details, actions, and file-level results appear here." />
            ) : (
              <div className="flex h-full flex-col">
                <Panel
                  className="m-4 mb-0 shrink-0"
                  title={selectedJob.name}
                  description={selectedJob.source_path}
                  action={<StatusBadge status={selectedJob.status} />}
                >
                  <div className="space-y-4 p-4">
                    {selectedJob.total_count > 0 && (
                      <ProgressLine
                        value={progressPct(selectedJob.total_count, selectedJob.processed_count)}
                        label={`${selectedJob.processed_count} / ${selectedJob.total_count} processed`}
                      />
                    )}
                    <div className="grid grid-cols-3 gap-2 text-xs md:grid-cols-5">
                      <ImportMetric label="Success" value={selectedJob.success_count} tone="success" />
                      <ImportMetric label="Skipped" value={selectedJob.skipped_count} />
                      <ImportMetric label="Failed" value={selectedJob.failed_count} tone="danger" />
                      <ImportMetric label="Processed" value={selectedJob.processed_count} tone="accent" />
                      <ImportMetric label="Total" value={selectedJob.total_count} />
                    </div>
                    {selectedJob.error_message && <StatusBanner tone="danger">{selectedJob.error_message}</StatusBanner>}
                    <div className="flex flex-wrap gap-2">
                      {(selectedJob.status === 'pending' || selectedJob.status === 'paused') && (
                        <Button size="sm" variant="primary" onClick={() => handleAction('start', selectedJob.id)}>
                          {selectedJob.status === 'paused' ? 'Resume' : 'Start'}
                        </Button>
                      )}
                      {selectedJob.status === 'running' && (
                        <Button size="sm" variant="secondary" onClick={() => handleAction('pause', selectedJob.id)}>Pause</Button>
                      )}
                      {selectedJob.failed_count > 0 && selectedJob.status !== 'running' && (
                        <Button size="sm" variant="secondary" onClick={() => handleAction('retry', selectedJob.id)}>Retry failed</Button>
                      )}
                      {(selectedJob.status === 'running' || selectedJob.status === 'pending' || selectedJob.status === 'paused') && (
                        <Button
                          size="sm"
                          variant="danger"
                          onClick={() => { if (window.confirm('Cancel this import job?')) handleAction('cancel', selectedJob.id) }}
                        >
                          Cancel
                        </Button>
                      )}
                    </div>
                  </div>
                </Panel>

                <div className="flex min-h-0 flex-1 flex-col p-4">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-semibold text-foreground">Files</div>
                      <div className="text-xs text-muted-foreground">Latest 100 results for this job</div>
                    </div>
                    <SegmentedControl options={ITEM_FILTERS} value={itemFilter} onChange={handleItemFilterChange} />
                  </div>
                  <div className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-border bg-surface">
                    {itemsLoading && <SkeletonRows count={6} />}
                    {!itemsLoading && items.length === 0 && <EmptyState compact icon="file" title="No files" description="No files match the current filter." />}
                    {items.map(item => (
                      <div key={item.id} className="flex items-start justify-between gap-3 border-b border-border px-3 py-2.5 last:border-b-0">
                        <div className="min-w-0 flex-1">
                          <div className="truncate font-mono text-xs text-foreground" title={item.file_path}>
                            {item.file_path.split('/').slice(-2).join('/')}
                          </div>
                          {item.error_message && (
                            <div className="mt-0.5 truncate text-xs text-destructive" title={item.error_message}>
                              {item.error_message}
                            </div>
                          )}
                        </div>
                        <StatusBadge status={item.status} />
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </main>
        </div>
      </div>
    </div>
  )
}

function ImportMetric({ label, value, tone = 'neutral' }: { label: string; value: number; tone?: 'neutral' | 'accent' | 'success' | 'danger' }) {
  return (
    <div className="rounded-md border border-border bg-background/45 px-3 py-2">
      <div
        className={cn(
          'font-mono text-lg font-semibold',
          tone === 'accent' && 'text-primary',
          tone === 'success' && 'text-emerald-600 dark:text-emerald-300',
          tone === 'danger' && 'text-destructive',
        )}
      >
        {value.toLocaleString()}
      </div>
      <div className="text-xs text-muted-foreground">{label}</div>
    </div>
  )
}
