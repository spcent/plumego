import { useEffect, useState } from 'react'
import { searchAPI, type IndexStatus } from '../api/search'

function StatBox({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="flex flex-col items-center p-4 rounded-lg border border-border bg-background">
      <span className={`text-2xl font-semibold font-mono ${color}`}>{value.toLocaleString()}</span>
      <span className="text-xs text-muted-foreground mt-1">{label}</span>
    </div>
  )
}

export default function IndexPage() {
  const [status, setStatus] = useState<IndexStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [reindexing, setReindexing] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    load()
    // Refresh every 5s while pending > 0.
    const id = setInterval(() => {
      if (status && (status.pending > 0 || status.failed > 0 || status.stale > 0)) {
        load()
      }
    }, 5000)
    return () => clearInterval(id)
  }, [status?.pending])

  async function load() {
    try {
      const s = await searchAPI.getIndexStatus()
      setStatus(s)
    } catch {
      // ignore
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
      await load()
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Reindex failed')
    } finally {
      setReindexing(false)
    }
  }

  function formatDate(iso?: string) {
    if (!iso) return '—'
    return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
  }

  const indexed = status?.indexed ?? 0
  const total = status?.total_documents ?? 0
  const pct = total > 0 ? Math.round((indexed / total) * 100) : 0

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="max-w-2xl mx-auto px-6 py-6 space-y-6">
        <div>
          <h2 className="text-base font-semibold text-foreground">Search Index</h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            SQLite FTS5 index status and maintenance
          </p>
        </div>

        {loading && (
          <div className="text-sm text-muted-foreground">Loading…</div>
        )}

        {status && (
          <>
            {/* Progress bar */}
            <div>
              <div className="flex justify-between text-xs text-muted-foreground mb-1">
                <span>{indexed.toLocaleString()} / {total.toLocaleString()} indexed</span>
                <span>{pct}%</span>
              </div>
              <div className="w-full bg-gray-100 rounded-full h-2 overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
              <StatBox label="Indexed" value={status.indexed} color="text-green-600" />
              <StatBox label="Pending" value={status.pending} color="text-amber-500" />
              <StatBox label="Failed" value={status.failed} color="text-red-500" />
              <StatBox label="Stale" value={status.stale} color="text-blue-500" />
            </div>

            {/* Last indexed */}
            <div className="text-xs text-muted-foreground">
              Last indexed: {formatDate(status.last_indexed_at)}
            </div>

            {/* Actions */}
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => handleReindex('all')}
                disabled={reindexing}
                className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded hover:opacity-90 disabled:opacity-50"
              >
                {reindexing ? 'Scheduling…' : 'Reindex All'}
              </button>
              {status.failed > 0 && (
                <button
                  onClick={() => handleReindex('failed')}
                  disabled={reindexing}
                  className="px-4 py-2 text-sm border border-amber-400 text-amber-600 rounded hover:bg-amber-50 disabled:opacity-50"
                >
                  Retry Failed ({status.failed})
                </button>
              )}
              {status.stale > 0 && (
                <button
                  onClick={() => handleReindex('stale')}
                  disabled={reindexing}
                  className="px-4 py-2 text-sm border border-border rounded hover:bg-accent disabled:opacity-50"
                >
                  Reindex Stale ({status.stale})
                </button>
              )}
              <button
                onClick={load}
                className="px-4 py-2 text-sm border border-border rounded hover:bg-accent"
              >
                Refresh
              </button>
            </div>

            {message && (
              <p className="text-xs text-muted-foreground">{message}</p>
            )}

            {/* Info box */}
            <div className="border border-border rounded p-3 text-xs text-muted-foreground space-y-1">
              <div className="font-medium text-foreground">Notes</div>
              <div>• The background indexer runs every {5} seconds and processes pending documents.</div>
              <div>• "Stale" means the document was updated after it was last indexed.</div>
              <div>• Files larger than the configured max_content_size_mb are indexed by title and headings only.</div>
              <div>• Chinese text is tokenized at the Unicode character level (unicode61 tokenizer).</div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
