import { useState } from 'react'
import type { Tag } from '../api/tags'
import type { DocumentSummary, ListDocumentsParams } from '../api/documents'

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
    month: 'short', day: 'numeric', year: 'numeric',
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

  function setFilter<K extends keyof DocumentFilters>(key: K, value: DocumentFilters[K]) {
    onFiltersChange({ ...filters, [key]: value })
  }

  const activeFilterCount = [
    filters.status && filters.status !== 'active',
    filters.sourceType,
    filters.isFavorite != null,
    filters.tagId,
  ].filter(Boolean).length

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-border flex gap-2">
        <button
          onClick={onNew}
          className="flex-1 py-1.5 px-3 bg-primary text-primary-foreground text-sm font-medium rounded-md hover:opacity-90 transition-opacity"
        >
          + New
        </button>
        <button
          onClick={() => setShowFilters(v => !v)}
          className={`px-2.5 py-1.5 text-sm rounded-md border transition-colors ${
            showFilters || activeFilterCount > 0
              ? 'border-primary bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-accent'
          }`}
          title="Filters"
        >
          ⚙{activeFilterCount > 0 ? ` ${activeFilterCount}` : ''}
        </button>
      </div>

      {/* Search */}
      <div className="px-3 py-2 border-b border-border">
        <input
          type="text"
          placeholder="Search documents…"
          value={filters.q}
          onChange={e => setFilter('q', e.target.value)}
          className="w-full px-3 py-1.5 text-sm border border-border rounded-md bg-background focus:outline-none focus:ring-1 focus:ring-primary/50"
        />
      </div>

      {/* Filter panel */}
      {showFilters && (
        <div className="px-3 py-2 border-b border-border space-y-2 text-xs bg-accent/30">
          {/* Status */}
          <div>
            <div className="text-muted-foreground mb-1">Status</div>
            <div className="flex gap-1 flex-wrap">
              {(['active', 'archived', 'all'] as const).map(s => (
                <button
                  key={s}
                  onClick={() => setFilter('status', filters.status === s ? '' : s)}
                  className={`px-2 py-0.5 rounded-full border text-xs transition-colors ${
                    (filters.status === s) || (s === 'active' && !filters.status)
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'border-border text-foreground hover:bg-accent'
                  }`}
                >
                  {s === 'active' ? 'Active' : s === 'archived' ? 'Archived' : 'All'}
                </button>
              ))}
            </div>
          </div>

          {/* Source */}
          <div>
            <div className="text-muted-foreground mb-1">Source</div>
            <div className="flex gap-1">
              {(['manual', 'imported'] as const).map(s => (
                <button
                  key={s}
                  onClick={() => setFilter('sourceType', filters.sourceType === s ? '' : s)}
                  className={`px-2 py-0.5 rounded-full border text-xs transition-colors ${
                    filters.sourceType === s
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'border-border text-foreground hover:bg-accent'
                  }`}
                >
                  {s === 'manual' ? 'Manual' : 'Imported'}
                </button>
              ))}
            </div>
          </div>

          {/* Favorites */}
          <div>
            <label className="flex items-center gap-1.5 cursor-pointer">
              <input
                type="checkbox"
                checked={filters.isFavorite === true}
                onChange={e => setFilter('isFavorite', e.target.checked ? true : null)}
                className="accent-primary"
              />
              <span className="text-foreground">Favorites only</span>
            </label>
          </div>

          {/* Tag filter */}
          {tags.length > 0 && (
            <div>
              <div className="text-muted-foreground mb-1">Tag</div>
              <select
                value={filters.tagId}
                onChange={e => setFilter('tagId', e.target.value)}
                className="w-full px-2 py-1 text-xs border border-border rounded bg-background"
              >
                <option value="">All tags</option>
                {tags.map(t => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
            </div>
          )}

          {activeFilterCount > 0 && (
            <button
              onClick={() => onFiltersChange({ q: filters.q, status: '', sourceType: '', isFavorite: null, tagId: '' })}
              className="text-xs text-muted-foreground underline"
            >
              Clear filters
            </button>
          )}
        </div>
      )}

      {/* List */}
      <div className="flex-1 overflow-y-auto">
        {loading && (
          <div className="p-4 text-sm text-muted-foreground text-center">Loading…</div>
        )}
        {!loading && documents.length === 0 && (
          <div className="p-4 text-sm text-muted-foreground text-center">
            {filters.q ? 'No matching documents' : 'No documents yet'}
          </div>
        )}
        {documents.map(doc => (
          <button
            key={doc.id}
            onClick={() => onSelect(doc.id)}
            className={`w-full text-left px-3 py-2.5 border-b border-border hover:bg-accent transition-colors ${
              selectedId === doc.id ? 'bg-accent' : ''
            } ${doc.source_type === 'archived' ? 'opacity-60' : ''}`}
          >
            <div className="flex items-start justify-between gap-1">
              <span className="text-sm font-medium truncate text-foreground flex-1 min-w-0">
                {doc.title}
              </span>
              {doc.is_favorite && (
                <span className="text-amber-400 text-xs shrink-0">★</span>
              )}
            </div>
            <div className="flex items-center gap-1.5 mt-0.5 text-xs text-muted-foreground flex-wrap">
              <span>{formatDate(doc.updated_at)}</span>
              <span>·</span>
              <span>v{doc.version}</span>
              <span>·</span>
              <span>{formatSize(doc.size_bytes)}</span>
              {doc.source_type === 'imported' && (
                <>
                  <span>·</span>
                  <span className="text-blue-500">↑</span>
                </>
              )}
            </div>
          </button>
        ))}
      </div>

      {/* Footer */}
      <div className="p-2 border-t border-border text-xs text-muted-foreground text-center">
        {documents.length} document{documents.length !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
