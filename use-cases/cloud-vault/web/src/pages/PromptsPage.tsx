import { useEffect, useState } from 'react'
import { deletePrompt, listPrompts, type AIPrompt } from '../api/ai'
import { Badge, Button, EmptyState, PageFrame, Panel, SegmentedControl, SkeletonRows, StatusBanner } from '../components/ui'

const SCENARIOS = [
  { value: '', label: 'All' },
  { value: 'summarization', label: 'Summary' },
  { value: 'code_review', label: 'Code review' },
  { value: 'qa', label: 'Q&A' },
  { value: 'writing', label: 'Writing' },
  { value: 'analysis', label: 'Analysis' },
] as const

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<AIPrompt[]>([])
  const [total, setTotal] = useState(0)
  const [scenario, setScenario] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<AIPrompt | null>(null)

  useEffect(() => {
    loadPrompts()
  }, [scenario])

  async function loadPrompts() {
    setLoading(true)
    setError('')
    try {
      const data = await listPrompts(scenario || undefined, 50, 0)
      setPrompts(data.items ?? [])
      setTotal(data.total)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load prompts')
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this prompt from the library?')) return
    setError('')
    try {
      await deletePrompt(id)
      setPrompts(prev => prev.filter(p => p.id !== id))
      setTotal(t => Math.max(0, t - 1))
      if (selected?.id === id) setSelected(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete prompt')
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).catch(() => {})
  }

  if (selected) {
    return (
      <PageFrame
        title={selected.title}
        description="Prompt detail and extracted source metadata."
        action={
          <>
            <Button size="sm" variant="secondary" icon="chevronLeft" onClick={() => setSelected(null)}>Back</Button>
            <Button size="sm" variant="secondary" icon="copy" onClick={() => copyToClipboard(selected.content)}>Copy</Button>
            <Button size="sm" variant="danger" icon="trash" onClick={() => handleDelete(selected.id)}>Delete</Button>
          </>
        }
        width="wide"
      >
        {error && <StatusBanner tone="danger">{error}</StatusBanner>}
        <Panel>
          <div className="space-y-4 p-4">
            <div className="flex flex-wrap gap-2">
              {selected.scenario && <Badge>{selected.scenario}</Badge>}
              {selected.model_hint && <Badge tone="accent">{selected.model_hint}</Badge>}
              {selected.quality_score > 0 && <Badge tone="success">Quality {(selected.quality_score * 100).toFixed(0)}%</Badge>}
            </div>
            <pre className="whitespace-pre-wrap rounded-lg border border-border bg-background/60 p-4 font-mono text-xs leading-6 text-foreground">
              {selected.content}
            </pre>
            {selected.source_document_id && (
              <div className="text-xs text-muted-foreground">Source document: <span className="font-mono">{selected.source_document_id}</span></div>
            )}
          </div>
        </Panel>
      </PageFrame>
    )
  }

  return (
    <PageFrame
      title="Prompt Library"
      description={`${total.toLocaleString()} prompt${total === 1 ? '' : 's'} extracted from vault documents.`}
      action={<Button variant="secondary" size="sm" icon="refresh" onClick={loadPrompts}>Refresh</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      <SegmentedControl options={SCENARIOS} value={scenario} onChange={setScenario} />

      <Panel>
        {loading && <SkeletonRows count={8} />}
        {!loading && prompts.length === 0 && (
          <EmptyState compact icon="bolt" title="No prompts" description="Extract prompts from documents in the vault to build this library." />
        )}
        {!loading && prompts.length > 0 && (
          <div className="divide-y divide-border">
            {prompts.map(prompt => (
              <button
                key={prompt.id}
                type="button"
                onClick={() => setSelected(prompt)}
                className="flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-accent/70"
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium text-foreground">{prompt.title}</div>
                  <div className="mt-1 truncate font-mono text-xs text-muted-foreground">
                    {prompt.content.slice(0, 120)}{prompt.content.length > 120 ? '...' : ''}
                  </div>
                </div>
                <div className="flex shrink-0 flex-col items-end gap-1">
                  {prompt.scenario && <Badge>{prompt.scenario}</Badge>}
                  {prompt.quality_score > 0 && <span className="text-xs text-muted-foreground">{(prompt.quality_score * 100).toFixed(0)}%</span>}
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  icon="trash"
                  onClick={event => { event.stopPropagation(); handleDelete(prompt.id) }}
                  aria-label="Delete prompt"
                />
              </button>
            ))}
          </div>
        )}
      </Panel>
    </PageFrame>
  )
}
