import { useEffect, useState } from 'react'
import { cancelTask, enqueueQA, listTasks, type AITask } from '../api/ai'
import { documentsAPI, type DocumentSummary } from '../api/documents'
import { Badge, Button, EmptyState, PageFrame, Panel, SegmentedControl, SkeletonRows, StatusBanner, TextareaInput } from '../components/ui'

const STATUS_FILTERS = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'dead_letter', label: 'Dead Letter' },
] as const

const TYPE_LABELS: Record<string, string> = {
  document_summary: 'Summary',
  topic_summary: 'Topic Summary',
  search_answer: 'Q&A',
  prompt_extract: 'Prompt Extract',
}

function taskTone(status: string): 'neutral' | 'accent' | 'success' | 'warning' | 'danger' {
  if (status === 'pending') return 'warning'
  if (status === 'running') return 'accent'
  if (status === 'completed') return 'success'
  if (status === 'failed' || status === 'dead_letter') return 'danger'
  return 'neutral'
}

export default function AITasksPage() {
  const [tasks, setTasks] = useState<AITask[]>([])
  const [total, setTotal] = useState(0)
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [expanded, setExpanded] = useState<string | null>(null)

  const [showQA, setShowQA] = useState(false)
  const [question, setQuestion] = useState('')
  const [docs, setDocs] = useState<DocumentSummary[]>([])
  const [selectedDocIDs, setSelectedDocIDs] = useState<string[]>([])
  const [qaSaving, setQASaving] = useState(false)

  useEffect(() => {
    loadTasks()
  }, [status])

  async function loadTasks() {
    setLoading(true)
    setError('')
    try {
      const data = await listTasks(status || undefined, 50, 0)
      setTasks(data.items ?? [])
      setTotal(data.total)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load AI tasks')
    } finally {
      setLoading(false)
    }
  }

  async function loadDocs() {
    try {
      const data = await documentsAPI.list({ limit: 100, offset: 0 })
      setDocs(data.items ?? [])
    } catch {
      // Keep Q&A composer usable; the empty document state explains the next step.
    }
  }

  async function handleCancel(id: string) {
    setError('')
    try {
      await cancelTask(id)
      await loadTasks()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to cancel task')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to enqueue question')
    } finally {
      setQASaving(false)
    }
  }

  function toggleDocSelect(id: string) {
    setSelectedDocIDs(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }

  function parseOutput(task: AITask) {
    if (!task.output_json) return null
    try {
      return JSON.parse(task.output_json)
    } catch {
      return null
    }
  }

  return (
    <PageFrame
      title="AI Tasks"
      description={`${total.toLocaleString()} task${total === 1 ? '' : 's'} tracked across summaries, answers, and prompt extraction.`}
      action={
        <>
          <Button variant="secondary" size="sm" icon="refresh" onClick={loadTasks}>Refresh</Button>
          <Button variant="primary" size="sm" icon="spark" onClick={() => { setShowQA(true); loadDocs() }}>Ask</Button>
        </>
      }
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}

      {showQA && (
        <Panel title="Ask a grounded question" description="Select documents first, then enqueue a background answer task.">
          <div className="space-y-3 p-4">
            <TextareaInput rows={3} placeholder="Your question..." value={question} onChange={e => setQuestion(e.target.value)} />
            <div className="rounded-lg border border-border bg-background/45 p-2">
              {docs.length === 0 ? (
                <div className="px-2 py-3 text-xs text-muted-foreground">Loading documents...</div>
              ) : (
                <div className="max-h-44 space-y-1 overflow-y-auto">
                  {docs.map(doc => (
                    <label key={doc.id} className="flex cursor-pointer items-center gap-2 rounded px-2 py-1 text-xs hover:bg-accent">
                      <input type="checkbox" checked={selectedDocIDs.includes(doc.id)} onChange={() => toggleDocSelect(doc.id)} className="accent-primary" />
                      <span className="truncate">{doc.title}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>
            <div className="flex gap-2">
              <Button onClick={handleAskQA} disabled={qaSaving || !question.trim() || selectedDocIDs.length === 0} variant="primary">
                {qaSaving ? 'Enqueueing...' : 'Ask'}
              </Button>
              <Button onClick={() => setShowQA(false)} variant="ghost">Cancel</Button>
            </div>
          </div>
        </Panel>
      )}

      <div className="flex items-center justify-between gap-3">
        <SegmentedControl options={STATUS_FILTERS} value={status} onChange={setStatus} />
      </div>

      <Panel>
        {loading && <SkeletonRows count={7} />}
        {!loading && tasks.length === 0 && (
          <EmptyState compact icon="spark" title="No AI tasks" description="Summarize documents or ask a grounded question to create work here." />
        )}
        {!loading && tasks.length > 0 && (
          <div className="divide-y divide-border">
            {tasks.map(task => {
              const out = parseOutput(task)
              const isExpanded = expanded === task.id
              return (
                <div key={task.id}>
                  <button
                    type="button"
                    onClick={() => setExpanded(isExpanded ? null : task.id)}
                    className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-accent/70"
                  >
                    <Badge tone={taskTone(task.status)}>{task.status}</Badge>
                    <span className="shrink-0 text-xs font-medium text-muted-foreground">{TYPE_LABELS[task.task_type] ?? task.task_type}</span>
                    <span className="min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground">{task.id}</span>
                    <span className="hidden shrink-0 text-xs text-muted-foreground sm:block">{task.created_at.slice(0, 10)}</span>
                    {task.status === 'pending' && (
                      <Button size="sm" variant="ghost" onClick={e => { e.stopPropagation(); handleCancel(task.id) }}>Cancel</Button>
                    )}
                  </button>

                  {isExpanded && (
                    <div className="space-y-3 border-t border-border bg-background/45 px-4 py-3 text-sm">
                      {task.error_message && <StatusBanner tone="danger">Error: {task.error_message}</StatusBanner>}
                      {task.output_document_id && (
                        <div className="text-xs text-muted-foreground">Output document: <span className="font-mono">{task.output_document_id}</span></div>
                      )}
                      {out && (
                        <div className="space-y-3">
                          {out.summary && <OutputBlock title="Summary" text={out.summary} />}
                          {out.answer && <OutputBlock title="Answer" text={out.answer} />}
                          {out.key_points?.length > 0 && (
                            <div>
                              <div className="mb-1 text-xs font-medium text-foreground">Key Points</div>
                              <ul className="list-disc space-y-1 pl-4 text-xs leading-5 text-muted-foreground">
                                {out.key_points.map((point: string, index: number) => <li key={index}>{point}</li>)}
                              </ul>
                            </div>
                          )}
                          {out.citations?.length > 0 && (
                            <div>
                              <div className="mb-1 text-xs font-medium text-foreground">Citations</div>
                              <div className="space-y-2">
                                {out.citations.map((citation: any, index: number) => (
                                  <div key={index} className="border-l-2 border-border pl-3 text-xs leading-5 text-muted-foreground">
                                    <span className="font-medium text-foreground">{citation.document_title}</span>: {citation.excerpt}
                                  </div>
                                ))}
                              </div>
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
        )}
      </Panel>
    </PageFrame>
  )
}

function OutputBlock({ title, text }: { title: string; text: string }) {
  return (
    <div>
      <div className="mb-1 text-xs font-medium text-foreground">{title}</div>
      <div className="whitespace-pre-wrap rounded-md border border-border bg-surface p-3 text-xs leading-5 text-muted-foreground">{text}</div>
    </div>
  )
}
