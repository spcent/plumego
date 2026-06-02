import { useEffect, useState } from 'react'
import {
  createCollection,
  deleteCollection,
  getCollection,
  listCollections,
  removeDocumentFromCollection,
  type Collection,
  type CollectionDocument,
} from '../api/collections'
import { Badge, Button, EmptyState, Field, PageFrame, Panel, SkeletonRows, StatusBanner, TextInput, cn } from '../components/ui'

type View = 'list' | 'detail'

export default function CollectionsPage() {
  const [view, setView] = useState<View>('list')
  const [collections, setCollections] = useState<Collection[]>([])
  const [selected, setSelected] = useState<Collection | null>(null)
  const [docs, setDocs] = useState<CollectionDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load collections')
    } finally {
      setLoading(false)
    }
  }

  async function openCollection(collection: Collection) {
    setLoading(true)
    setError('')
    try {
      const detail = await getCollection(collection.id)
      setSelected(detail.collection)
      setDocs(detail.documents ?? [])
      setView('detail')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to open collection')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create collection')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this collection? Documents will not be deleted.')) return
    setError('')
    try {
      await deleteCollection(id)
      setCollections(prev => prev.filter(collection => collection.id !== id))
      if (selected?.id === id) {
        setView('list')
        setSelected(null)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete collection')
    }
  }

  async function handleRemoveDoc(docId: string) {
    if (!selected) return
    setError('')
    try {
      await removeDocumentFromCollection(selected.id, docId)
      setDocs(prev => prev.filter(doc => doc.document_id !== docId))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to remove document')
    }
  }

  if (view === 'detail' && selected) {
    return (
      <PageFrame
        title={selected.name}
        description={selected.description || `${docs.length.toLocaleString()} document${docs.length === 1 ? '' : 's'} in this collection.`}
        action={
          <>
            <Button size="sm" variant="secondary" icon="chevronLeft" onClick={() => setView('list')}>Collections</Button>
            <Button size="sm" variant="danger" icon="trash" onClick={() => handleDelete(selected.id)}>Delete</Button>
          </>
        }
        width="wide"
      >
        {error && <StatusBanner tone="danger">{error}</StatusBanner>}
        <Panel>
          {docs.length === 0 ? (
            <EmptyState compact icon="folder" title="No documents" description="Add documents to this collection from the vault." />
          ) : (
            <div className="divide-y divide-border">
              {docs.map(doc => (
                <div key={doc.document_id} className="flex items-center gap-3 px-4 py-3">
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-medium text-foreground">{doc.title}</div>
                    {doc.note && <div className="truncate text-xs text-muted-foreground">{doc.note}</div>}
                  </div>
                  <Badge tone={doc.status === 'active' ? 'success' : 'neutral'}>{doc.status}</Badge>
                  <Button size="sm" variant="ghost" icon="close" aria-label="Remove from collection" onClick={() => handleRemoveDoc(doc.document_id)} />
                </div>
              ))}
            </div>
          )}
        </Panel>
      </PageFrame>
    )
  }

  return (
    <PageFrame
      title="Collections"
      description={`${collections.length.toLocaleString()} collection${collections.length === 1 ? '' : 's'} grouping related documents.`}
      action={<Button variant="primary" size="sm" icon="plus" onClick={() => setCreating(true)}>New Collection</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      {creating && (
        <Panel title="New Collection" description="Create a named workspace for related documents.">
          <div className="grid gap-3 p-4 md:grid-cols-[1fr_1.4fr_auto] md:items-end">
            <Field label="Name">
              <TextInput value={newName} onChange={e => setNewName(e.target.value)} placeholder="Project notes" autoFocus />
            </Field>
            <Field label="Description">
              <TextInput value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Optional context" />
            </Field>
            <div className="flex gap-2">
              <Button onClick={handleCreate} disabled={saving || !newName.trim()} variant="primary">{saving ? 'Creating...' : 'Create'}</Button>
              <Button onClick={() => setCreating(false)} variant="ghost">Cancel</Button>
            </div>
          </div>
        </Panel>
      )}

      <Panel>
        {loading && <SkeletonRows count={6} />}
        {!loading && collections.length === 0 && (
          <EmptyState compact icon="folder" title="No collections" description="Create a collection to group related documents." />
        )}
        {!loading && collections.length > 0 && (
          <div className="grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
            {collections.map(collection => (
              <button
                key={collection.id}
                type="button"
                onClick={() => openCollection(collection)}
                className={cn('rounded-lg border border-border bg-background/45 p-4 text-left transition-colors hover:bg-accent/70')}
              >
                <div className="truncate text-sm font-semibold text-foreground">{collection.name}</div>
                {collection.description && <div className="mt-1 truncate text-xs text-muted-foreground">{collection.description}</div>}
                <div className="mt-3">
                  <Badge>{collection.doc_count} doc{collection.doc_count === 1 ? '' : 's'}</Badge>
                </div>
              </button>
            ))}
          </div>
        )}
      </Panel>
    </PageFrame>
  )
}
