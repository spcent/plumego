import { useState, useEffect } from 'react'
import {
  listTopics,
  getTopic,
  buildTopics,
  type Topic,
  type DuplicateDoc,
} from '../api/organize'
import { createCollectionFromSearch } from '../api/collections'

type View = 'list' | 'detail'

export default function TopicsPage() {
  const [view, setView] = useState<View>('list')
  const [topics, setTopics] = useState<Topic[]>([])
  const [selected, setSelected] = useState<Topic | null>(null)
  const [docs, setDocs] = useState<DuplicateDoc[]>([])
  const [loading, setLoading] = useState(false)
  const [building, setBuilding] = useState(false)
  const [error, setError] = useState('')

  // Create collection from topic
  const [converting, setConverting] = useState(false)

  useEffect(() => {
    loadTopics()
  }, [])

  async function loadTopics() {
    setLoading(true)
    setError('')
    try {
      const data = await listTopics()
      setTopics(data.items ?? [])
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function openTopic(t: Topic) {
    setLoading(true)
    setError('')
    try {
      const detail = await getTopic(t.id)
      setSelected(detail.topic)
      setDocs(detail.documents ?? [])
      setView('detail')
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleBuild() {
    setBuilding(true)
    setError('')
    try {
      await buildTopics()
      await loadTopics()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setBuilding(false)
    }
  }

  async function handleConvertToCollection() {
    if (!selected) return
    setConverting(true)
    setError('')
    try {
      await createCollectionFromSearch({
        name: selected.name,
        description: `Created from topic: ${selected.name}`,
        document_ids: docs.map(d => d.id),
      })
      alert(`Collection "${selected.name}" created.`)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setConverting(false)
    }
  }

  if (view === 'detail' && selected) {
    return (
      <div className="h-full flex flex-col overflow-hidden">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-border shrink-0">
          <button
            onClick={() => setView('list')}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            ← Topics
          </button>
          <span className="text-xs text-muted-foreground">/</span>
          <h1 className="text-sm font-semibold">{selected.name}</h1>
          <span className="text-xs text-muted-foreground ml-1">({docs.length} docs)</span>
          <div className="ml-auto">
            <button
              onClick={handleConvertToCollection}
              disabled={converting || docs.length === 0}
              className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {converting ? 'Creating…' : 'Create Collection'}
            </button>
          </div>
        </div>

        {selected.description && (
          <p className="px-4 pt-2 text-xs text-muted-foreground">{selected.description}</p>
        )}

        {error && (
          <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
            {error}
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4 space-y-1">
          {docs.length === 0 && (
            <p className="text-xs text-muted-foreground">No documents in this topic.</p>
          )}
          {docs.map(doc => (
            <div
              key={doc.id}
              className="flex items-center gap-3 px-3 py-2 border border-border rounded hover:bg-muted/30"
            >
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate">{doc.title}</p>
                {doc.original_path && (
                  <p className="text-xs text-muted-foreground truncate">{doc.original_path}</p>
                )}
              </div>
              <span
                className={`text-xs px-1.5 py-0.5 rounded shrink-0 ${
                  doc.status === 'active'
                    ? 'bg-green-100 text-green-700'
                    : 'bg-muted text-muted-foreground'
                }`}
              >
                {doc.status}
              </span>
              {doc.is_favorite && <span className="text-xs text-yellow-500">★</span>}
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">Topics</h1>
          <p className="text-xs text-muted-foreground mt-0.5">
            {topics.length} topic{topics.length !== 1 ? 's' : ''} (rule-based)
          </p>
        </div>
        <button
          onClick={handleBuild}
          disabled={building}
          className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {building ? 'Building…' : 'Build Topics'}
        </button>
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && topics.length === 0 && (
          <p className="text-xs text-muted-foreground">
            No topics yet. Click "Build Topics" to generate from tags and paths.
          </p>
        )}

        <div className="grid grid-cols-2 gap-2">
          {topics.map(t => (
            <div
              key={t.id}
              className="border border-border rounded-lg p-3 cursor-pointer hover:bg-muted/20"
              onClick={() => openTopic(t)}
            >
              <p className="text-sm font-medium truncate">{t.name}</p>
              {t.description && (
                <p className="text-xs text-muted-foreground truncate mt-0.5">{t.description}</p>
              )}
              <div className="flex items-center gap-2 mt-1.5">
                <span className="text-xs text-muted-foreground">
                  {t.doc_count} doc{t.doc_count !== 1 ? 's' : ''}
                </span>
                <span className="text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                  {t.source}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
