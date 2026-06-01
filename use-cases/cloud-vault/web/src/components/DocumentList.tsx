import { useEffect, useMemo, useState } from 'react'
import type { Tag } from '../api/tags'
import type { DocumentSummary, ListDocumentsParams } from '../api/documents'
import { Badge, Button, EmptyState, IconButton, SelectInput, SkeletonRows, TextInput, cn } from './ui'
import { Icon } from './icons'

export interface DocumentFilters {
  q: string
  status: string      // '' | 'active' | 'archived' | 'all'
  sourceType: string  // '' | 'manual' | 'imported'
  isFavorite: boolean | null
  tagId: string
}

interface DocumentListProps {
  documents: DocumentSummary[]
  selectedId: string | null
  filters: DocumentFilters
  tags: Tag[]
  onSelect: (id: string) => void
  onNew: () => void
  onFiltersChange: (f: DocumentFilters) => void
  loading: boolean
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric',
  })
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function buildListParams(f: DocumentFilters, limit = 100): ListDocumentsParams {
  const p: ListDocumentsParams = { limit }
  if (f.q) p.q = f.q
  if (f.status) p.status = f.status
  if (f.sourceType) p.source_type = f.sourceType
  if (f.isFavorite != null) p.is_favorite = f.isFavorite
  if (f.tagId) p.tag_id = f.tagId
  return p
}

