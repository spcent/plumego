import { useEffect, useState } from 'react'
import {
  acceptTagSuggestion,
  confirmSimilarity,
  getReviewQueue,
  ignoreSimilarity,
  rejectTagSuggestion,
  runAll,
  type ReviewItem,
} from '../api/organize'
import { Badge, Button, EmptyState, PageFrame, Panel, SegmentedControl, SkeletonRows, StatusBanner } from '../components/ui'

const REVIEW_TYPES = [
  { value: '', label: 'All' },
  { value: 'duplicates', label: 'Duplicates' },
  { value: 'similar', label: 'Similar' },
  { value: 'tag_suggestions', label: 'Tags' },
  { value: 'prompt_candidates', label: 'Prompts' },
  { value: 'low_quality', label: 'Quality' },
] as const

function typeTone(type: string): 'neutral' | 'accent' | 'success' | 'warning' | 'danger' {
  if (type === 'duplicate') return 'danger'
  if (type === 'similar') return 'accent'
  if (type === 'tag_suggestion') return 'success'
  if (type === 'low_quality') return 'warning'
  return 'neutral'
}

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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load review queue')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to run analyzers')
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
      setItems(prev => prev.filter(queueItem => queueItem !== item))
      setTotal(t => Math.max(0, t - 1))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update review item')
    }
  }

  return (
    <PageFrame
      title="Review Queue"
      description={`${total.toLocaleString()} item${total === 1 ? '' : 's'} waiting for curation or confirmation.`}
      action={<Button variant="primary" size="sm" icon="check" onClick={handleRunAll} disabled={running}>{running ? 'Running...' : 'Run Analyzers'}</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      <SegmentedControl options={REVIEW_TYPES} value={activeType} onChange={setActiveType} />
      <Panel>
        {loading && <SkeletonRows count={8} />}
        {!loading && items.length === 0 && (
          <EmptyState compact icon="check" title="Nothing to review" description="Run analyzers to populate this queue." />
        )}
        {!loading && items.length > 0 && (
          <div className="divide-y divide-border">
            {items.map((item, index) => (
              <ReviewItemCard key={index} item={item} onAction={action => handleAction(item, action)} />
            ))}
          </div>
        )}
      </Panel>
    </PageFrame>
  )
}

function ReviewItemCard({ item, onAction }: { item: ReviewItem; onAction: (action: string) => void }) {
  return (
    <div className="flex items-start gap-3 px-4 py-3">
      <Badge tone={typeTone(item.type)}>{item.type.replace('_', ' ')}</Badge>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium text-foreground">{item.title}</div>
        {item.type === 'tag_suggestion' && item.extra?.tag_name && (
          <div className="text-xs text-muted-foreground">Suggested tag: <span className="font-medium text-foreground">{item.extra.tag_name}</span> · confidence {formatScore(item.score)}%</div>
        )}
        {item.type === 'similar' && item.extra?.other_title && (
          <div className="truncate text-xs text-muted-foreground">Similar to: {item.extra.other_title} · {formatScore(item.score, 100)}%</div>
        )}
        {item.type === 'prompt_candidate' && <div className="text-xs text-muted-foreground">Prompt score: {formatScore(item.score)}</div>}
        {item.type === 'low_quality' && <div className="text-xs text-muted-foreground">Quality score: {formatScore(item.score)}</div>}
        {item.type === 'duplicate' && item.extra?.count && (
          <div className="text-xs text-muted-foreground">{item.extra.count} copies with hash {item.extra.content_hash?.slice(0, 8)}...</div>
        )}
      </div>
      <div className="flex shrink-0 gap-2">
        {item.type === 'tag_suggestion' && (
          <>
            <Button size="sm" variant="secondary" onClick={() => onAction('accept')}>Accept</Button>
            <Button size="sm" variant="ghost" onClick={() => onAction('reject')}>Reject</Button>
          </>
        )}
        {item.type === 'similar' && (
          <>
            <Button size="sm" variant="secondary" onClick={() => onAction('confirm')}>Confirm</Button>
            <Button size="sm" variant="ghost" onClick={() => onAction('ignore')}>Ignore</Button>
          </>
        )}
      </div>
    </div>
  )
}

function formatScore(score: number | undefined, multiplier = 1) {
  if (typeof score !== 'number' || Number.isNaN(score)) return '—'
  return Math.round(score * multiplier).toLocaleString()
}
