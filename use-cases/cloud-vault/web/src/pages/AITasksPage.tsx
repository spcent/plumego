import { useState, useEffect } from 'react'
import {
  listTasks,
  cancelTask,
  enqueueQA,
  type AITask,
} from '../api/ai'
import { documentsAPI } from '../api/documents'
import type { DocumentSummary } from '../api/documents'

const STATUS_COLORS: Record<string, string> = {
  pending:   'bg-yellow-100 text-yellow-700',
  running:   'bg-blue-100 text-blue-700',
  completed: 'bg-green-100 text-green-700',
  failed:    'bg-red-100 text-red-700',
  cancelled: 'bg-muted text-muted-foreground',
}

const TYPE_LABELS: Record<string, string> = {
  document_summary: 'Summary',
  topic_summary:    'Topic Summary',
  search_answer:    'Q&A',
  prompt_extract:   'Prompt Extract',
}

export default function AITasksPage() {
  const [tasks, setTasks] = useState<AITask[]>([])
  const [total, setTotal] = useState(0)
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [expanded, setExpanded] = useState<string | null>(null)

  // Q&A form
  const [showQA, setShowQA] = useState(false)
  const [question, setQuestion] = useState('')
  const [docs, setDocs] = useState<DocumentSummary[]>([])
  const [selectedDocIDs, setSelectedDocIDs] = useState<string[]>([])
  const [qaSaving, setQASaving] = useState(false)

  useEffect(() => { loadTasks() }, [status])

  async function loadTasks() {
    setLoading(true)
    setError('')
    try {
      const data = await listTasks(status || undefined, 50, 0)
      setTasks(data.items ?? [])
      setTotal(data.total)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function loadDocs() {
    try {
      const data = await documentsAPI.list({ limit: 100, offset: 0 })
      setDocs(data.items ?? [])
    } catch {
      // ignore
    }
  }

  async function handleCancel(id: string) {
    setError('')
    try {
      await cancelTask(id)
      await loadTasks()
    } catch (e: any) {
      setError(e.message)
    }
  }

  async function handleAskQA() {
    if (!question.trim() || selectedDocIDs.length === 0) return
    setQASaving(true)
    setError('')
    try {
      await enqueueQA(question.trim(), selectedDocIDs)
      setShowQA(false)
      setQuestion('')
      setSelectedDocIDs([])
      await loadTasks()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setQASaving(false)
    }
  }

  function toggleDocSelect(id: string) {
    setSelectedDocIDs(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    )
  }

  function parseOutput(task: AITask) {
    if (!task.output_json) return null
    try { return JSON.parse(task.output_json) } catch { return null }
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">AI Tasks</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{total} task{total !== 1 ? 's' : ''}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => { setShowQA(true); loadDocs() }}
            className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90"
          >
            Ask a Question
          </button>
          <button onClick={loadTasks} className="text-xs px-2 py-1.5 rounded border border-border hover:bg-muted">
            Refresh
          </button>
        </div>
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      {/* Status filter */}
      <div className="flex gap-1 px-4 pt-3 shrink-0">
        {['', 'pending', 'running', 'completed', 'failed'].map(s => (
          <button
            key={s}
            onClick={() => setStatus(s)}
            className={`text-xs px-2 py-1 rounded border ${status === s ? 'bg-primary text-primary-foreground border-primary' : 'border-border hover:bg-muted'}`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      {/* Q&A form */}
      {showQA && (
        <div className="mx-4 mt-3 p-3 border border-border rounded-lg bg-muted/20 space-y-2 shrink-0">
          <p className="text-xs font-medium">Ask a question (grounded in selected documents)</p>
          <textarea
            rows={3}
            placeholder="Your question..."
            value={question}
            onChange={e => setQuestion(e.target.value)}
            className="w-full text-xs border border-border rounded px-2 py-1.5 bg-background resize-none"
          />
          <p className="text-xs text-muted-foreground">Select documents to ground the answer:</p>
          <div className="max-h-40 overflow-y-auto space-y-1 border border-border rounded p-2 bg-background">
            {docs.length === 0 && <p className="text-xs text-muted-foreground">Loading...</p>}
            {docs.map(d => (
              <label key={d.id} className="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted/30 rounded px-1">
                <input
                  type="checkbox"
                  checked={selectedDocIDs.includes(d.id)}
                  onChange={() => toggleDocSelect(d.id)}
                  className="accent-primary"
                />
                <span className="truncate">{d.title}</span>
              </label>
            ))}
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleAskQA}
              disabled={qaSaving || !question.trim() || selectedDocIDs.length === 0}
              className="text-xs px-3 py-1 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {qaSaving ? 'Enqueueing…' : 'Ask'}
            </button>
            <button onClick={() => setShowQA(false)} className="text-xs px-3 py-1 rounded border border-border hover:bg-muted">
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && tasks.length === 0 && (
          <p className="text-xs text-muted-foreground">No tasks yet. Use the Vault tab to summarize documents or ask questions.</p>
        )}

        {tasks.map(task => {
          const out = parseOutput(task)
          const isExpanded = expanded === task.id
          return (
            <div key={task.id} className="border border-border rounded-lg overflow-hidden">
              <div
                className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-muted/20"
                onClick={() => setExpanded(isExpanded ? null : task.id)}
              >
                <span className={`text-xs px-1.5 py-0.5 rounded shrink-0 ${STATUS_COLORS[task.status] ?? 'bg-muted text-muted-foreground'}`}>
                  {task.status}
                </span>
                <span className="text-xs text-muted-foreground shrink-0">
                  {TYPE_LABELS[task.task_type] ?? task.task_type}
                </span>
                <span className="text-xs text-muted-foreground truncate flex-1">{task.id}</span>
                <span className="text-xs text-muted-foreground shrink-0">{task.created_at.slice(0, 10)}</span>
                {task.status === 'pending' && (
                  <button
                    onClick={e => { e.stopPropagation(); handleCancel(task.id) }}
                    className="text-xs text-muted-foreground hover:text-destructive shrink-0"
                  >
                    Cancel
                  </button>
                )}
              </div>

              {isExpanded && (
                <div className="px-3 pb-3 space-y-2 border-t border-border bg-muted/10">
                  {task.error_message && (
                    <p className="text-xs text-destructive mt-2">Error: {task.error_message}</p>
                  )}
                  {task.output_document_id && (
                    <p className="text-xs text-muted-foreground mt-2">
                      Output document: <span className="font-mono">{task.output_document_id}</span>
                    </p>
                  )}
                  {out && (
                    <div className="mt-2">
                      {out.summary && (
                        <div>
                          <p className="text-xs font-medium mb-1">Summary</p>
                          <p className="text-xs text-muted-foreground">{out.summary}</p>
                        </div>
                      )}
                      {out.answer && (
                        <div>
                          <p className="text-xs font-medium mb-1">Answer</p>
                          <p className="text-xs text-muted-foreground whitespace-pre-wrap">{out.answer}</p>
                        </div>
                      )}
                      {out.key_points?.length > 0 && (
                        <div className="mt-2">
                          <p className="text-xs font-medium mb-1">Key Points</p>
                          <ul className="list-disc list-inside space-y-0.5">
                            {out.key_points.map((p: string, i: number) => (
                              <li key={i} className="text-xs text-muted-foreground">{p}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                      {out.citations?.length > 0 && (
                        <div className="mt-2">
                          <p className="text-xs font-medium mb-1">Citations</p>
                          {out.citations.map((c: any, i: number) => (
                            <div key={i} className="text-xs text-muted-foreground border-l-2 border-border pl-2 mb-1">
                              <span className="font-medium">{c.document_title}</span>: {c.excerpt}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