export function DocumentList({
  documents, selectedId, filters, tags, onSelect, onNew, onFiltersChange, loading,
}: DocumentListProps) {
  const [showFilters, setShowFilters] = useState(false)
  const [query, setQuery] = useState(filters.q)

  useEffect(() => {
    setQuery(filters.q)
  }, [filters.q])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (query !== filters.q) {
        onFiltersChange({ ...filters, q: query })
      }
    }, 220)
    return () => window.clearTimeout(timer)
  }, [filters, onFiltersChange, query])

  function setFilter<K extends keyof DocumentFilters>(key: K, value: DocumentFilters[K]) {
    onFiltersChange({ ...filters, [key]: value })
  }

  const tagNameById = useMemo(() => new Map(tags.map(tag => [tag.id, tag.name])), [tags])
  const activeFilters = [
    filters.status && filters.status !== 'active' ? { key: 'status' as const, label: filters.status === 'archived' ? 'Archived' : 'All statuses' } : null,
    filters.sourceType ? { key: 'sourceType' as const, label: filters.sourceType === 'manual' ? 'Manual' : 'Imported' } : null,
    filters.isFavorite ? { key: 'isFavorite' as const, label: 'Favorites' } : null,
    filters.tagId ? { key: 'tagId' as const, label: tagNameById.get(filters.tagId) ?? 'Tag' } : null,
  ].filter(Boolean) as Array<{ key: keyof DocumentFilters; label: string }>

  function clearFilter(key: keyof DocumentFilters) {
    if (key === 'isFavorite') setFilter(key, null)
    else setFilter(key, '')
  }

  function clearFilters() {
    onFiltersChange({ q: filters.q, status: '', sourceType: '', isFavorite: null, tagId: '' })
  }

  return (
    <div className="flex h-full flex-col bg-surface">
      <div className="border-b border-border p-3">
        <div className="flex gap-2">
          <Button onClick={onNew} variant="primary" icon="plus" className="flex-1">
            New
          </Button>
          <IconButton
            icon="filter"
            label="Filters"
            active={showFilters || activeFilters.length > 0}
            onClick={() => setShowFilters(v => !v)}
          />
        </div>
        <div className="mt-3">
          <div className="relative">
            <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <TextInput
              type="text"
              placeholder="Search documents"
              value={query}
              onChange={e => setQuery(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>
        {activeFilters.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {activeFilters.map(filter => (
              <button
                key={filter.key}
                type="button"
                onClick={() => clearFilter(filter.key)}
                className="inline-flex items-center gap-1 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5 text-[11px] font-medium text-primary transition-colors hover:bg-primary/15"
              >
                {filter.label}
                <Icon name="close" className="h-3 w-3" />
              </button>
            ))}
            <button
              type="button"
              onClick={clearFilters}
              className="rounded-full px-2 py-0.5 text-[11px] font-medium text-muted-foreground hover:text-foreground"
            >
              Clear
            </button>
          </div>
        )}
      </div>

      {showFilters && (
        <div className="border-b border-border bg-background/60 px-3 py-3 text-xs">
          <div className="space-y-3">
            <div>
              <div className="mb-1.5 text-[11px] font-medium uppercase tracking-[0.08em] text-muted-foreground">Status</div>
              <div className="grid grid-cols-3 gap-1">
                {(['active', 'archived', 'all'] as const).map(status => (
                  <button
                    key={status}
                    type="button"
                    onClick={() => setFilter('status', filters.status === status ? '' : status)}
                    className={cn(
                      'h-8 rounded-md border text-xs font-medium transition-colors',
                      (filters.status === status) || (status === 'active' && !filters.status)
                        ? 'border-primary bg-primary text-primary-foreground'
                        : 'border-border bg-surface text-muted-foreground hover:bg-accent hover:text-foreground',
                    )}
                  >
                    {status === 'active' ? 'Active' : status === 'archived' ? 'Archive' : 'All'}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <div className="mb-1.5 text-[11px] font-medium uppercase tracking-[0.08em] text-muted-foreground">Source</div>
              <div className="grid grid-cols-2 gap-1">
                {(['manual', 'imported'] as const).map(source => (
                  <button
                    key={source}
                    type="button"
                    onClick={() => setFilter('sourceType', filters.sourceType === source ? '' : source)}
                    className={cn(
                      'h-8 rounded-md border text-xs font-medium transition-colors',
                      filters.sourceType === source
                        ? 'border-primary bg-primary text-primary-foreground'
                        : 'border-border bg-surface text-muted-foreground hover:bg-accent hover:text-foreground',
                    )}
                  >
                    {source === 'manual' ? 'Manual' : 'Imported'}
                  </button>
                ))}
              </div>
            </div>

            <label className="flex h-8 cursor-pointer items-center gap-2 rounded-md border border-border bg-surface px-2 text-foreground">
              <input
                type="checkbox"
                checked={filters.isFavorite === true}
                onChange={e => setFilter('isFavorite', e.target.checked ? true : null)}
                className="accent-primary"
              />
              <span>Favorites only</span>
            </label>

            {tags.length > 0 && (
              <div>
                <div className="mb-1.5 text-[11px] font-medium uppercase tracking-[0.08em] text-muted-foreground">Tag</div>
                <SelectInput
                  value={filters.tagId}
                  onChange={e => setFilter('tagId', e.target.value)}
                  className="h-8 w-full text-xs"
                >
                  <option value="">All tags</option>
                  {tags.map(tag => (
                    <option key={tag.id} value={tag.id}>{tag.name}</option>
                  ))}
                </SelectInput>
              </div>
            )}
          </div>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-y-auto">
        {loading && <SkeletonRows count={7} />}
        {!loading && documents.length === 0 && (
          <EmptyState
            compact
            icon={filters.q || activeFilters.length > 0 ? 'search' : 'file'}
            title={filters.q || activeFilters.length > 0 ? 'No matching documents' : 'No documents yet'}
            description={filters.q || activeFilters.length > 0 ? 'Adjust the search or clear filters to widen the list.' : 'Create a note or import a markdown folder to start building the vault.'}
            action={filters.q || activeFilters.length > 0 ? (
              <Button size="sm" variant="secondary" onClick={() => { setQuery(''); onFiltersChange({ q: '', status: '', sourceType: '', isFavorite: null, tagId: '' }) }}>
                Reset list
              </Button>
            ) : (
              <Button size="sm" variant="primary" icon="plus" onClick={onNew}>
                New document
              </Button>
            )}
          />
        )}
        {!loading && documents.map(doc => (
          <button
            key={doc.id}
            type="button"
            onClick={() => onSelect(doc.id)}
            className={cn(
              'group w-full border-b border-border px-3 py-3 text-left transition-colors hover:bg-accent/70',
              selectedId === doc.id && 'bg-primary/10 hover:bg-primary/10',
            )}
          >
            <div className="flex items-start gap-2">
              <div className={cn(
                'mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground',
                selectedId === doc.id && 'border-primary/25 bg-primary/10 text-primary',
              )}>
                <Icon name={doc.source_type === 'imported' ? 'import' : 'file'} className="h-3.5 w-3.5" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-start gap-1.5">
                  <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">
                    {doc.title || 'Untitled'}
                  </span>
                  {doc.is_favorite && <Icon name="star" className="h-3.5 w-3.5 shrink-0 text-amber-500" />}
                </div>
                {doc.summary && (
                  <div className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                    {doc.summary}
                  </div>
                )}
                <div className="mt-1.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
                  <span>{formatDate(doc.updated_at)}</span>
                  <span className="h-1 w-1 rounded-full bg-muted-foreground/35" />
                  <span className="font-mono">v{doc.version}</span>
                  <span className="h-1 w-1 rounded-full bg-muted-foreground/35" />
                  <span>{formatSize(doc.size_bytes)}</span>
                  {doc.source_type === 'imported' && <Badge tone="accent">Imported</Badge>}
                </div>
              </div>
            </div>
          </button>
        ))}
      </div>

      <div className="border-t border-border px-3 py-2 text-[11px] text-muted-foreground">
        {documents.length.toLocaleString()} document{documents.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
