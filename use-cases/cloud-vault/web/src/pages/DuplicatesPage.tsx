import { useEffect, useState } from 'react'
import { detectDuplicates, listDuplicates, resolveDuplicates, type DuplicateDoc, type DuplicateGroup } from '../api/organize'
import { Badge, Button, EmptyState, PageFrame, Panel, SkeletonRows, StatusBanner, cn } from '../components/ui'

export default function DuplicatesPage() {
  const [groups, setGroups] = useState<DuplicateGroup[]>([])
  const [loading, setLoading] = useState(false)
  const [detecting, setDetecting] = useState(false)
  const [error, setError] = useState('')
  const [keepMap, setKeepMap] = useState<Record<string, string>>({})

  useEffect(() => {
    loadDuplicates()
  }, [])

  async function loadDuplicates() {
    setLoading(true)
    setError('')
    try {
      const data = await listDuplicates()
      setGroups(data.groups ?? [])
      const defaults: Record<string, string> = {}
      for (const group of data.groups ?? []) {
        if (group.documents.length > 0) defaults[group.content_hash] = group.documents[0].id
      }
      setKeepMap(defaults)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load duplicates')
    } finally {
      setLoading(false)
    }
  }

  async function handleDetect() {
    setDetecting(true)
    setError('')
    try {
      await detectDuplicates()
      await loadDuplicates()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to detect duplicates')
    } finally {
      setDetecting(false)
    }
  }

  async function handleResolve(group: DuplicateGroup, action: string) {
    const keepId = keepMap[group.content_hash] || group.documents[0]?.id
    const duplicateIds = group.documents.filter(doc => doc.id !== keepId).map(doc => doc.id)
    if (duplicateIds.length === 0) return

    setError('')
    try {
      await resolveDuplicates({ keep_document_id: keepId, duplicate_document_ids: duplicateIds, action })
      setGroups(prev => prev.filter(g => g.content_hash !== group.content_hash))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to resolve duplicate group')
    }
  }

  return (
    <PageFrame
      title="Duplicate Documents"
      description={`${groups.length.toLocaleString()} duplicate group${groups.length === 1 ? '' : 's'} found.`}
      action={<Button variant="primary" size="sm" icon="copy" onClick={handleDetect} disabled={detecting}>{detecting ? 'Detecting...' : 'Detect'}</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      <Panel>
        {loading && <SkeletonRows count={6} />}
        {!loading && groups.length === 0 && (
          <EmptyState compact icon="copy" title="No duplicates found" description="Run detection to compare content hashes across documents." action={<Button variant="secondary" size="sm" onClick={handleDetect}>Detect duplicates</Button>} />
        )}
        {!loading && groups.length > 0 && (
          <div className="space-y-4 p-4">
            {groups.map(group => (
              <DuplicateGroupCard
                key={group.content_hash}
                group={group}
                keepId={keepMap[group.content_hash] || group.documents[0]?.id}
                onSetKeep={id => setKeepMap(prev => ({ ...prev, [group.content_hash]: id }))}
                onResolve={action => handleResolve(group, action)}
              />
            ))}
          </div>
        )}
      </Panel>
    </PageFrame>
  )
}

function DuplicateGroupCard({
  group,
  keepId,
  onSetKeep,
  onResolve,
}: {
  group: DuplicateGroup
  keepId: string
  onSetKeep: (id: string) => void
  onResolve: (action: string) => void
}) {
  return (
    <div className="overflow-hidden rounded-lg border border-border bg-surface">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-background/45 px-3 py-3">
        <div className="min-w-0">
          <div className="text-sm font-medium text-foreground">{group.documents.length} copies</div>
          <div className="truncate font-mono text-xs text-muted-foreground">hash {group.content_hash.slice(0, 18)}...</div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="secondary" onClick={() => onResolve('archive')}>Archive duplicates</Button>
          <Button size="sm" variant="secondary" onClick={() => onResolve('mark_duplicate')}>Mark duplicate</Button>
          <Button size="sm" variant="ghost" onClick={() => onResolve('ignore')}>Ignore</Button>
        </div>
      </div>
      <div className="divide-y divide-border">
        {group.documents.map(doc => (
          <DocRow key={doc.id} doc={doc} isKeep={doc.id === keepId} onSelect={() => onSetKeep(doc.id)} />
        ))}
      </div>
    </div>
  )
}

function DocRow({ doc, isKeep, onSelect }: { doc: DuplicateDoc; isKeep: boolean; onSelect: () => void }) {
  return (
    <label className={cn('flex cursor-pointer items-center gap-3 px-3 py-2.5 transition-colors hover:bg-accent/70', isKeep && 'bg-emerald-500/10')}>
      <input type="radio" checked={isKeep} onChange={onSelect} className="shrink-0 accent-primary" />
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium text-foreground">{doc.title}</div>
        {doc.original_path && <div className="truncate text-xs text-muted-foreground">{doc.original_path}</div>}
      </div>
      <Badge tone={doc.status === 'active' ? 'success' : 'neutral'}>{doc.status}</Badge>
      {doc.is_favorite && <Badge tone="warning">Favorite</Badge>}
      {isKeep && <Badge tone="success">Keep</Badge>}
    </label>
  )
}
