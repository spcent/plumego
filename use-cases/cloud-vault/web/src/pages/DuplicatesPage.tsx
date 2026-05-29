import { useState, useEffect } from 'react'
import {
  detectDuplicates,
  listDuplicates,
  resolveDuplicates,
  type DuplicateGroup,
  type DuplicateDoc,
} from '../api/organize'

export default function DuplicatesPage() {
  const [groups, setGroups] = useState<DuplicateGroup[]>([])
  const [loading, setLoading] = useState(false)
  const [detecting, setDetecting] = useState(false)
  const [error, setError] = useState('')

  // Per-group selected keep document
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
      // Default keep = first (oldest) document per group
      const defaults: Record<string, string> = {}
      for (const g of data.groups ?? []) {
        if (g.documents.length > 0) {
          defaults[g.content_hash] = g.documents[0].id
        }
      }
      setKeepMap(defaults)
    } catch (e: any) {
      setError(e.message)
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
    } catch (e: any) {
      setError(e.message)
    } finally {
      setDetecting(false)
    }
  }

  async function handleResolve(group: DuplicateGroup, action: string) {
    const keepId = keepMap[group.content_hash] || group.documents[0]?.id
    const dupIds = group.documents.filter(d => d.id !== keepId).map(d => d.id)
    if (dupIds.length === 0) return

    setError('')
    try {
      await resolveDuplicates({
        keep_document_id: keepId,
        duplicate_document_ids: dupIds,
        action,
      })
      setGroups(prev => prev.filter(g => g.content_hash !== group.content_hash))
    } catch (e: any) {
      setError(e.message)
    }
  }

  function setKeep(hash: string, docId: string) {
    setKeepMap(prev => ({ ...prev, [hash]: docId }))
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div>
          <h1 className="text-sm font-semibold">Duplicate Documents</h1>
          <p className="text-xs text-muted-foreground mt-0.5">
            {groups.length} duplicate group{groups.length !== 1 ? 's' : ''} found
          </p>
        </div>
        <button
          onClick={handleDetect}
          disabled={detecting}
          className="text-xs px-3 py-1.5 rounded bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
        >
          {detecting ? 'Detecting…' : 'Detect Duplicates'}
        </button>
      </div>

      {error && (
        <div className="mx-4 mt-3 text-xs text-destructive border border-destructive/30 rounded px-3 py-2">
          {error}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {loading && <p className="text-xs text-muted-foreground">Loading…</p>}
        {!loading && groups.length === 0 && (
          <p className="text-xs text-muted-foreground">
            No duplicates found. Click "Detect Duplicates" to scan.
          </p>
        )}

        {groups.map(group => (
          <DuplicateGroupCard
            key={group.content_hash}
            group={group}
            keepId={keepMap[group.content_hash] || group.documents[0]?.id}
            onSetKeep={id => setKeep(group.content_hash, id)}
            onResolve={action => handleResolve(group, action)}
          />
        ))}
      </div>
    </div>
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
    <div className="border border-border rounded-lg overflow-hidden">
      <div className="px-3 py-2 bg-muted/30 flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">
          Hash: {group.content_hash.slice(0, 12)}… · {group.documents.length} copies
        </span>
        <div className="flex gap-2">
          <button
            onClick={() => onResolve('archive')}
            className="text-xs px-2 py-1 rounded border border-border hover:bg-muted"
          >
            Archive duplicates
          </button>
          <button
            onClick={() => onResolve('mark_duplicate')}
            className="text-xs px-2 py-1 rounded border border-border hover:bg-muted"
          >
            Mark duplicate
          </button>
          <button
            onClick={() => onResolve('ignore')}
            className="text-xs px-2 py-1 rounded border border-border hover:bg-muted text-muted-foreground"
          >
            Ignore
          </button>
        </div>
      </div>

      <div className="divide-y divide-border">
        {group.documents.map(doc => (
          <DocRow
            key={doc.id}
            doc={doc}
            isKeep={doc.id === keepId}
            onSelect={() => onSetKeep(doc.id)}
          />
        ))}
      </div>
    </div>
  )
}

function DocRow({
  doc,
  isKeep,
  onSelect,
}: {
  doc: DuplicateDoc
  isKeep: boolean
  onSelect: () => void
}) {
  return (
    <div
      className={`flex items-center gap-3 px-3 py-2 ${isKeep ? 'bg-green-50 dark:bg-green-950/20' : ''}`}
    >
      <input
        type="radio"
        checked={isKeep}
        onChange={onSelect}
        title="Keep this document"
        className="shrink-0"
      />
      <div className="flex-1 min-w-0">
        <p className="text-xs font-medium truncate">{doc.title}</p>
        {doc.original_path && (
          <p className="text-xs text-muted-foreground truncate">{doc.original_path}</p>
        )}
      </div>
      <div className="shrink-0 text-right">
        <span
          className={`text-xs px-1.5 py-0.5 rounded ${
            doc.status === 'active'
              ? 'bg-green-100 text-green-700'
              : 'bg-muted text-muted-foreground'
          }`}
        >
          {doc.status}
        </span>
        {doc.is_favorite && (
          <span className="ml-1 text-xs text-yellow-500">★</span>
        )}
      </div>
      {isKeep && (
        <span className="shrink-0 text-xs text-green-600 font-medium">Keep</span>
      )}
    </div>
  )
}
