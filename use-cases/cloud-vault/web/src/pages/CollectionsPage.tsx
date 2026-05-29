import { useState, useEffect } from 'react'
import {
  listCollections,
  createCollection,
  getCollection,
  deleteCollection,
  removeDocumentFromCollection,
  type Collection,
  type CollectionDocument,
} from '../api/collections'

type View = 'list' | 'detail'

export default function CollectionsPage() {
  const [view, setView] = useState<View>('list')
  const [collections, setCollections] = useState<Collection[]>([])
  const [selected, setSelected] = useState<Collection | null>(null)
  const [docs, setDocs] = useState<CollectionDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // Create form
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadCollections()
  }, [])

  async function loadCollections() {
    setLoading(true)
    setError('')
    try {
      const data = await listCollections()
      setCollections(data.items ?? [])
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function openCollection(c: Collection) {
    setLoading(true)
    setError('')
    try {
      const detail = await getCollection(c.id)
      setSelected(detail.collection)
      setDocs(detail.documents ?? [])
      setView('detail')
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate() {
    if (!newName.trim()) return
    setSaving(true)
    setError('')
    try {
      await createCollection({ name: newName.trim(), description: newDesc.trim() })
      setNewName('')
      setNewDesc('')
      setCreating(false)
      await loadCollections()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this collection? Documents will not be deleted.')) return
    setError('')
    try {
      await deleteCollection(id)
      setCollections(prev => prev.filter(c => c.id !== id))
      if (selected?.id === id) {
        setView('list')
        setSelected(null)
      }
    } catch (e: any) {
      setError(e.message)
    }
  }

  async function handleRemoveDoc(docId: string) {
    if (!selected) return
    setError('')
    try {
      await removeDocumentFromCollection(selected.id, docId)
      setDocs(prev => prev.filter(d => d.document_id !== docId))
    } catch (e: any) {
      setError(e.message)
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
            ← Collections
          </button>
          <span className="text-xs text-muted-foreground">/</span>
          <h1 className="text-sm font-semibold">{selected.name}</h1>
          {selected.description && (
            <span className="text-xs text-muted-foreground ml-2">{selected.description}</span>
          )}
          <div className="ml-auto">
            <button
              onClick={() => handleDelete(selected.id)}
              className="text-xs text-destructive hover:underline"
            >
              Delete collection
            </button>
          </div>
        </div>

        {error && (
          <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
            {error}
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4">
          <p className="text-xs text-muted-foreground mb-3">
            {docs.length} document{docs.length !== 1 ? 's' : ''}
          </p>

          {docs.length === 0 && (
            <p className="text-xs text-muted-foreground">No documents in this collection.</p>
          )}

          <div className="space-y-1">
            {docs.map(doc => (
              <div
                key={doc.document_id}
                className="flex items-center gap-3 px-3 py-2 border border-border rounded hover:bg-muted/30"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-medium truncate">{doc.title}</p>
                  {doc.note && (
                    <p className="text-xs text-muted-foreground truncate">{doc.note}</p>
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
                <button
                  onClick={() => handleRemoveDoc(doc.document_id)}
                  className="text-xs text-muted-foreground hover:text-destructive shrink-0"
                  title="Remove from collection"
                >
                  ✕
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">Collections</h1>
          <p className="text-xs text-muted-foreground mt-0.5">
            {collections.length} collection{collections.length !== 1 ? 's' : ''}
          </p>
        </div>
        <button
          onClick={() => setCreating(true)}
          className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90"
        >
          + New Collection
        </button>
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      {creating && (
        <div className="mx-4 mt-3 p-3 border border-border rounded-lg bg-muted/20 space-y-2">
          <input
            type="text"
            placeholder="Collection name"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            className="w-full text-xs border border-border rounded px-2 py-1.5 bg-background"
            autoFocus
          />
          <input
            type="text"
            placeholder="Description (optional)"
            value={newDesc}
            onChange={e => setNewDesc(e.target.value)}
            className="w-full text-xs border border-border rounded px-2 py-1.5 bg-background"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              disabled={saving || !newName.trim()}
              className="text-xs px-3 py-1 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {saving ? 'Creating…' : 'Create'}
            </button>
            <button
              onClick={() => setCreating(false)}
              className="text-xs px-3 py-1 rounded border border-border hover:bg-muted"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && collections.length === 0 && (
          <p className="text-xs text-muted-foreground">
            No collections yet. Create one to group related documents.
          </p>
        )}

        <div className="space-y-2">
          {collections.map(c => (
            <div
              key={c.id}
              className="flex items-center gap-3 px-3 py-2.5 border border-border rounded-lg hover:bg-muted/20 cursor-pointer"
              onClick={() => openCollection(c)}
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{c.name}</p>
                {c.description && (
                  <p className="text-xs text-muted-foreground truncate">{c.description}</p>
                )}
              </div>
              <div className="shrink-0 text-right">
                <span className="text-xs text-muted-foreground">
                  {c.doc_count} doc{c.doc_count !== 1 ? 's' : ''}
                </span>
                <span className="ml-2 text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground capitalize">
                  {c.type}
                </span>
              </div>
              <button
                onClick={e => { e.stopPropagation(); handleDelete(c.id) }}
                className="text-xs text-muted-foreground hover:text-destructive shrink-0"
                title="Delete collection"
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
