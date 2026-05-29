import { useState, useEffect } from 'react'
import {
  listPrompts,
  deletePrompt,
  type AIPrompt,
} from '../api/ai'

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<AIPrompt[]>([])
  const [total, setTotal] = useState(0)
  const [scenario, setScenario] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<AIPrompt | null>(null)

  useEffect(() => { loadPrompts() }, [scenario])

  async function loadPrompts() {
    setLoading(true)
    setError('')
    try {
      const data = await listPrompts(scenario || undefined, 50, 0)
      setPrompts(data.items ?? [])
      setTotal(data.total)
    } catch (e: any) {
      setError(e.message)
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
      setTotal(t => t - 1)
      if (selected?.id === id) setSelected(null)
    } catch (e: any) {
      setError(e.message)
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).catch(() => {})
  }

  if (selected) {
    return (
      <div className="h-full flex flex-col overflow-hidden">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border shrink-0">
          <button
            onClick={() => setSelected(null)}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            ← Prompts
          </button>
          <span className="text-xs text-muted-foreground">/</span>
          <h1 className="text-sm font-semibold truncate">{selected.title}</h1>
          <div className="ml-auto flex gap-2">
            <button
              onClick={() => copyToClipboard(selected.content)}
              className="text-xs px-3 py-1 rounded border border-border hover:bg-muted"
            >
              Copy
            </button>
            <button
              onClick={() => handleDelete(selected.id)}
              className="text-xs text-destructive hover:underline"
            >
              Delete
            </button>
          </div>
        </div>

        {error && (
          <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
            {error}
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4">
          <div className="flex gap-4 mb-3 flex-wrap">
            {selected.scenario && (
              <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">
                {selected.scenario}
              </span>
            )}
            {selected.model_hint && (
              <span className="text-xs px-2 py-0.5 rounded bg-blue-50 text-blue-700">
                {selected.model_hint}
              </span>
            )}
            {selected.quality_score > 0 && (
              <span className="text-xs text-muted-foreground">
                Quality: {(selected.quality_score * 100).toFixed(0)}%
              </span>
            )}
          </div>

          <div className="border border-border rounded-lg p-4 bg-muted/10 font-mono text-xs whitespace-pre-wrap leading-relaxed">
            {selected.content}
          </div>

          {selected.source_document_id && (
            <p className="text-xs text-muted-foreground mt-3">
              Source document: <span className="font-mono">{selected.source_document_id}</span>
            </p>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">Prompt Library</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{total} prompt{total !== 1 ? 's' : ''}</p>
        </div>
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      {/* Scenario filter */}
      <div className="flex gap-1 px-4 pt-3 shrink-0 flex-wrap">
        {['', 'summarization', 'code_review', 'qa', 'writing', 'analysis'].map(s => (
          <button
            key={s}
            onClick={() => setScenario(s)}
            className={`text-xs px-2 py-1 rounded border ${scenario === s ? 'bg-primary text-primary-foreground border-primary' : 'border-border hover:bg-muted'}`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && prompts.length === 0 && (
          <p className="text-xs text-muted-foreground">
            No prompts yet. Use the Vault tab to extract prompts from documents.
          </p>
        )}

        {prompts.map(p => (
          <div
            key={p.id}
            className="flex items-start gap-3 px-3 py-2.5 border border-border rounded-lg hover:bg-muted/20 cursor-pointer"
            onClick={() => setSelected(p)}
          >
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{p.title}</p>
              <p className="text-xs text-muted-foreground truncate mt-0.5 font-mono">
                {p.content.slice(0, 100)}{p.content.length > 100 ? '…' : ''}
              </p>
            </div>
            <div className="shrink-0 text-right space-y-1">
              {p.scenario && (
                <span className="block text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                  {p.scenario}
                </span>
              )}
              {p.quality_score > 0 && (
                <span className="block text-xs text-muted-foreground">
                  {(p.quality_score * 100).toFixed(0)}%
                </span>
              )}
            </div>
            <button
              onClick={e => { e.stopPropagation(); handleDelete(p.id) }}
              className="text-xs text-muted-foreground hover:text-destructive shrink-0 mt-0.5"
              title="Delete prompt"
            >
              ✕
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
