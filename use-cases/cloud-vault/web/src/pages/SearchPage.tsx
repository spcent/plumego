import { useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import { searchAPI, type HistoryItem, type SearchParams, type SearchResultItem } from '../api/search'
import { type Tag, tagsAPI } from '../api/tags'
import { Badge, Button, EmptyState, IconButton, SelectInput, SkeletonRows, TextInput, cn } from '../components/ui'
import { Icon } from '../components/icons'

interface SearchPageProps {
  onOpenDocument: (id: string, query: string) => void
}

interface FilterChip {
  key: string
  label: string
  clear: () => void
}

function sanitizeSnippet(html: string) {
  return html
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/&#39;/g, "'")
    .replace(/&lt;mark&gt;/g, '<mark>')
    .replace(/&lt;\/mark&gt;/g, '</mark>')
}

function HighlightSnippet({ html }: { html: string }) {
  return (
    <span
      className="text-xs leading-5 text-foreground/80"
      dangerouslySetInnerHTML={{ __html: sanitizeSnippet(html) }}
    />
  )
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

function ResultRow({ item, query, active, onOpen }: {
  item: SearchResultItem
  query: string
  active: boolean
  onOpen: (id: string, query: string) => void
}) {
  return (
    <button
      type="button"
      className={cn(
        'group w-full border-b border-border px-5 py-4 text-left transition-colors hover:bg-accent/70',
        active && 'bg-primary/10 hover:bg-primary/10',
      )}
      onClick={() => onOpen(item.id, query)}
    >
      <div className="flex items-start gap-3">
        <div className={cn(
          'mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-border bg-surface text-muted-foreground',
          active && 'border-primary/25 bg-primary/10 text-primary',
        )}>
          <Icon name={item.source_type === 'imported' ? 'import' : 'file'} className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-start gap-2">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-1.5">
                <span className="truncate text-sm font-semibold text-foreground">{item.title || 'Untitled'}</span>
                {item.is_favorite && <Icon name="star" className="h-3.5 w-3.5 shrink-0 text-amber-500" />}
              </div>
              {item.original_path && (
                <div className="mt-0.5 truncate font-mono text-[11px] text-muted-foreground" title={item.original_path}>
                  {item.original_path}
                </div>
              )}
            </div>
            <div className="flex shrink-0 items-center gap-1.5">
              {item.source_type === 'imported' && <Badge tone="accent">Imported</Badge>}
              <span className="text-xs text-muted-foreground">{formatDate(item.updated_at)}</span>
            </div>
          </div>

          {item.highlights && item.highlights.length > 0 ? (
            <div className="mt-3 space-y-1.5">
              {item.highlights.map((highlight, idx) => (
                <div key={idx} className="rounded-md border border-amber-500/20 bg-amber-500/10 px-3 py-2">
                  <HighlightSnippet html={highlight} />
                </div>
              ))}
            </div>
          ) : item.summary ? (
            <div className="mt-2 line-clamp-2 text-xs leading-5 text-muted-foreground">{item.summary}</div>
          ) : null}

          {item.tags && item.tags.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-1.5">
              {item.tags.map(tag => (
                <Badge key={tag}>{tag}</Badge>
              ))}
            </div>
          )}
        </div>
      </div>
    </button>
  )
}

export default function SearchPage({ onOpenDocument }: SearchPageProps) {
  const [q, setQ] = useState('')
  const [inputVal, setInputVal] = useState('')
  const [results, setResults] = useState<SearchResultItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [offset, setOffset] = useState(0)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const limit = 20

  const [showFilters, setShowFilters] = useState(false)
  const [tagFilter, setTagFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sourceFilter, setSourceFilter] = useState('')
  const [favFilter, setFavFilter] = useState<boolean | null>(null)
  const [fromFilter, setFromFilter] = useState('')
  const [toFilter, setToFilter] = useState('')
  const [allTags, setAllTags] = useState<Tag[]>([])

  const [history, setHistory] = useState<HistoryItem[]>([])
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    tagsAPI.list().then(r => setAllTags(r.items)).catch(() => {})
    loadHistory()
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [results])

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (results.length === 0) return
      if (event.key === 'ArrowDown') {
        event.preventDefault()
        setSelectedIndex(index => Math.min(results.length - 1, index + 1))
      } else if (event.key === 'ArrowUp') {
        event.preventDefault()
        setSelectedIndex(index => Math.max(0, index - 1))
      } else if (event.key === 'Enter' && document.activeElement !== inputRef.current) {
        event.preventDefault()
        const item = results[selectedIndex]
        if (item) onOpenDocument(item.id, q)
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [onOpenDocument, q, results, selectedIndex])

  async function loadHistory() {
    try {
      const resp = await searchAPI.getHistory()
      setHistory(resp.items)
    } catch { /* ignore */ }
  }

  async function doSearch(newQ: string, newOffset = 0) {
    setLoading(true)
    const params: SearchParams = {
      q: newQ,
      tag: tagFilter || undefined,
      status: statusFilter || undefined,
      source_type: sourceFilter || undefined,
      is_favorite: favFilter ?? undefined,
      from: fromFilter || undefined,
      to: toFilter || undefined,
      limit,
      offset: newOffset,
    }
    try {
      const result = await searchAPI.search(params)
      setResults(result.items)
      setTotal(result.total)
      setOffset(newOffset)
    } catch {
      setResults([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const trimmed = inputVal.trim()
    setQ(trimmed)
    doSearch(trimmed, 0)
    loadHistory()
  }

  function handleHistoryClick(item: HistoryItem) {
    setInputVal(item.query)
    setQ(item.query)
    doSearch(item.query, 0)
  }

  async function handleClearHistory() {
    await searchAPI.clearHistory()
    setHistory([])
  }

  function clearFilters() {
    setTagFilter('')
    setStatusFilter('')
    setSourceFilter('')
    setFavFilter(null)
    setFromFilter('')
    setToFilter('')
  }

  const tagNameById = useMemo(() => new Map(allTags.map(tag => [tag.id, tag.name])), [allTags])
  const activeFilters: FilterChip[] = [
    tagFilter ? { key: 'tag', label: tagNameById.get(tagFilter) ?? 'Tag', clear: () => setTagFilter('') } : null,
    statusFilter ? { key: 'status', label: statusFilter === 'active' ? 'Active' : statusFilter === 'archived' ? 'Archived' : 'All statuses', clear: () => setStatusFilter('') } : null,
    sourceFilter ? { key: 'source', label: sourceFilter === 'manual' ? 'Manual' : 'Imported', clear: () => setSourceFilter('') } : null,
    favFilter ? { key: 'favorite', label: 'Favorites', clear: () => setFavFilter(null) } : null,
    fromFilter ? { key: 'from', label: `From ${fromFilter}`, clear: () => setFromFilter('') } : null,
    toFilter ? { key: 'to', label: `To ${toFilter}`, clear: () => setToFilter('') } : null,
  ].filter(Boolean) as FilterChip[]

  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit)
  const hasSearched = q !== '' || loading

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background">
      <header className="shrink-0 border-b border-border bg-surface px-4 py-4">
        <form onSubmit={handleSubmit} className="mx-auto flex max-w-4xl gap-2">
          <div className="relative min-w-0 flex-1">
            <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <TextInput
              ref={inputRef}
              type="text"
              value={inputVal}
              onChange={e => setInputVal(e.target.value)}
              placeholder="Search documents"
              className="h-10 pl-9"
              autoFocus
            />
          </div>
          <Button type="submit" variant="primary" icon="search" className="h-10">
            Search
          </Button>
          <IconButton
            icon="filter"
            label="Filters"
            active={showFilters || activeFilters.length > 0}
            onClick={() => setShowFilters(v => !v)}
            className="h-10 w-10"
          />
        </form>

        {(showFilters || activeFilters.length > 0) && (
          <div className="mx-auto mt-3 max-w-4xl">
            {activeFilters.length > 0 && (
              <div className="mb-3 flex flex-wrap gap-1.5">
                {activeFilters.map(filter => (
                  <button
                    key={filter.key}
                    type="button"
                    onClick={filter.clear}
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
                  Clear filters
                </button>
              </div>
            )}

            {showFilters && (
              <div className="grid gap-2 rounded-lg border border-border bg-background/70 p-3 text-xs sm:grid-cols-2 lg:grid-cols-6">
                <SelectInput value={tagFilter} onChange={e => setTagFilter(e.target.value)} className="h-8 text-xs lg:col-span-2">
                  <option value="">All tags</option>
                  {allTags.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
                </SelectInput>
                <SelectInput value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="h-8 text-xs">
                  <option value="">Any status</option>
                  <option value="active">Active</option>
                  <option value="archived">Archived</option>
                  <option value="all">All</option>
                </SelectInput>
                <SelectInput value={sourceFilter} onChange={e => setSourceFilter(e.target.value)} className="h-8 text-xs">
                  <option value="">Any source</option>
                  <option value="manual">Manual</option>
                  <option value="imported">Imported</option>
                </SelectInput>
                <TextInput type="date" value={fromFilter} onChange={e => setFromFilter(e.target.value)} className="h-8 text-xs" title="From date" />
                <TextInput type="date" value={toFilter} onChange={e => setToFilter(e.target.value)} className="h-8 text-xs" title="To date" />
                <label className="flex h-8 items-center gap-2 rounded-md border border-border bg-surface px-2 text-muted-foreground lg:col-span-2">
                  <input
                    type="checkbox"
                    checked={favFilter === true}
                    onChange={e => setFavFilter(e.target.checked ? true : null)}
                    className="accent-primary"
                  />
                  Favorites only
                </label>
              </div>
            )}
          </div>
        )}
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {hasSearched ? (
          <>
            {loading && <SkeletonRows count={6} />}
            {!loading && results.length === 0 && (
              <EmptyState
                icon="search"
                title="No results found"
                description="Try fewer filters, a shorter phrase, or search by filename and tag."
                action={activeFilters.length > 0 ? <Button size="sm" onClick={clearFilters}>Clear filters</Button> : undefined}
              />
            )}
            {!loading && results.length > 0 && (
              <>
                <div className="sticky top-0 z-10 border-b border-border bg-background/95 px-5 py-2 text-xs text-muted-foreground">
                  {total.toLocaleString()} result{total !== 1 ? 's' : ''} for <span className="font-medium text-foreground">{q || 'all documents'}</span>
                </div>
                {results.map((item, idx) => (
                  <ResultRow
                    key={item.id}
                    item={item}
                    query={q}
                    active={idx === selectedIndex}
                    onOpen={onOpenDocument}
                  />
                ))}
                {totalPages > 1 && (
                  <div className="flex items-center gap-2 border-t border-border px-5 py-4 text-sm">
                    <Button
                      icon="chevronLeft"
                      onClick={() => doSearch(q, Math.max(0, offset - limit))}
                      disabled={currentPage === 0}
                    >
                      Previous
                    </Button>
                    <span className="text-xs text-muted-foreground">
                      Page {currentPage + 1} of {totalPages}
                    </span>
                    <Button
                      icon="chevronRight"
                      onClick={() => doSearch(q, offset + limit)}
                      disabled={currentPage >= totalPages - 1}
                    >
                      Next
                    </Button>
                  </div>
                )}
              </>
            )}
          </>
        ) : (
          <div className="mx-auto max-w-3xl px-5 py-8">
            {history.length > 0 ? (
              <section className="rounded-lg border border-border bg-surface">
                <div className="flex items-center justify-between border-b border-border px-4 py-3">
                  <div>
                    <div className="text-sm font-semibold text-foreground">Recent searches</div>
                    <div className="text-xs text-muted-foreground">Reuse a query or clear local history.</div>
                  </div>
                  <Button size="sm" variant="ghost" onClick={handleClearHistory}>
                    Clear
                  </Button>
                </div>
                <div className="divide-y divide-border">
                  {history.slice(0, 10).map(item => (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => handleHistoryClick(item)}
                      className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors hover:bg-accent"
                    >
                      <span className="min-w-0 truncate text-sm font-medium text-foreground">{item.query}</span>
                      <span className="shrink-0 text-xs text-muted-foreground">{item.result_count} results</span>
                    </button>
                  ))}
                </div>
              </section>
            ) : (
              <EmptyState
                icon="search"
                title="Search your vault"
                description="Find documents by content, tags, source path, date range, and favorite state."
              />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
