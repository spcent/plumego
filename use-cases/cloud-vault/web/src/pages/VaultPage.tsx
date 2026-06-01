import { useCallback, useEffect, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { DocumentList, buildListParams, type DocumentFilters } from '../components/DocumentList'
import { MarkdownEditor } from '../editor/MarkdownEditor'
import { APIResponseError } from '../api/client'
import { type DocumentDetail, type DocumentSummary, documentsAPI } from '../api/documents'
import { type Tag, tagsAPI } from '../api/tags'
import { enqueueSummary, enqueuePromptExtract } from '../api/ai'
import { Badge, Button, EmptyState, IconButton, PanelHeader, SelectInput, TextInput, cn } from '../components/ui'
import { Icon } from '../components/icons'

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
  const [showPreview, setShowPreview] = useState(true)
  const [showDetails, setShowDetails] = useState(true)
  const [showMore, setShowMore] = useState(false)
  const [deleteRequested, setDeleteRequested] = useState(false)

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [currentVersion, setCurrentVersion] = useState(0)
  const [meta, setMeta] = useState<Partial<DocumentDetail>>({})
  const [activeQuery, setActiveQuery] = useState('')

  const [saveStatus, setSaveStatus] = useState<SaveStatus>('saved')
  const [statusMessage, setStatusMessage] = useState('')
  const [actionMessage, setActionMessage] = useState('')

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
    setActionMessage('')
    setDeleteRequested(false)
    setShowMore(false)
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
      setActionMessage(newVal ? 'Added to favorites' : 'Removed from favorites')
    } catch { /* ignore */ }
  }

  async function handleArchive() {
    if (!selectedId) return
    const isArchived = meta.status === 'archived'
    const newStatus = isArchived ? 'active' : 'archived'
    try {
      await documentsAPI.updateStatus(selectedId, newStatus)
      setMeta(prev => ({ ...prev, status: newStatus }))
      setActionMessage(isArchived ? 'Document restored' : 'Document archived')
      setShowMore(false)
      await loadList()
    } catch { /* ignore */ }
  }

  async function handleDelete() {
    if (!selectedId) return
    if (!deleteRequested) {
      setDeleteRequested(true)
      return
    }
    try {
      await documentsAPI.delete(selectedId)
      setSelectedId(null)
      setTitle('')
      setContent('')
      setMeta({})
      setDocTags([])
      setDeleteRequested(false)
      setShowMore(false)
      setActionMessage('Document deleted')
      await loadList()
    } catch { /* ignore */ }
  }

  async function handleAddTag(tagId: string) {
    if (!selectedId || !tagId) return
    const ids = [...docTags.map(t => t.id), tagId]
    try {
      await tagsAPI.setDocumentTags(selectedId, ids)
      await loadDocTags(selectedId)
      setActionMessage('Tag added')
    } catch { /* ignore */ }
  }

  async function handleRemoveTag(tagId: string) {
    if (!selectedId) return
    try {
      await tagsAPI.removeDocumentTag(selectedId, tagId)
      setDocTags(prev => prev.filter(t => t.id !== tagId))
      setActionMessage('Tag removed')
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

  const statusLabel: Record<SaveStatus, string> = {
    saved: 'Saved', unsaved: 'Unsaved', saving: 'Saving…', conflict: 'Conflict', error: 'Error',
  }

  const isArchived = meta.status === 'archived'
  const availableTagsToAdd = allTags.filter(t => !docTags.find(dt => dt.id === t.id))
  const hasDocument = selectedId != null || saveStatus === 'unsaved'

  return (
    <div className="flex h-full flex-col overflow-hidden bg-background">
      <header className="flex min-h-12 shrink-0 items-center gap-2 border-b border-border bg-surface px-3 md:px-4">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          {selectedId && (
            <IconButton
              icon="star"
              label={meta.is_favorite ? 'Remove from favorites' : 'Add to favorites'}
              active={meta.is_favorite}
              tone="favorite"
              onClick={handleToggleFavorite}
            />
          )}
          <div className="min-w-0">
            <div className="truncate text-sm font-medium text-foreground">
              {hasDocument ? title || 'Untitled' : 'No document selected'}
            </div>
            <div className="truncate text-xs text-muted-foreground">
              {statusMessage || actionMessage || 'Use Command-S or Ctrl-S to save changes'}
            </div>
          </div>
        </div>

        <SaveStatusPill status={saveStatus} label={statusLabel[saveStatus]} />
        <Button
          onClick={handleSave}
          disabled={saveStatus === 'saving' || saveStatus === 'saved'}
          variant="primary"
          size="sm"
          icon="save"
        >
          Save
        </Button>

        <div className="hidden items-center gap-1 lg:flex">
          <IconButton
            icon="layout"
            label={showPreview ? 'Hide preview' : 'Show preview'}
            active={showPreview}
            onClick={() => setShowPreview(v => !v)}
          />
          <IconButton
            icon="tag"
            label={showDetails ? 'Hide details' : 'Show details'}
            active={showDetails}
            onClick={() => setShowDetails(v => !v)}
          />
        </div>

        {selectedId && (
          <div className="relative">
            <IconButton icon="more" label="More actions" active={showMore} onClick={() => setShowMore(v => !v)} />
            {showMore && (
              <div className="absolute right-0 top-10 z-20 w-52 overflow-hidden rounded-lg border border-border bg-surface shadow-lg shadow-slate-950/10">
                <button
                  type="button"
                  onClick={() => { enqueueSummary(selectedId).then(() => setActionMessage('Summary task queued')).catch(() => setActionMessage('Failed to queue summary')); setShowMore(false) }}
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground hover:bg-accent"
                >
                  <Icon name="spark" className="h-4 w-4 text-muted-foreground" />
                  Summarize
                </button>
                <button
                  type="button"
                  onClick={() => { enqueuePromptExtract(selectedId).then(() => setActionMessage('Prompt extraction queued')).catch(() => setActionMessage('Failed to queue extraction')); setShowMore(false) }}
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground hover:bg-accent"
                >
                  <Icon name="bolt" className="h-4 w-4 text-muted-foreground" />
                  Extract prompt
                </button>
                <button
                  type="button"
                  onClick={handleArchive}
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground hover:bg-accent"
                >
                  <Icon name="archive" className="h-4 w-4 text-muted-foreground" />
                  {isArchived ? 'Restore' : 'Archive'}
                </button>
                <div className="border-t border-border" />
                <button
                  type="button"
                  onClick={handleDelete}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-destructive/10',
                    deleteRequested ? 'text-destructive' : 'text-muted-foreground hover:text-destructive',
                  )}
                >
                  <Icon name="trash" className="h-4 w-4" />
                  {deleteRequested ? 'Confirm delete' : 'Delete'}
                </button>
              </div>
            )}
          </div>
        )}
      </header>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden md:flex-row">
        <aside className="h-72 shrink-0 overflow-hidden border-b border-border md:h-auto md:w-72 md:border-b-0 md:border-r">
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

        <main className="flex min-w-0 flex-1 flex-col overflow-hidden border-r border-border bg-surface">
          {hasDocument ? (
            <>
              <div className="shrink-0 border-b border-border px-4 py-3">
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={title}
                    onChange={e => handleTitleChange(e.target.value)}
                    placeholder="Document title"
                    className="min-w-0 flex-1 bg-transparent text-xl font-semibold tracking-tight text-foreground outline-none placeholder:text-muted-foreground"
                  />
                  {meta.source_type === 'imported' && <Badge tone="accent">Imported</Badge>}
                  {isArchived && <Badge tone="warning">Archived</Badge>}
                </div>
              </div>
              <div className="min-h-0 flex-1 overflow-hidden">
                <MarkdownEditor
                  value={content}
                  onChange={handleContentChange}
                  onSave={handleSave}
                  highlightQuery={activeQuery}
                />
              </div>
            </>
          ) : (
            <EmptyState
              icon="book"
              title="Select or create a document"
              description="The editor opens here with live preview and document details kept close by."
              action={<Button variant="primary" icon="plus" onClick={handleNew}>New document</Button>}
            />
          )}
        </main>

        {(showPreview || showDetails) && (
          <aside className="hidden w-80 shrink-0 flex-col overflow-hidden bg-background xl:flex">
            {showPreview && (
              <section className="min-h-0 flex-1 overflow-hidden border-b border-border">
                <PanelHeader title="Preview" description="Rendered markdown" />
                <div className="h-full overflow-y-auto px-5 py-4">
                  <div className="prose text-sm">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
                  </div>
                </div>
              </section>
            )}

            {showDetails && (
              <section className="shrink-0">
                {selectedId ? (
                  <>
                    <PanelHeader
                      title="Details"
                      description={meta.updated_at ? `Updated ${formatDateTime(meta.updated_at)}` : undefined}
                      action={
                        <Button size="sm" variant="ghost" icon={showTagEditor ? 'check' : 'plus'} onClick={() => setShowTagEditor(v => !v)}>
                          {showTagEditor ? 'Done' : 'Tag'}
                        </Button>
                      }
                    />
                    <div className="space-y-4 p-4">
                      <div>
                        <div className="mb-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">Tags</div>
                        <div className="flex min-h-7 flex-wrap gap-1.5">
                          {docTags.map(tag => (
                            <span
                              key={tag.id}
                              className="inline-flex items-center gap-1 rounded-full border border-border bg-surface px-2 py-0.5 text-xs"
                              style={tag.color ? { borderColor: tag.color, color: tag.color } : undefined}
                            >
                              {tag.name}
                              <button type="button" onClick={() => handleRemoveTag(tag.id)} className="text-muted-foreground hover:text-destructive">
                                <Icon name="close" className="h-3 w-3" />
                              </button>
                            </span>
                          ))}
                          {docTags.length === 0 && !showTagEditor && (
                            <span className="text-xs text-muted-foreground">No tags yet</span>
                          )}
                        </div>
                        {showTagEditor && (
                          <div className="mt-3 space-y-2">
                            {availableTagsToAdd.length > 0 && (
                              <SelectInput
                                defaultValue=""
                                onChange={e => { handleAddTag(e.target.value); e.target.value = '' }}
                                className="h-8 w-full text-xs"
                              >
                                <option value="" disabled>Select tag</option>
                                {availableTagsToAdd.map(tag => (
                                  <option key={tag.id} value={tag.id}>{tag.name}</option>
                                ))}
                              </SelectInput>
                            )}
                            <div className="flex gap-2">
                              <TextInput
                                type="text"
                                value={newTagName}
                                onChange={e => setNewTagName(e.target.value)}
                                onKeyDown={e => { if (e.key === 'Enter') handleCreateTag() }}
                                placeholder="New tag"
                                className="h-8 text-xs"
                              />
                              <Button size="sm" variant="primary" disabled={!newTagName.trim()} onClick={handleCreateTag}>
                                Create
                              </Button>
                            </div>
                          </div>
                        )}
                      </div>

                      {meta.updated_at && (
                        <div className="space-y-2 text-xs text-muted-foreground">
                          <InfoRow label="Version" value={String(meta.version ?? currentVersion)} mono />
                          <InfoRow label="Size" value={meta.size_bytes != null ? formatSize(meta.size_bytes) : '-'} mono />
                          <InfoRow label="Words" value={meta.word_count != null ? String(meta.word_count) : '-'} mono />
                          <InfoRow label="Lines" value={meta.line_count != null ? String(meta.line_count) : '-'} mono />
                          {meta.review_status && <InfoRow label="Review" value={meta.review_status} mono highlight={meta.review_status === 'reviewed'} />}
                          {meta.original_path && (
                            <InfoRow
                              label="Source"
                              value={meta.original_path.split('/').pop() || meta.original_path}
                              mono
                              title={meta.original_path}
                            />
                          )}
                        </div>
                      )}
                    </div>
                  </>
                ) : (
                  <EmptyState compact icon="tag" title="Details appear here" description="Select a saved document to manage tags and metadata." />
                )}
              </section>
            )}
          </aside>
        )}
      </div>
    </div>
  )
}

function SaveStatusPill({ status, label }: { status: SaveStatus; label: string }) {
  const tone = status === 'saved' ? 'success' : status === 'unsaved' || status === 'saving' ? 'warning' : 'danger'
  return <Badge tone={tone}>{label}</Badge>
}

function InfoRow({ label, value, mono, highlight, title }: { label: string; value: string; mono?: boolean; highlight?: boolean; title?: string }) {
  return (
    <div className="flex items-start justify-between gap-3">
      <span>{label}</span>
      <span
        title={title}
        className={cn(
          'min-w-0 truncate text-right text-foreground',
          mono && 'font-mono',
          highlight && 'text-emerald-600 dark:text-emerald-300',
        )}
      >
        {value}
      </span>
    </div>
  )
}
