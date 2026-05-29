import { useEffect, useRef, useState } from 'react'
import { searchAPI, type HistoryItem, type SearchParams, type SearchResultItem } from '../api/search'
import { type Tag, tagsAPI } from '../api/tags'

interface SearchPageProps {
  onOpenDocument: (id: string, query: string) => void
}

function HighlightBadge({ html }: { html: string }) {
  return (
    <span
      className="text-xs text-foreground/80 leading-relaxed"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

function ResultCard({ item, query, onOpen }: {
  item: SearchResultItem
  query: string
  onOpen: (id: string, query: string) => void
}) {
  return (
    <div
      className="px-4 py-3 border-b border-border hover:bg-accent/50 cursor-pointer transition-colors"
      onClick={() => onOpen(item.id, query)}
    >
      <div className="flex items-start justify-between gap-2 mb-1">
        <div className="flex-1 min-w-0">
          <span className="text-sm font-medium text-foreground">{item.title}</span>
          {item.is_favorite && <span className="ml-1.5 text-amber-400 text-xs">★</span>}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {item.source_type === 'imported' && (
            <span className="text-[10px] text-blue-500 border border-blue-200 rounded px-1">imported</span>
          )}
          <span className="text-xs text-muted-foreground">
            {new Date(item.updated_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
          </span>
        </div>
      </div>

      {item.original_path && (
        <div className="text-[11px] font-mono text-muted-foreground truncate mb-1" title={item.original_path}>
          {item.original_path}
        </div>
      )}

      {item.highlights && item.highlights.length > 0 ? (
        <div className="space-y-0.5 mb-1.5">
          {item.highlights.map((h, i) => (
            <div key={i} className="bg-yellow-50 border border-yellow-100 rounded px-2 py-1">
              <HighlightBadge html={h} />
            </div>
          ))}
        </div>
      ) : item.summary ? (
        <div className="text-xs text-muted-foreground line-clamp-2 mb-1.5">{item.summary}</div>
      ) : null}

      {item.tags && item.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {item.tags.map(tag => (
            <span key={tag} className="px-1.5 py-0.5 text-[10px] rounded-full border border-border bg-accent">
              {tag}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

export default function SearchPage({ onOpenDocument }: SearchPageProps) {
  const [q, setQ] = useState('')
  const [inputVal, setInputVal] = useState('')
  const [results, setResults] = useState<SearchResultItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [offset, setOffset] = useState(0)
  const limit = 20

  // Filters
  const [showFilters, setShowFilters] = useState(false)
  const [tagFilter, setTagFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sourceFilter, setSourceFilter] = useState('')
  const [favFilter, setFavFilter] = useState<boolean | null>(null)
  const [fromFilter, setFromFilter] = useState('')
  const [toFilter, setToFilter] = useState('')
  const [allTags, setAllTags] = useState<Tag[]>([])

  // History
  const [history, setHistory] = useState<HistoryItem[]>([])

  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    tagsAPI.list().then(r => setAllTags(r.items)).catch(() => {})
    loadHistory()
  }, [])

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

  function handleSubmit(e: React.FormEvent) {
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

  const totalPages = Math.ceil(total / limit)
  const currentPage = Math.floor(offset / limit)
  const hasSearched = q !== '' || loading

  return (
    <div className="h-full flex flex-col overflow-hidden bg-background">
      {/* Search header */}
      <header className="shrink-0 px-4 pt-4 pb-3 border-b border-border">
        <form onSubmit={handleSubmit} className="flex gap-2 max-w-2xl">
          <input
            ref={inputRef}
            type="text"
            value={inputVal}
            onChange={e => setInputVal(e.target.value)}
            placeholder="Search documents… (press Enter)"
            className="flex-1 px-3 py-2 text-sm border border-border rounded-md bg-background focus:outline-none focus:ring-1 focus:ring-primary/50"
            autoFocus
          />
          <button
            type="submit"
            className="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-md hover:opacity-90"
          >
            Search
          </button>
          <button
            type="button"
            onClick={() => setShowFilters(v => !v)}
            className={`px-3 py-2 text-sm rounded-md border transition-colors ${
              showFilters
                ? 'bg-primary/10 border-primary text-primary'
                : 'border-border text-muted-foreground hover:bg-accent'
            }`}
          >
            Filters
          </button>
        </form>

        {/* Filter panel */}
        {showFilters && (
          <div className="mt-3 flex flex-wrap gap-3 text-xs max-w-2xl">
            {/* Tag */}
            <select
              value={tagFilter}
              onChange={e => setTagFilter(e.target.value)}
              className="px-2 py-1 border border-border rounded bg-background text-xs"
            >
              <option value="">All tags</option>
              {allTags.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
            {/* Status */}
            <select
              value={statusFilter}
              onChange={e => setStatusFilter(e.target.value)}
              className="px-2 py-1 border border-border rounded bg-background text-xs"
            >
              <option value="">Any status</option>
              <option value="active">Active</option>
              <option value="archived">Archived</option>
              <option value="all">All</option>
            </select>
            {/* Source */}
            <select
              value={sourceFilter}
              onChange={e => setSourceFilter(e.target.value)}
              className="px-2 py-1 border border-border rounded bg-background text-xs"
            >
              <option value="">Any source</option>
              <option value="manual">Manual</option>
              <option value="imported">Imported</option>
            </select>
            {/* Favorites */}
            <label className="flex items-center gap-1 cursor-pointer">
              <input
                type="checkbox"
                checked={favFilter === true}
                onChange={e => setFavFilter(e.target.checked ? true : null)}
                className="accent-primary"
              />
              Favorites only
            </label>
            {/* Date range */}
            <input
              type="date"
              value={fromFilter}
              onChange={e => setFromFilter(e.target.value)}
              className="px-2 py-1 border border-border rounded bg-background text-xs"
              title="From date"
            />
            <input
              type="date"
              value={toFilter}
              onChange={e => setToFilter(e.target.value)}
              className="px-2 py-1 border border-border rounded bg-background text-xs"
              title="To date"
            />
            {(tagFilter || statusFilter || sourceFilter || favFilter || fromFilter || toFilter) && (
              <button
                type="button"
                onClick={() => { setTagFilter(''); setStatusFilter(''); setSourceFilter(''); setFavFilter(null); setFromFilter(''); setToFilter('') }}
                className="text-muted-foreground underline"
              >
                Clear filters
              </button>
            )}
          </div>
        )}
      </header>

      <div className="flex-1 overflow-y-auto">
        {/* Results */}
        {hasSearched && (
          <>
            {loading && (
              <div className="p-6 text-sm text-muted-foreground text-center">Searching…</div>
            )}
            {!loading && results.length === 0 && (
              <div className="p-6 text-sm text-muted-foreground text-center">No results found.</div>
            )}
            {!loading && results.length > 0 && (
              <>
                <div className="px-4 py-2 text-xs text-muted-foreground border-b border-border">
                  {total} result{total !== 1 ? 's' : ''}
                </div>
                {results.map(item => (
                  <ResultCard key={item.id} item={item} query={q} onOpen={onOpenDocument} />
                ))}
                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="px-4 py-3 flex items-center gap-2 text-sm border-t border-border">
                    <button
                      onClick={() => doSearch(q, Math.max(0, offset - limit))}
                      disabled={currentPage === 0}
                      className="px-3 py-1 border border-border rounded disabled:opacity-40 hover:bg-accent"
                    >
                      ← Prev
                    </button>
                    <span className="text-muted-foreground text-xs">
                      Page {currentPage + 1} / {totalPages}
                    </span>
                    <button
                      onClick={() => doSearch(q, offset + limit)}
                      disabled={currentPage >= totalPages - 1}
                      className="px-3 py-1 border border-border rounded disabled:opacity-40 hover:bg-accent"
                    >
                      Next →
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {/* History (shown when no search has been performed) */}
        {!hasSearched && history.length > 0 && (
          <div className="px-4 py-3">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-medium text-foreground">Recent searches</span>
              <button
                onClick={handleClearHistory}
                className="text-xs text-muted-foreground hover:text-foreground underline"
              >
                Clear
              </button>
            </div>
            <div className="space-y-1">
              {history.slice(0, 10).map(item => (
                <button
                  key={item.id}
                  onClick={() => handleHistoryClick(item)}
                  className="w-full text-left px-3 py-1.5 rounded hover:bg-accent text-sm flex items-center justify-between gap-2 group"
                >
                  <span className="truncate text-foreground">{item.query}</span>
                  <span className="text-xs text-muted-foreground shrink-0">
                    {item.result_count} results
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}

        {!hasSearched && history.length === 0 && (
          <div className="flex-1 flex items-center justify-center h-48">
            <div className="text-sm text-muted-foreground text-center">
              <div className="text-2xl mb-2">🔍</div>
              Enter a query above to search your documents.
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
