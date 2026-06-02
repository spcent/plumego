import { useEffect, useState } from 'react'
import { searchAPI, type IndexStatus } from '../api/search'
import { Badge, Button, MetricCard, PageFrame, Panel, ProgressLine, SkeletonRows, StatusBanner } from '../components/ui'

export default function IndexPage() {
  const [status, setStatus] = useState<IndexStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [reindexing, setReindexing] = useState(false)
  const [message, setMessage] = useState('')
  const [messageTone, setMessageTone] = useState<'neutral' | 'danger'>('neutral')

  useEffect(() => {
    load()
    const id = setInterval(() => {
      if (status && (status.pending > 0 || status.failed > 0 || status.stale > 0)) {
        load()
      }
    }, 5000)
    return () => clearInterval(id)
  }, [status?.pending, status?.failed, status?.stale])

  async function load() {
    try {
      const s = await searchAPI.getIndexStatus()
      setStatus(s)
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Failed to load index status')
      setMessageTone('danger')
    } finally {
      setLoading(false)
    }
  }

  async function handleReindex(scope: string) {
    setReindexing(true)
    setMessage('')
    try {
      await searchAPI.reindex({ scope })
      setMessage(`Reindex (${scope}) scheduled. Background indexer will process documents shortly.`)
      setMessageTone('neutral')
      await load()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Reindex failed')
      setMessageTone('danger')
    } finally {
      setReindexing(false)
    }
  }

  function formatDate(iso?: string) {
    if (!iso) return 'Never'
    return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
  }

  const indexed = status?.indexed ?? 0
  const total = status?.total_documents ?? 0
  const pct = total > 0 ? Math.round((indexed / total) * 100) : 0
  const needsAttention = Boolean(status && (status.failed > 0 || status.stale > 0 || status.pending > 0))

  return (
    <PageFrame
      title="Search Index"
      description="Maintain the local full-text index used by search, review, and document navigation."
      action={<Button variant="secondary" size="sm" icon="refresh" onClick={load}>Refresh</Button>}
      width="wide"
    >
      {loading && <Panel><SkeletonRows count={4} /></Panel>}

      {status && (
        <>
          <Panel
            title="Coverage"
            description="Indexed documents against total vault documents."
            action={<Badge tone={needsAttention ? 'warning' : 'success'}>{needsAttention ? 'Attention needed' : 'Current'}</Badge>}
          >
            <div className="space-y-4 p-4">
              <ProgressLine value={pct} label={`${indexed.toLocaleString()} / ${total.toLocaleString()} indexed`} />
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                <MetricCard label="Indexed" value={status.indexed.toLocaleString()} tone="success" />
                <MetricCard label="Pending" value={status.pending.toLocaleString()} tone="warning" />
                <MetricCard label="Failed" value={status.failed.toLocaleString()} tone="danger" />
                <MetricCard label="Stale" value={status.stale.toLocaleString()} tone="accent" />
              </div>
              <div className="text-xs text-muted-foreground">Last indexed: {formatDate(status.last_indexed_at)}</div>
            </div>
          </Panel>

          <Panel title="Maintenance" description="Schedule background index work without blocking your current session.">
            <div className="space-y-4 p-4">
              <div className="flex flex-wrap gap-2">
                <Button onClick={() => handleReindex('all')} disabled={reindexing} variant="primary" icon="refresh">
                  {reindexing ? 'Scheduling...' : 'Reindex All'}
                </Button>
                {status.failed > 0 && (
                  <Button onClick={() => handleReindex('failed')} disabled={reindexing} variant="secondary">
                    Retry Failed ({status.failed})
                  </Button>
                )}
                {status.stale > 0 && (
                  <Button onClick={() => handleReindex('stale')} disabled={reindexing} variant="secondary">
                    Reindex Stale ({status.stale})
                  </Button>
                )}
              </div>

              {message && <StatusBanner tone={messageTone}>{message}</StatusBanner>}

              <div className="grid gap-3 text-xs leading-5 text-muted-foreground md:grid-cols-2">
                <div className="rounded-lg border border-border bg-background/45 p-3">
                  <div className="mb-1 font-medium text-foreground">Background behavior</div>
                  The indexer checks for pending documents every 5 seconds and processes them in the background.
                </div>
                <div className="rounded-lg border border-border bg-background/45 p-3">
                  <div className="mb-1 font-medium text-foreground">Large files</div>
                  Files beyond the configured content limit are indexed by title and headings only.
                </div>
                <div className="rounded-lg border border-border bg-background/45 p-3">
                  <div className="mb-1 font-medium text-foreground">Stale documents</div>
                  Stale means a document changed after its last indexed version.
                </div>
                <div className="rounded-lg border border-border bg-background/45 p-3">
                  <div className="mb-1 font-medium text-foreground">Chinese text</div>
                  Chinese text uses Unicode character tokenization for reliable local search.
                </div>
              </div>
            </div>
          </Panel>
        </>
      )}
    </PageFrame>
  )
}
