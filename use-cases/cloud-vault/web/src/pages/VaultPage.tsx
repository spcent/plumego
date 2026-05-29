import { useCallback, useEffect, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { DocumentList, buildListParams, type DocumentFilters } from '../components/DocumentList'
import { MarkdownEditor } from '../editor/MarkdownEditor'
import { APIResponseError } from '../api/client'
import { type DocumentDetail, type DocumentSummary, documentsAPI } from '../api/documents'
import { type Tag, tagsAPI } from '../api/tags'

const DEFAULT_CONTENT = '# Untitled\n\nStart writing…\n'

type SaveStatus = 'saved' | 'unsaved' | 'saving' | 'conflict' | 'error'

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const defaultFilters: DocumentFilters = {
  q: '', status: '', sourceType: '', isFavorite: null, tagId: '',
}

interface VaultPageProps {
  initialDocId?: string | null
  highlightQuery?: string
  onDocumentOpened?: () => void
}

export default function VaultPage({ initialDocId, highlightQuery = '', onDocumentOpened }: VaultPageProps) {
  const [documents, setDocuments] = useState<DocumentSummary[]>([])
  const [listLoading, setListLoading] = useState(true)
  const [filters, setFilters] = useState<DocumentFilters>(defaultFilters)
  const [allTags, setAllTags] = useState<Tag[]>([])
  const [docTags, setDocTags] = useState<Tag[]>([])
  const [showTagEditor, setShowTagEditor] = useState(false)
  const [newTagName, setNewTagName] = useState('')

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [currentVersion, setCurrentVersion] = useState(0)
  const [meta, setMeta] = useState<Partial<DocumentDetail>>({})
  const [activeQuery, setActiveQuery] = useState('')

  const [saveStatus, setSaveStatus] = useState<SaveStatus>('saved')
  const [statusMessage, setStatusMessage] = useState('')

  const originalTitle = useRef('')
  const originalContent = useRef('')

  // Open document from search when initialDocId prop changes.
  useEffect(() => {
    if (initialDocId) {
      setActiveQuery(highlightQuery)
      handleSelectDocument(initialDocId)
      onDocumentOpened?.()
    }
  }, [initialDocId])

  useEffect(() => {
    loadList()
    loadAllTags()
  }, [])

  useEffect(() => {
    loadList()
  }, [filters])

  async function loadList() {
    setListLoading(true)
    try {
      const result = await documentsAPI.list(buildListParams(filters))
      setDocuments(result.items)
    } catch {
      // ignore
    } finally {
      setListLoading(false)
    }
  }

  async function loadAllTags() {
    try {
      const result = await tagsAPI.list()
      setAllTags(result.items)
    } catch {
      // ignore
    }
  }

  async function loadDocTags(docId: string) {
    try {
      const result = await tagsAPI.getDocumentTags(docId)
      setDocTags(result.items)
    } catch {
      setDocTags([])
    }
  }

  function handleTitleChange(value: string) {
    setTitle(value)
    const isDirty = value !== originalTitle.current || content !== originalContent.current
    setSaveStatus(isDirty ? 'unsaved' : 'saved')
  }

  function handleContentChange(value: string) {
    setContent(value)
    const isDirty = title !== originalTitle.current || value !== originalContent.current
    setSaveStatus(isDirty ? 'unsaved' : 'saved')
  }

  async function handleNew() {
    setSelectedId(null)
    setTitle('Untitled')
    setContent(DEFAULT_CONTENT)
    setCurrentVersion(0)
    setMeta({})
    setDocTags([])
    originalTitle.current = ''
    originalContent.current = ''
    setSaveStatus('unsaved')
    setStatusMessage('')
  }

  async function handleSelectDocument(id: string) {
    setSelectedId(id)
    setSaveStatus('saving')
    setStatusMessage('')
    try {
      const doc = await documentsAPI.get(id)
      await loadDocTags(id)
      setTitle(doc.title)
      setContent(doc.content)
      setCurrentVersion(doc.version)
      setMeta(doc)
      originalTitle.current = doc.title
      originalContent.current = doc.content
      setSaveStatus('saved')
    } catch (err) {
      setSaveStatus('error')
      setStatusMessage(err instanceof Error ? err.message : 'Failed to load document')
    }
  }

  const handleSave = useCallback(async () => {
    if (saveStatus === 'saving') return
    setSaveStatus('saving')
    setStatusMessage('')
    try {
      if (!selectedId) {
        const result = await documentsAPI.create({ title: title || 'Untitled', content })
        setSelectedId(result.id)
        setCurrentVersion(result.version)
        originalTitle.current = result.title
        originalContent.current = content
        setSaveStatus('saved')
        setStatusMessage(`Created v${result.version}`)
        await loadList()
      } else {
        const result = await documentsAPI.update(selectedId, {
          title: title || 'Untitled', content, base_version: currentVersion,
        })
        setCurrentVersion(result.version)
        setMeta(prev => ({ ...prev, version: result.version, updated_at: result.updated_at }))
        originalTitle.current = result.title
        originalContent.current = content
        setSaveStatus('saved')
        setStatusMessage(result.changed ? `Saved v${result.version}` : 'No changes')
        await loadList()
      }
    } catch (err) {
      if (err instanceof APIResponseError && err.code === 'DOCUMENT_VERSION_CONFLICT') {
        setSaveStatus('conflict')
        setStatusMessage('Conflict: document updated elsewhere. Reload to see changes.')
      } else {
        setSaveStatus('error')
        setStatusMessage(err instanceof Error ? err.message : 'Save failed')
      }
    }
  }, [selectedId, title, content, currentVersion, saveStatus])

  async function handleToggleFavorite() {
    if (!selectedId) return
    const newVal = !meta.is_favorite
    try {
      await documentsAPI.updateFavorite(selectedId, newVal)
      setMeta(prev => ({ ...prev, is_favorite: newVal }))
      setDocuments(prev => prev.map(d => d.id === selectedId ? { ...d, is_favorite: newVal } : d))
    } catch { /* ignore */ }
  }

  async function handleArchive() {
    if (!selectedId) return
    const isArchived = meta.status === 'archived'
    const newStatus = isArchived ? 'active' : 'archived'
    try {
      await documentsAPI.updateStatus(selectedId, newStatus)
      setMeta(prev => ({ ...prev, status: newStatus }))
      await loadList()
    } catch { /* ignore */ }
  }

  async function handleDelete() {
    if (!selectedId || !window.confirm('Permanently delete this document?')) return
    try {
      await documentsAPI.delete(selectedId)
      setSelectedId(null)
      setTitle('')
      setContent('')
      setMeta({})
      setDocTags([])
      await loadList()
    } catch { /* ignore */ }
  }

  async function handleAddTag(tagId: string) {
    if (!selectedId || !tagId) return
    const ids = [...docTags.map(t => t.id), tagId]
    try {
      await tagsAPI.setDocumentTags(selectedId, ids)
      await loadDocTags(selectedId)
    } catch { /* ignore */ }
  }

  async function handleRemoveTag(tagId: string) {
    if (!selectedId) return
    try {
      await tagsAPI.removeDocumentTag(selectedId, tagId)
      setDocTags(prev => prev.filter(t => t.id !== tagId))
    } catch { /* ignore */ }
  }

  async function handleCreateTag() {
    const name = newTagName.trim()
    if (!name) return
    try {
      const tag = await tagsAPI.create(name)
      setAllTags(prev => [...prev, tag])
      setNewTagName('')
      if (selectedId) await handleAddTag(tag.id)
    } catch { /* ignore */ }
  }

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault()
        handleSave()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [handleSave])

  const statusColor: Record<SaveStatus, string> = {
    saved: 'text-green-600', unsaved: 'text-amber-500',
    saving: 'text-blue-500', conflict: 'text-red-600', error: 'text-red-600',
  }
  const statusLabel: Record<SaveStatus, string> = {
    saved: 'Saved', unsaved: 'Unsaved', saving: 'Saving…', conflict: 'Conflict', error: 'Error',
  }

  const isArchived = meta.status === 'archived'
  const availableTagsToAdd = allTags.filter(t => !docTags.find(dt => dt.id === t.id))

  return (
    <div className="h-full flex flex-col overflow-hidden bg-background">
      {/* Top bar */}
      <header className="h-12 flex items-center px-4 border-b border-border gap-3 shrink-0">
        {selectedId && (
          <>
            <button
              onClick={handleToggleFavorite}
              title={meta.is_favorite ? 'Remove from favorites' : 'Add to favorites'}
              className={`text-lg leading-none transition-colors ${meta.is_favorite ? 'text-amber-400' : 'text-muted-foreground hover:text-amber-400'}`}
            >
              {meta.is_favorite ? '★' : '☆'}
            </button>
            <button
              onClick={handleArchive}
              className="text-xs text-muted-foreground hover:text-foreground transition-colors border border-border rounded px-2 py-0.5"
            >
              {isArchived ? 'Restore' : 'Archive'}
            </button>
            <button
              onClick={handleDelete}
              className="text-xs text-red-500 hover:text-red-600 transition-colors"
            >
              Delete
            </button>
          </>
        )}
        <div className="flex-1" />
        <span className={`text-xs font-medium ${statusColor[saveStatus]}`}>
          {statusLabel[saveStatus]}{statusMessage ? ` — ${statusMessage}` : ''}
        </span>
        <button
          onClick={handleSave}
          disabled={saveStatus === 'saving' || saveStatus === 'saved'}
          className="px-3 py-1 text-xs font-medium bg-primary text-primary-foreground rounded hover:opacity-90 disabled:opacity-50 transition-opacity"
        >
          Save
        </button>
      </header>

      {/* Main three-panel layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left panel: document list */}
        <aside className="w-64 shrink-0 border-r border-border overflow-hidden">
          <DocumentList
            documents={documents}
            selectedId={selectedId}
            filters={filters}
            tags={allTags}
            onSelect={handleSelectDocument}
            onNew={handleNew}
            onFiltersChange={setFilters}
            loading={listLoading}
          />
        </aside>

        {/* Center panel: editor */}
        <main className="flex-1 flex flex-col overflow-hidden border-r border-border min-w-0">
          <div className="px-4 py-2 border-b border-border shrink-0 flex items-center gap-2">
            <input
              type="text"
              value={title}
              onChange={e => handleTitleChange(e.target.value)}
              placeholder="Document title…"
              className="flex-1 text-lg font-semibold bg-transparent focus:outline-none text-foreground placeholder:text-muted-foreground"
            />
            {meta.source_type === 'imported' && (
              <span className="text-xs text-blue-500 border border-blue-200 rounded px-1.5 py-0.5 shrink-0">
                imported
              </span>
            )}
          </div>
          <div className="flex-1 overflow-hidden">
            <MarkdownEditor
              value={content}
              onChange={handleContentChange}
              onSave={handleSave}
              highlightQuery={activeQuery}
            />
          </div>
        </main>

        {/* Right panel: preview + tags + meta */}
        <aside className="w-80 shrink-0 flex flex-col overflow-hidden">
          {/* Preview */}
          <div className="flex-1 overflow-y-auto p-4">
            <div className="prose text-sm">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
            </div>
          </div>

          {/* Tags */}
          {selectedId && (
            <div className="shrink-0 border-t border-border p-3">
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-xs font-medium text-foreground">Tags</span>
                <button
                  onClick={() => setShowTagEditor(v => !v)}
                  className="text-xs text-primary hover:opacity-80"
                >
                  {showTagEditor ? 'Done' : '+ Add'}
                </button>
              </div>
              <div className="flex flex-wrap gap-1 min-h-[20px]">
                {docTags.map(tag => (
                  <span
                    key={tag.id}
                    className="inline-flex items-center gap-1 px-2 py-0.5 text-xs rounded-full border border-border bg-accent"
                    style={tag.color ? { borderColor: tag.color, color: tag.color } : undefined}
                  >
                    {tag.name}
                    <button onClick={() => handleRemoveTag(tag.id)} className="hover:text-red-500 leading-none">×</button>
                  </span>
                ))}
                {docTags.length === 0 && !showTagEditor && (
                  <span className="text-xs text-muted-foreground">No tags</span>
                )}
              </div>
              {showTagEditor && (
                <div className="mt-2 space-y-1">
                  {availableTagsToAdd.length > 0 && (
                    <select
                      defaultValue=""
                      onChange={e => { handleAddTag(e.target.value); e.target.value = '' }}
                      className="w-full px-2 py-1 text-xs border border-border rounded bg-background"
                    >
                      <option value="" disabled>Select tag…</option>
                      {availableTagsToAdd.map(t => (
                        <option key={t.id} value={t.id}>{t.name}</option>
                      ))}
                    </select>
                  )}
                  <div className="flex gap-1">
                    <input
                      type="text"
                      value={newTagName}
                      onChange={e => setNewTagName(e.target.value)}
                      onKeyDown={e => { if (e.key === 'Enter') handleCreateTag() }}
                      placeholder="New tag…"
                      className="flex-1 px-2 py-1 text-xs border border-border rounded bg-background focus:outline-none"
                    />
                    <button
                      onClick={handleCreateTag}
                      disabled={!newTagName.trim()}
                      className="px-2 py-1 text-xs bg-primary text-primary-foreground rounded disabled:opacity-40"
                    >
                      Create
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Meta info */}
          {selectedId && meta.updated_at && (
            <div className="shrink-0 border-t border-border p-3 text-xs text-muted-foreground space-y-1">
              <div className="font-medium text-foreground mb-1">Document Info</div>
              <div className="flex justify-between">
                <span>Version</span>
                <span className="font-mono">{meta.version ?? currentVersion}</span>
              </div>
              <div className="flex justify-between">
                <span>Size</span>
                <span className="font-mono">{meta.size_bytes != null ? formatSize(meta.size_bytes) : '—'}</span>
              </div>
              <div className="flex justify-between">
                <span>Words</span>
                <span className="font-mono">{meta.word_count ?? '—'}</span>
              </div>
              <div className="flex justify-between">
                <span>Lines</span>
                <span className="font-mono">{meta.line_count ?? '—'}</span>
              </div>
              {meta.review_status && (
                <div className="flex justify-between">
                  <span>Review</span>
                  <span className={`font-mono ${meta.review_status === 'reviewed' ? 'text-green-600' : ''}`}>
                    {meta.review_status}
                  </span>
                </div>
              )}
              <div className="flex justify-between">
                <span>Updated</span>
                <span className="font-mono text-right">{formatDateTime(meta.updated_at)}</span>
              </div>
              {meta.original_path && (
                <div className="flex justify-between gap-2">
                  <span className="shrink-0">Source</span>
                  <span className="font-mono text-right truncate text-blue-500" title={meta.original_path}>
                    {meta.original_path.split('/').pop()}
                  </span>
                </div>
              )}
            </div>
          )}
        </aside>
      </div>
    </div>
  )
}
