import { useEffect, useState } from 'react'
import { createCollectionFromSearch } from '../api/collections'
import { buildTopics, getTopic, listTopics, type DuplicateDoc, type Topic } from '../api/organize'
import { Badge, Button, EmptyState, PageFrame, Panel, SkeletonRows, StatusBanner } from '../components/ui'

type View = 'list' | 'detail'

export default function TopicsPage() {
  const [view, setView] = useState<View>('list')
  const [topics, setTopics] = useState<Topic[]>([])
  const [selected, setSelected] = useState<Topic | null>(null)
  const [docs, setDocs] = useState<DuplicateDoc[]>([])
  const [loading, setLoading] = useState(false)
  const [building, setBuilding] = useState(false)
  const [error, setError] = useState('')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load topics')
    } finally {
      setLoading(false)
    }
  }

  async function openTopic(topic: Topic) {
    setLoading(true)
    setError('')
    try {
      const detail = await getTopic(topic.id)
      setSelected(detail.topic)
      setDocs(detail.documents ?? [])
      setView('detail')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to open topic')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to build topics')
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
        document_ids: docs.map(doc => doc.id),
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create collection')
    } finally {
      setConverting(false)
    }
  }

  if (view === 'detail' && selected) {
    return (
      <PageFrame
        title={selected.name}
        description={selected.description || `${docs.length.toLocaleString()} document${docs.length === 1 ? '' : 's'} in this topic.`}
        action={
          <>
            <Button variant="secondary" size="sm" icon="chevronLeft" onClick={() => setView('list')}>Topics</Button>
            <Button variant="primary" size="sm" icon="folder" onClick={handleConvertToCollection} disabled={converting || docs.length === 0}>
              {converting ? 'Creating...' : 'Create Collection'}
            </Button>
          </>
        }
        width="wide"
      >
        {error && <StatusBanner tone="danger">{error}</StatusBanner>}
        <Panel>
          {docs.length === 0 ? (
            <EmptyState compact icon="grid" title="No documents" description="This topic does not have matching documents yet." />
          ) : (
            <div className="divide-y divide-border">
              {docs.map(doc => <DocumentRow key={doc.id} doc={doc} />)}
            </div>
          )}
        </Panel>
      </PageFrame>
    )
  }

  return (
    <PageFrame
      title="Topics"
      description={`${topics.length.toLocaleString()} rule-based topic${topics.length === 1 ? '' : 's'} generated from tags and paths.`}
      action={<Button variant="primary" size="sm" icon="grid" onClick={handleBuild} disabled={building}>{building ? 'Building...' : 'Build Topics'}</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      <Panel>
        {loading && <SkeletonRows count={6} />}
        {!loading && topics.length === 0 && (
          <EmptyState compact icon="grid" title="No topics" description="Build topics to group documents by tags and folder paths." />
        )}
        {!loading && topics.length > 0 && (
          <div className="grid gap-3 p-4 sm:grid-cols-2 xl:grid-cols-3">
            {topics.map(topic => (
              <button
                key={topic.id}
                type="button"
                onClick={() => openTopic(topic)}
                className="rounded-lg border border-border bg-background/45 p-4 text-left transition-colors hover:bg-accent/70"
              >
                <div className="truncate text-sm font-semibold text-foreground">{topic.name}</div>
                {topic.description && <div className="mt-1 truncate text-xs text-muted-foreground">{topic.description}</div>}
                <div className="mt-3 flex items-center gap-2">
                  <Badge>{topic.doc_count} docs</Badge>
                  <Badge tone="accent">{topic.source}</Badge>
                </div>
              </button>
            ))}
          </div>
        )}
      </Panel>
    </PageFrame>
  )
}

function DocumentRow({ doc }: { doc: DuplicateDoc }) {
  return (
    <div className="flex items-center gap-3 px-4 py-3">
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium text-foreground">{doc.title}</div>
        {doc.original_path && <div className="truncate text-xs text-muted-foreground">{doc.original_path}</div>}
      </div>
      <Badge tone={doc.status === 'active' ? 'success' : 'neutral'}>{doc.status}</Badge>
      {doc.is_favorite && <Badge tone="warning">Favorite</Badge>}
    </div>
  )
}
