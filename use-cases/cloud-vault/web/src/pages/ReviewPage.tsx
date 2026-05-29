import { useState, useEffect } from 'react'
import {
  getReviewQueue,
  acceptTagSuggestion,
  rejectTagSuggestion,
  ignoreSimilarity,
  confirmSimilarity,
  runAll,
  type ReviewItem,
} from '../api/organize'

const REVIEW_TYPES = [
  { key: '', label: 'All' },
  { key: 'duplicates', label: 'Duplicates' },
  { key: 'similar', label: 'Similar' },
  { key: 'tag_suggestions', label: 'Tag Suggestions' },
  { key: 'prompt_candidates', label: 'Prompt Candidates' },
  { key: 'low_quality', label: 'Low Quality' },
]

export default function ReviewPage() {
  const [activeType, setActiveType] = useState('')
  const [items, setItems] = useState<ReviewItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [running, setRunning] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    loadQueue()
  }, [activeType])

  async function loadQueue() {
    setLoading(true)
    setError('')
    try {
      const result = await getReviewQueue({ type: activeType || undefined, limit: 100 })
      setItems(result.items ?? [])
      setTotal(result.total)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleRunAll() {
    setRunning(true)
    setError('')
    try {
      await runAll()
      await loadQueue()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setRunning(false)
    }
  }

  async function handleAction(item: ReviewItem, action: string) {
    setError('')
    try {
      if (item.type === 'tag_suggestion' && item.extra?.suggestion_id) {
        if (action === 'accept') await acceptTagSuggestion(item.extra.suggestion_id)
        else if (action === 'reject') await rejectTagSuggestion(item.extra.suggestion_id)
      } else if (item.type === 'similar' && item.extra?.similarity_id) {
        if (action === 'ignore') await ignoreSimilarity(item.extra.similarity_id)
        else if (action === 'confirm') await confirmSimilarity(item.extra.similarity_id)
      }
      setItems(prev => prev.filter(i => i !== item))
      setTotal(t => Math.max(0, t - 1))
    } catch (e: any) {
      setError(e.message)
    }
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">Review Queue</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{total} item{total !== 1 ? 's' : ''}</p>
        </div>
        <button
          onClick={handleRunAll}
          disabled={running}
          className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {running ? 'Running…' : 'Run All Analyzers'}
        </button>
      </div>

      {/* Type filter */}
      <div className="flex gap-1 px-4 py-2 border-b border-border overflow-x-auto shrink-0">
        {REVIEW_TYPES.map(t => (
          <button
            key={t.key}
            onClick={() => setActiveType(t.key)}
            className={`text-xs px-3 py-1 rounded whitespace-nowrap ${
              activeType === t.key
                ? 'bg-primary text-primary-foreground'
                : 'border border-border hover:bg-muted'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && items.length === 0 && (
          <p className="text-xs text-muted-foreground">
            No items in this queue. Run analyzers first.
          </p>
        )}

        {items.map((item, i) => (
          <ReviewItemCard key={i} item={item} onAction={action => handleAction(item, action)} />
        ))}
      </div>
    </div>
  )
}

function ReviewItemCard({
  item,
  onAction,
}: {
  item: ReviewItem
  onAction: (action: string) => void
}) {
  const typeBadgeClass: Record<string, string> = {
    duplicate: 'bg-red-100 text-red-700',
    similar: 'bg-orange-100 text-orange-700',
    tag_suggestion: 'bg-blue-100 text-blue-700',
    prompt_candidate: 'bg-purple-100 text-purple-700',
    low_quality: 'bg-yellow-100 text-yellow-700',
  }

  return (
    <div className="border border-border rounded-lg px-3 py-2.5">
      <div className="flex items-start gap-2">
        <span
          className={`text-xs px-1.5 py-0.5 rounded shrink-0 ${
            typeBadgeClass[item.type] ?? 'bg-muted text-muted-foreground'
          }`}
        >
          {item.type.replace('_', ' ')}
        </span>
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium truncate">{item.title}</p>
          {item.type === 'tag_suggestion' && item.extra?.tag_name && (
            <p className="text-xs text-muted-foreground">
              Suggested tag: <strong>{item.extra.tag_name}</strong> · confidence:{' '}
              {Math.round(item.score)}%
            </p>
          )}
          {item.type === 'similar' && item.extra?.other_title && (
            <p className="text-xs text-muted-foreground truncate">
              Similar to: {item.extra.other_title} ({item.extra.similarity_type?.replace('_', ' ')}{' '}
              · {Math.round(item.score * 100)}%)
            </p>
          )}
          {item.type === 'prompt_candidate' && (
            <p className="text-xs text-muted-foreground">
              Prompt score: {Math.round(item.score)}
            </p>
          )}
          {item.type === 'low_quality' && (
            <p className="text-xs text-muted-foreground">
              Quality score: {Math.round(item.score)}
            </p>
          )}
          {item.type === 'duplicate' && item.extra?.count && (
            <p className="text-xs text-muted-foreground">
              {item.extra.count} copies with hash {item.extra.content_hash?.slice(0, 8)}…
            </p>
          )}
        </div>
        <div className="flex gap-1.5 shrink-0">
          {item.type === 'tag_suggestion' && (
            <>
              <button
                onClick={() => onAction('accept')}
                className="text-xs px-2 py-1 rounded bg-green-100 text-green-700 hover:bg-green-200"
              >
                Accept
              </button>
              <button
                onClick={() => onAction('reject')}
                className="text-xs px-2 py-1 rounded border border-border hover:bg-muted text-muted-foreground"
              >
                Reject
              </button>
            </>
          )}
          {item.type === 'similar' && (
            <>
              <button
                onClick={() => onAction('confirm')}
                className="text-xs px-2 py-1 rounded bg-blue-100 text-blue-700 hover:bg-blue-200"
              >
                Confirm
              </button>
              <button
                onClick={() => onAction('ignore')}
                className="text-xs px-2 py-1 rounded border border-border hover:bg-muted text-muted-foreground"
              >
                Ignore
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
