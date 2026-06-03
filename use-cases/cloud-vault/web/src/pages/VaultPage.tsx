import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { buildListParams, type DocumentFilters } from '../components/DocumentList'
import {
  MarkdownEditor,
  type CursorPosition,
  type MarkdownCommand,
  type MarkdownEditorHandle,
} from '../editor/MarkdownEditor'
import { APIResponseError } from '../api/client'
import { type DocumentDetail, type DocumentSummary, type VersionSummary, documentsAPI } from '../api/documents'
import { type Tag, tagsAPI } from '../api/tags'
import { enqueuePromptExtract, enqueueSummary } from '../api/ai'
import { Badge, Button, EmptyState, SelectInput, TextInput, cn } from '../components/ui'
import { Icon, type IconName } from '../components/icons'

const DEFAULT_CONTENT = '# Untitled\n\nStart writing...\n'
const STORAGE_LIMIT_BYTES = 10 * 1024 * 1024 * 1024

type SaveStatus = 'saved' | 'unsaved' | 'saving' | 'conflict' | 'error'

const defaultFilters: DocumentFilters = {
  q: '',
  status: '',
  sourceType: '',
  isFavorite: null,
  tagId: '',
}

interface VaultPageProps {
  initialDocId?: string | null
  highlightQuery?: string
  onDocumentOpened?: () => void
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

function formatDateCompact(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function statusText(status: SaveStatus) {
  switch (status) {
    case 'saved':
      return '已保存'
    case 'unsaved':
      return '未保存'
    case 'saving':
      return '保存中'
    case 'conflict':
      return '冲突'
    case 'error':
      return '错误'
  }
}

export default function VaultPage({ initialDocId, highlightQuery = '', onDocumentOpened }: VaultPageProps) {
  const [documents, setDocuments] = useState<DocumentSummary[]>([])
  const [listLoading, setListLoading] = useState(true)
  const [filters, setFilters] = useState<DocumentFilters>(defaultFilters)
  const [allTags, setAllTags] = useState<Tag[]>([])
  const [docTags, setDocTags] = useState<Tag[]>([])
  const [versions, setVersions] = useState<VersionSummary[]>([])
  const [versionsLoading, setVersionsLoading] = useState(false)
  const [showTagEditor, setShowTagEditor] = useState(false)
  const [snapshotNote, setSnapshotNote] = useState('')
  const [newTagName, setNewTagName] = useState('')
  const [showPreview, setShowPreview] = useState(true)
  const [showDetails, setShowDetails] = useState(true)
  const [sideTab, setSideTab] = useState<'preview' | 'details'>('preview')
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
  const [cursor, setCursor] = useState<CursorPosition>({ line: 1, column: 1 })

  const originalTitle = useRef('')
  const originalContent = useRef('')
  const editorRef = useRef<MarkdownEditorHandle>(null)

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
      // Keep the current document open when sidebar refresh fails.
    } finally {
      setListLoading(false)
    }
  }

  async function loadAllTags() {
    try {
      const result = await tagsAPI.list()
      setAllTags(result.items)
    } catch {
      // Tag filters are optional for the editor shell.
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

  async function loadVersions(docId: string) {
    setVersionsLoading(true)
    try {
      const result = await documentsAPI.listVersions(docId)
      setVersions(result.items)
    } catch {
      setVersions([])
    } finally {
      setVersionsLoading(false)
    }
  }

  function handleTitleChange(value: string) {
    setTitle(value)
    const dirty = value !== originalTitle.current || content !== originalContent.current
    setSaveStatus(dirty ? 'unsaved' : 'saved')
  }

  function handleContentChange(value: string) {
    setContent(value)
    const dirty = title !== originalTitle.current || value !== originalContent.current
    setSaveStatus(dirty ? 'unsaved' : 'saved')
  }

  async function handleNew() {
    setSelectedId(null)
    setTitle('Untitled')
    setContent(DEFAULT_CONTENT)
    setCurrentVersion(0)
    setMeta({})
    setDocTags([])
    setVersions([])
    setSnapshotNote('')
    originalTitle.current = ''
    originalContent.current = ''
    setSaveStatus('unsaved')
    setStatusMessage('')
    setActionMessage('')
    setDeleteRequested(false)
    window.setTimeout(() => editorRef.current?.focus(), 0)
  }

  async function handleSelectDocument(id: string) {
    setSelectedId(id)
    setSaveStatus('saving')
    setStatusMessage('')
    setActionMessage('')
    setDeleteRequested(false)
    setShowTagEditor(false)
    try {
      const doc = await documentsAPI.get(id)
      await Promise.all([loadDocTags(id), loadVersions(id)])
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
        await loadVersions(result.id)
      } else {
        const result = await documentsAPI.update(selectedId, {
          title: title || 'Untitled',
          content,
          base_version: currentVersion,
        })
        setCurrentVersion(result.version)
        setMeta(prev => ({ ...prev, version: result.version, updated_at: result.updated_at }))
        originalTitle.current = result.title
        originalContent.current = content
        setSaveStatus('saved')
        setStatusMessage(result.changed ? `Saved v${result.version}` : 'No changes')
        await loadList()
        await loadVersions(selectedId)
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
    const next = !meta.is_favorite
    try {
      await documentsAPI.updateFavorite(selectedId, next)
      setMeta(prev => ({ ...prev, is_favorite: next }))
      setDocuments(prev => prev.map(doc => doc.id === selectedId ? { ...doc, is_favorite: next } : doc))
      setActionMessage(next ? 'Added to favorites' : 'Removed from favorites')
    } catch {
      // ignore optional favorite failures
    }
  }

  async function handleArchive() {
    if (!selectedId) return
    const archived = meta.status === 'archived'
    const nextStatus = archived ? 'active' : 'archived'
    try {
      await documentsAPI.updateStatus(selectedId, nextStatus)
      setMeta(prev => ({ ...prev, status: nextStatus }))
      setActionMessage(archived ? 'Document restored' : 'Document archived')
      await loadList()
    } catch {
      // ignore optional archive failures
    }
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
      setVersions([])
      setSnapshotNote('')
      setDeleteRequested(false)
      setActionMessage('Document deleted')
      await loadList()
    } catch {
      // ignore optional delete failures
    }
  }

  async function handleAddTag(tagId: string) {
    if (!selectedId || !tagId) return
    const ids = [...docTags.map(tag => tag.id), tagId]
    try {
      await tagsAPI.setDocumentTags(selectedId, ids)
      await loadDocTags(selectedId)
      setActionMessage('Tag added')
    } catch {
      // ignore optional tag failures
    }
  }

  async function handleRemoveTag(tagId: string) {
    if (!selectedId) return
    try {
      await tagsAPI.removeDocumentTag(selectedId, tagId)
      setDocTags(prev => prev.filter(tag => tag.id !== tagId))
      setActionMessage('Tag removed')
    } catch {
      // ignore optional tag failures
    }
  }

  async function handleCreateTag() {
    const name = newTagName.trim()
    if (!name) return
    try {
      const tag = await tagsAPI.create(name)
      setAllTags(prev => [...prev, tag])
      setNewTagName('')
      if (selectedId) await handleAddTag(tag.id)
    } catch {
      // ignore optional tag failures
    }
  }

  async function handleCreateSnapshot() {
    if (!selectedId || saveStatus === 'unsaved' || saveStatus === 'saving') return
    setActionMessage('')
    try {
      const snapshot = await documentsAPI.createSnapshot(selectedId, snapshotNote)
      setVersions(prev => {
        const next = prev.filter(item => item.version !== snapshot.version)
        return [snapshot, ...next].sort((a, b) => b.version - a.version)
      })
      setSnapshotNote('')
      setActionMessage(`关键版本 v${snapshot.version} 已保存`)
    } catch (err) {
      setActionMessage(err instanceof Error ? err.message : '关键版本保存失败')
    }
  }

  async function handleRestoreVersion(version: number) {
    if (!selectedId || saveStatus === 'unsaved' || saveStatus === 'saving') return
    setSaveStatus('saving')
    setActionMessage('')
    setStatusMessage('')
    try {
      const result = await documentsAPI.restoreVersion(selectedId, version)
      const doc = await documentsAPI.get(selectedId)
      setTitle(doc.title)
      setContent(doc.content)
      setCurrentVersion(doc.version)
      setMeta(doc)
      originalTitle.current = doc.title
      originalContent.current = doc.content
      setSaveStatus('saved')
      setStatusMessage(result.changed ? `Restored v${version} as v${result.version}` : 'No changes')
      await Promise.all([loadList(), loadVersions(selectedId)])
    } catch (err) {
      setSaveStatus('error')
      setStatusMessage(err instanceof Error ? err.message : '版本恢复失败')
    }
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

  const hasDocument = selectedId != null || saveStatus === 'unsaved'
  const isArchived = meta.status === 'archived'
  const rightPanelOpen = showPreview || showDetails
  const availableTagsToAdd = allTags.filter(tag => !docTags.find(docTag => docTag.id === tag.id))
  const charCount = content.replace(/\s/g, '').length
  const wordCount = content.trim() ? content.trim().split(/\s+/).filter(Boolean).length : 0
  const readingMinutes = Math.max(1, Math.ceil(Math.max(charCount, wordCount) / 420))
  const totalBytes = documents.reduce((sum, doc) => sum + doc.size_bytes, 0)

  function runCommand(command: MarkdownCommand) {
    editorRef.current?.applyCommand(command)
  }

  return (
    <div className="vault-workspace-grid h-full overflow-hidden bg-[hsl(var(--workspace-bg))] text-[hsl(var(--workspace-text))]">
      <DocumentNavigator
        documents={documents}
        selectedId={selectedId}
        filters={filters}
        tags={allTags}
        loading={listLoading}
        totalBytes={totalBytes}
        onSelect={handleSelectDocument}
        onNew={handleNew}
        onFiltersChange={setFilters}
      />

      <main className="flex min-w-0 flex-col overflow-hidden border-r border-[hsl(var(--workspace-border))] bg-[hsl(var(--editor-bg))]">
        <EditorTabs
          title={hasDocument ? title || 'Untitled' : '未选择文档'}
          dirty={saveStatus === 'unsaved'}
          disabled={!hasDocument}
          onRename={handleTitleChange}
          onNew={handleNew}
        />
        <MarkdownToolbar
          onCommand={runCommand}
          onSave={handleSave}
          saveDisabled={saveStatus === 'saving' || saveStatus === 'saved'}
          previewOpen={rightPanelOpen}
          onTogglePreview={() => {
            const next = !rightPanelOpen
            setShowPreview(next)
            setShowDetails(next)
          }}
        />

        <div className="min-h-0 flex-1 overflow-hidden">
          {hasDocument ? (
            <MarkdownEditor
              ref={editorRef}
              value={content}
              onChange={handleContentChange}
              onSave={handleSave}
              highlightQuery={activeQuery}
              onCursorChange={setCursor}
            />
          ) : (
            <EmptyState
              icon="book"
              title="选择或新建文档"
              description="编辑器、实时预览和文档详情会在同一个工作区内展开。"
              action={<Button variant="primary" icon="plus" onClick={handleNew}>新建文档</Button>}
            />
          )}
        </div>

        <EditorStatusBar
          cursor={cursor}
          charCount={charCount}
          wordCount={wordCount}
          readingMinutes={readingMinutes}
          status={statusText(saveStatus)}
          message={statusMessage || actionMessage}
        />
      </main>

      {rightPanelOpen && (
        <aside className="hidden min-w-0 flex-col overflow-hidden bg-[hsl(var(--preview-bg))] xl:flex">
          <div className="flex h-12 shrink-0 items-center justify-between border-b border-[hsl(var(--workspace-border))] px-5">
            <div className="flex h-full items-center gap-5">
              {showPreview && (
                <SideTabButton active={sideTab === 'preview'} onClick={() => setSideTab('preview')}>预览</SideTabButton>
              )}
              {showDetails && (
                <SideTabButton active={sideTab === 'details'} onClick={() => setSideTab('details')}>详情</SideTabButton>
              )}
            </div>
            <button
              type="button"
              className="workspace-icon-button"
              aria-label="隐藏右侧面板"
              title="隐藏右侧面板"
              onClick={() => { setShowPreview(false); setShowDetails(false) }}
            >
              <Icon name="collapse" className="h-4 w-4" />
            </button>
          </div>

          {sideTab === 'preview' && showPreview ? (
            <div className="min-h-0 flex-1 overflow-y-auto px-8 py-7">
              {hasDocument ? (
                <div className="prose workspace-prose text-sm">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
                </div>
              ) : (
                <EmptyState compact icon="layout" title="暂无预览" description="选择文档后会显示渲染后的 Markdown。" />
              )}
            </div>
          ) : (
            <DocumentInspector
              selected={selectedId != null}
              meta={meta}
              currentVersion={currentVersion}
              isArchived={isArchived}
              docTags={docTags}
              allTags={allTags}
              availableTagsToAdd={availableTagsToAdd}
              versions={versions}
              versionsLoading={versionsLoading}
              showTagEditor={showTagEditor}
              newTagName={newTagName}
              snapshotNote={snapshotNote}
              snapshotDisabled={saveStatus !== 'saved'}
              deleteRequested={deleteRequested}
              onToggleFavorite={handleToggleFavorite}
              onToggleTagEditor={() => setShowTagEditor(prev => !prev)}
              onAddTag={handleAddTag}
              onRemoveTag={handleRemoveTag}
              onNewTagName={setNewTagName}
              onCreateTag={handleCreateTag}
              onSnapshotNoteChange={setSnapshotNote}
              onCreateSnapshot={handleCreateSnapshot}
              onRestoreVersion={handleRestoreVersion}
              onQueueSummary={() => selectedId && enqueueSummary(selectedId).then(() => setActionMessage('摘要任务已加入队列')).catch(() => setActionMessage('摘要任务创建失败'))}
              onQueuePrompt={() => selectedId && enqueuePromptExtract(selectedId).then(() => setActionMessage('提示词提取已加入队列')).catch(() => setActionMessage('提示词提取失败'))}
              onArchive={handleArchive}
              onDelete={handleDelete}
            />
          )}
        </aside>
      )}
    </div>
  )
}

function DocumentNavigator({
  documents,
  selectedId,
  filters,
  tags,
  loading,
  totalBytes,
  onSelect,
  onNew,
  onFiltersChange,
}: {
  documents: DocumentSummary[]
  selectedId: string | null
  filters: DocumentFilters
  tags: Tag[]
  loading: boolean
  totalBytes: number
  onSelect: (id: string) => void
  onNew: () => void
  onFiltersChange: (filters: DocumentFilters) => void
}) {
  const [query, setQuery] = useState(filters.q)
  const [customGroups, setCustomGroups] = useState<Array<{ id: string; label: string }>>([])
  const [creatingGroup, setCreatingGroup] = useState(false)
  const [groupName, setGroupName] = useState('')
  const [showDocActions, setShowDocActions] = useState(false)
  const [expanded, setExpanded] = useState<Record<string, boolean>>({
    product: true,
    design: true,
    docs: true,
    meetings: false,
    resources: false,
    archive: false,
  })

  useEffect(() => {
    setQuery(filters.q)
  }, [filters.q])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (query !== filters.q) onFiltersChange({ ...filters, q: query })
    }, 180)
    return () => window.clearTimeout(timer)
  }, [filters, onFiltersChange, query])

  const grouped = useMemo(() => {
    const taken = new Set<string>()
    const match = (doc: DocumentSummary, terms: string[]) => terms.some(term => doc.title.toLowerCase().includes(term))
    const pick = (terms: string[]) => documents.filter(doc => {
      if (taken.has(doc.id)) return false
      if (!match(doc, terms)) return false
      taken.add(doc.id)
      return true
    })

    const design = pick(['design', '设计', 'system', '系统', '组件', '图标', '颜色', '排版'])
    const meetings = pick(['meeting', '会议', '纪要', '记录'])
    const resources = documents.filter(doc => !taken.has(doc.id) && doc.source_type === 'imported').map(doc => {
      taken.add(doc.id)
      return doc
    })
    const archive = documents.filter(doc => !taken.has(doc.id) && doc.review_status === 'reviewed').map(doc => {
      taken.add(doc.id)
      return doc
    })
    const docs = documents.filter(doc => !taken.has(doc.id))

    return { design, docs, meetings, resources, archive }
  }, [documents])

  const storagePercent = Math.min(100, (totalBytes / STORAGE_LIMIT_BYTES) * 100)

  function toggle(key: string) {
    setExpanded(prev => ({ ...prev, [key]: !prev[key] }))
  }

  function startCreateGroup() {
    setCreatingGroup(true)
    setShowDocActions(false)
    setGroupName('')
  }

  function cancelCreateGroup() {
    setCreatingGroup(false)
    setGroupName('')
  }

  function setAllTreeExpanded(open: boolean) {
    const keys = ['product', 'design', 'docs', 'meetings', 'resources', 'archive', ...customGroups.map(group => group.id)]
    setExpanded(Object.fromEntries(keys.map(key => [key, open])))
    setShowDocActions(false)
  }

  function applyFilter(next: Partial<DocumentFilters>) {
    onFiltersChange({ ...filters, ...next })
    setShowDocActions(false)
  }

  function clearDocumentFilters() {
    setQuery('')
    onFiltersChange({ q: '', status: '', sourceType: '', isFavorite: null, tagId: '' })
    setShowDocActions(false)
  }

  function commitCreateGroup() {
    const name = groupName.trim()
    if (!name) return
    const id = `custom-${Date.now()}`
    setCustomGroups(prev => [...prev, { id, label: name }])
    setExpanded(prev => ({ ...prev, [id]: true }))
    setCreatingGroup(false)
    setGroupName('')
  }

  return (
    <aside className="flex min-w-0 flex-col overflow-hidden border-r border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel))]">
      <div className="shrink-0 px-4 pb-3 pt-5">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-base font-semibold tracking-tight text-[hsl(var(--workspace-text))]">文档</h2>
          <div className="flex items-center gap-1">
            <button type="button" className="workspace-icon-button" aria-label="新建文档" title="新建文档" onClick={onNew}>
              <Icon name="file" className="h-4 w-4" />
            </button>
            <button type="button" className={cn('workspace-icon-button', creatingGroup && 'bg-[hsl(var(--workspace-accent)/0.14)] text-[hsl(var(--workspace-accent))]')} aria-label="新建分组" title="新建分组" onClick={startCreateGroup}>
              <Icon name="folder" className="h-4 w-4" />
            </button>
            <div className="relative">
              <button
                type="button"
                className={cn('workspace-icon-button', showDocActions && 'bg-[hsl(var(--workspace-accent)/0.14)] text-[hsl(var(--workspace-accent))]')}
                aria-label="更多文档操作"
                title="更多文档操作"
                aria-haspopup="menu"
                aria-expanded={showDocActions}
                onClick={() => setShowDocActions(prev => !prev)}
              >
                <Icon name="more" className="h-4 w-4" />
              </button>
              {showDocActions && (
                <div className="absolute right-0 top-9 z-30 w-48 overflow-hidden rounded-lg border border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel))] py-1 text-sm shadow-lg shadow-slate-950/15" role="menu">
                  <DocActionItem icon="file" label="新建文档" onClick={() => { onNew(); setShowDocActions(false) }} />
                  <DocActionItem icon="folder" label="新建分组" onClick={startCreateGroup} />
                  <MenuDivider />
                  <DocActionItem icon="chevronDown" label="展开全部" onClick={() => setAllTreeExpanded(true)} />
                  <DocActionItem icon="chevronRight" label="折叠全部" onClick={() => setAllTreeExpanded(false)} />
                  <MenuDivider />
                  <DocActionItem icon="check" label="仅活跃文档" onClick={() => applyFilter({ status: '' })} />
                  <DocActionItem icon="archive" label="归档文档" onClick={() => applyFilter({ status: 'archived' })} />
                  <DocActionItem icon="star" label="收藏文档" onClick={() => applyFilter({ isFavorite: true })} />
                  <DocActionItem icon="import" label="导入文档" onClick={() => applyFilter({ sourceType: 'imported' })} />
                  <MenuDivider />
                  <DocActionItem icon="close" label="清除筛选" danger={filters.q !== '' || filters.status !== '' || filters.sourceType !== '' || filters.isFavorite != null || filters.tagId !== ''} onClick={clearDocumentFilters} />
                </div>
              )}
            </div>
          </div>
        </div>
        <div className="relative">
          <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[hsl(var(--workspace-muted))]" />
          <input
            value={query}
            onChange={event => setQuery(event.target.value)}
            placeholder="搜索文档"
            className="workspace-search-input pl-9"
          />
        </div>
        {tags.length > 0 && (
          <SelectInput
            value={filters.tagId}
            onChange={event => onFiltersChange({ ...filters, tagId: event.target.value })}
            className="mt-2 h-8 w-full border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] text-xs"
          >
            <option value="">所有标签</option>
            {tags.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
          </SelectInput>
        )}
        {creatingGroup && (
          <div className="mt-2 flex gap-1.5 rounded-md border border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] p-1.5">
            <input
              value={groupName}
              onChange={event => setGroupName(event.target.value)}
              onKeyDown={event => {
                if (event.key === 'Enter') commitCreateGroup()
                if (event.key === 'Escape') cancelCreateGroup()
              }}
              autoFocus
              placeholder="分组名称"
              className="min-w-0 flex-1 bg-transparent px-2 text-xs text-[hsl(var(--workspace-text))] outline-none placeholder:text-[hsl(var(--workspace-muted))]"
            />
            <button type="button" className="workspace-icon-button h-7 w-7" aria-label="创建分组" title="创建分组" onClick={commitCreateGroup} disabled={!groupName.trim()}>
              <Icon name="check" className="h-3.5 w-3.5" />
            </button>
            <button type="button" className="workspace-icon-button h-7 w-7" aria-label="取消" title="取消" onClick={cancelCreateGroup}>
              <Icon name="close" className="h-3.5 w-3.5" />
            </button>
          </div>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-3">
        {loading ? (
          <div className="space-y-2 px-1 py-2">
            {Array.from({ length: 8 }).map((_, index) => (
              <div key={index} className="h-8 animate-pulse rounded-md bg-[hsl(var(--workspace-panel-soft))]" />
            ))}
          </div>
        ) : documents.length === 0 && customGroups.length === 0 ? (
          <EmptyState compact icon="file" title="暂无文档" description="新建文档、创建分组或导入 Markdown 文件夹。" action={<Button size="sm" variant="primary" icon="plus" onClick={onNew}>新建</Button>} />
        ) : (
          <>
            <TreeRoot
              id="product"
              label="产品设计"
              count={documents.length}
              expanded={expanded.product}
              onToggle={() => toggle('product')}
            />
            {expanded.product && (
              <div className="ml-3 border-l border-[hsl(var(--workspace-border))] pl-2">
                <TreeSection id="design" label="设计系统" docs={grouped.design} selectedId={selectedId} expanded={expanded.design} onToggle={() => toggle('design')} onSelect={onSelect} />
                <TreeSection id="docs" label="产品文档" docs={grouped.docs} selectedId={selectedId} expanded={expanded.docs} onToggle={() => toggle('docs')} onSelect={onSelect} />
              </div>
            )}
            <TreeSection id="meetings" label="会议记录" docs={grouped.meetings} selectedId={selectedId} expanded={expanded.meetings} onToggle={() => toggle('meetings')} onSelect={onSelect} />
            <TreeSection id="resources" label="资源库" docs={grouped.resources} selectedId={selectedId} expanded={expanded.resources} onToggle={() => toggle('resources')} onSelect={onSelect} />
            <TreeSection id="archive" label="归档" docs={grouped.archive} selectedId={selectedId} expanded={expanded.archive} onToggle={() => toggle('archive')} onSelect={onSelect} />
            {customGroups.map(group => (
              <TreeSection key={group.id} id={group.id} label={group.label} docs={[]} selectedId={selectedId} expanded={expanded[group.id] ?? true} onToggle={() => toggle(group.id)} onSelect={onSelect} />
            ))}
          </>
        )}
      </div>

      <div className="shrink-0 border-t border-[hsl(var(--workspace-border))] px-4 py-3 text-xs text-[hsl(var(--workspace-muted))]">
        <div className="mb-2 flex items-center justify-between">
          <span>存储空间</span>
          <span className="font-mono">{formatSize(totalBytes)} / 10 GB</span>
        </div>
        <div className="h-1.5 overflow-hidden rounded-full bg-[hsl(var(--workspace-track))]">
          <div className="h-full rounded-full bg-[hsl(var(--workspace-accent))]" style={{ width: `${storagePercent}%` }} />
        </div>
      </div>
    </aside>
  )
}

function DocActionItem({ icon, label, danger, onClick }: { icon: IconName; label: string; danger?: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      role="menuitem"
      className={cn(
        'flex h-8 w-full items-center gap-2 px-3 text-left text-xs transition-colors hover:bg-[hsl(var(--workspace-panel-soft))]',
        danger ? 'text-destructive' : 'text-[hsl(var(--workspace-text))]',
      )}
      onClick={onClick}
    >
      <Icon name={icon} className="h-3.5 w-3.5 shrink-0 text-[hsl(var(--workspace-muted))]" />
      <span className="min-w-0 flex-1 truncate">{label}</span>
    </button>
  )
}

function MenuDivider() {
  return <div className="my-1 border-t border-[hsl(var(--workspace-border))]" />
}

function TreeRoot({ id, label, count, expanded, onToggle }: { id: string; label: string; count: number; expanded: boolean; onToggle: () => void }) {
  return (
    <button type="button" className="workspace-tree-toggle" onClick={onToggle} aria-expanded={expanded} aria-controls={id}>
      <Icon name={expanded ? 'chevronDown' : 'chevronRight'} className="h-3.5 w-3.5" />
      <span className="min-w-0 flex-1 truncate">{label}</span>
      <span className="font-mono text-[11px]">{count}</span>
    </button>
  )
}

function TreeSection({
  id,
  label,
  docs,
  selectedId,
  expanded,
  onToggle,
  onSelect,
}: {
  id: string
  label: string
  docs: DocumentSummary[]
  selectedId: string | null
  expanded: boolean
  onToggle: () => void
  onSelect: (id: string) => void
}) {
  return (
    <div id={id} className="mt-1">
      <button type="button" className="workspace-tree-toggle" onClick={onToggle} aria-expanded={expanded}>
        <Icon name={expanded ? 'chevronDown' : 'chevronRight'} className="h-3.5 w-3.5" />
        <span className="min-w-0 flex-1 truncate">{label}</span>
        <span className="font-mono text-[11px]">{docs.length}</span>
      </button>
      {expanded && docs.map(doc => (
        <button
          key={doc.id}
          type="button"
          className={cn('workspace-doc-row', selectedId === doc.id && 'is-active')}
          onClick={() => onSelect(doc.id)}
        >
          <Icon name={doc.source_type === 'imported' ? 'import' : 'file'} className="h-4 w-4 shrink-0" />
          <span className="min-w-0 flex-1 truncate">{doc.title || 'Untitled'}</span>
          {doc.is_favorite && <Icon name="star" className="h-3.5 w-3.5 shrink-0 text-amber-500" />}
          <span className="shrink-0 text-[11px] text-[hsl(var(--workspace-muted))]">{formatDateCompact(doc.updated_at)}</span>
        </button>
      ))}
    </div>
  )
}

function EditorTabs({
  title,
  dirty,
  disabled,
  onRename,
  onNew,
}: {
  title: string
  dirty: boolean
  disabled: boolean
  onRename: (value: string) => void
  onNew: () => void
}) {
  return (
    <div className="flex h-12 shrink-0 items-end border-b border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel))] px-3">
      <div className="flex h-10 min-w-0 max-w-[360px] items-center gap-2 rounded-t-lg border border-b-0 border-[hsl(var(--workspace-border))] bg-[hsl(var(--editor-bg))] px-3 text-sm font-medium">
        <input
          value={title}
          disabled={disabled}
          onChange={event => onRename(event.target.value)}
          className="min-w-0 flex-1 bg-transparent text-[hsl(var(--workspace-text))] outline-none disabled:cursor-default"
          aria-label="文档标题"
        />
        {dirty && <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-amber-500" aria-label="未保存" />}
        <Icon name="close" className="h-3.5 w-3.5 shrink-0 text-[hsl(var(--workspace-muted))]" />
      </div>
      <button type="button" className="mb-2 ml-2 flex h-7 w-7 items-center justify-center rounded-md text-[hsl(var(--workspace-muted))] transition-colors hover:bg-[hsl(var(--workspace-panel-soft))] hover:text-[hsl(var(--workspace-text))]" onClick={onNew} aria-label="新建标签">
        <Icon name="plus" className="h-4 w-4" />
      </button>
    </div>
  )
}

function MarkdownToolbar({
  onCommand,
  onSave,
  saveDisabled,
  previewOpen,
  onTogglePreview,
}: {
  onCommand: (command: MarkdownCommand) => void
  onSave: () => void
  saveDisabled: boolean
  previewOpen: boolean
  onTogglePreview: () => void
}) {
  const groups: Array<Array<{ icon: IconName; label: string; command: MarkdownCommand }>> = [
    [
      { icon: 'heading', label: '标题', command: 'heading' },
      { icon: 'bold', label: '加粗', command: 'bold' },
      { icon: 'italic', label: '斜体', command: 'italic' },
      { icon: 'code', label: '代码', command: 'code' },
      { icon: 'quote', label: '引用', command: 'quote' },
    ],
    [
      { icon: 'list', label: '无序列表', command: 'list' },
      { icon: 'listOrdered', label: '有序列表', command: 'orderedList' },
      { icon: 'checkbox', label: '任务', command: 'task' },
    ],
    [
      { icon: 'link', label: '链接', command: 'link' },
      { icon: 'image', label: '图片', command: 'image' },
      { icon: 'table', label: '表格', command: 'table' },
    ],
  ]

  return (
    <div className="flex h-12 shrink-0 items-center justify-between border-b border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] px-4">
      <div className="flex min-w-0 items-center gap-3">
        {groups.map((group, index) => (
          <div key={index} className="flex items-center gap-1 border-r border-[hsl(var(--workspace-border))] pr-3 last:border-r-0 last:pr-0">
            {group.map(item => (
              <button key={item.command} type="button" className="workspace-toolbar-button" title={item.label} aria-label={item.label} onClick={() => onCommand(item.command)}>
                <Icon name={item.icon} className="h-4 w-4" />
              </button>
            ))}
          </div>
        ))}
      </div>
      <div className="flex items-center gap-1">
        <button type="button" className="workspace-toolbar-button" title="撤销" aria-label="撤销" onClick={() => onCommand('undo')}>
          <Icon name="undo" className="h-4 w-4" />
        </button>
        <button type="button" className="workspace-toolbar-button" title="重做" aria-label="重做" onClick={() => onCommand('redo')}>
          <Icon name="redo" className="h-4 w-4" />
        </button>
        <button type="button" className={cn('workspace-toolbar-button', previewOpen && 'is-active')} title="切换预览" aria-label="切换预览" onClick={onTogglePreview}>
          <Icon name="layout" className="h-4 w-4" />
        </button>
        <button type="button" className="workspace-toolbar-button" disabled={saveDisabled} title="保存" aria-label="保存" onClick={onSave}>
          <Icon name="save" className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

function EditorStatusBar({
  cursor,
  charCount,
  wordCount,
  readingMinutes,
  status,
  message,
}: {
  cursor: CursorPosition
  charCount: number
  wordCount: number
  readingMinutes: number
  status: string
  message: string
}) {
  return (
    <div className="flex h-9 shrink-0 items-center justify-between border-t border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel))] px-4 text-xs text-[hsl(var(--workspace-muted))]">
      <div className="flex min-w-0 items-center gap-5">
        <span>行 {cursor.line}, 列 {cursor.column}</span>
        <span>Markdown</span>
        {message && <span className="hidden truncate md:inline">{message}</span>}
      </div>
      <div className="flex items-center gap-5">
        <span>共 {charCount} 字</span>
        <span>词数: {wordCount}</span>
        <span>预计阅读时间: {readingMinutes} 分钟</span>
        <span className="inline-flex items-center gap-1.5">
          <Icon name="check" className="h-3.5 w-3.5 text-[hsl(var(--workspace-accent))]" />
          {status}
        </span>
      </div>
    </div>
  )
}

function SideTabButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'relative flex h-full items-center text-sm font-medium text-[hsl(var(--workspace-muted))] transition-colors hover:text-[hsl(var(--workspace-text))]',
        active && 'text-[hsl(var(--workspace-text))]',
      )}
    >
      {children}
      {active && <span className="absolute bottom-0 left-0 h-0.5 w-full rounded-full bg-[hsl(var(--workspace-accent))]" />}
    </button>
  )
}

function DocumentInspector({
  selected,
  meta,
  currentVersion,
  isArchived,
  docTags,
  allTags,
  availableTagsToAdd,
  versions,
  versionsLoading,
  showTagEditor,
  newTagName,
  snapshotNote,
  snapshotDisabled,
  deleteRequested,
  onToggleFavorite,
  onToggleTagEditor,
  onAddTag,
  onRemoveTag,
  onNewTagName,
  onCreateTag,
  onSnapshotNoteChange,
  onCreateSnapshot,
  onRestoreVersion,
  onQueueSummary,
  onQueuePrompt,
  onArchive,
  onDelete,
}: {
  selected: boolean
  meta: Partial<DocumentDetail>
  currentVersion: number
  isArchived: boolean
  docTags: Tag[]
  allTags: Tag[]
  availableTagsToAdd: Tag[]
  versions: VersionSummary[]
  versionsLoading: boolean
  showTagEditor: boolean
  newTagName: string
  snapshotNote: string
  snapshotDisabled: boolean
  deleteRequested: boolean
  onToggleFavorite: () => void
  onToggleTagEditor: () => void
  onAddTag: (id: string) => void
  onRemoveTag: (id: string) => void
  onNewTagName: (name: string) => void
  onCreateTag: () => void
  onSnapshotNoteChange: (note: string) => void
  onCreateSnapshot: () => void
  onRestoreVersion: (version: number) => void
  onQueueSummary: () => void
  onQueuePrompt: () => void
  onArchive: () => void
  onDelete: () => void
}) {
  if (!selected) {
    return <EmptyState compact icon="tag" title="暂无详情" description="选择已保存文档后可管理标签和元数据。" />
  }

  return (
    <div className="min-h-0 flex-1 overflow-y-auto p-5">
      <div className="mb-5 flex flex-wrap gap-2">
        <Button size="sm" variant="secondary" icon="star" onClick={onToggleFavorite}>{meta.is_favorite ? '取消收藏' : '收藏'}</Button>
        <Button size="sm" variant="secondary" icon="spark" onClick={onQueueSummary}>摘要</Button>
        <Button size="sm" variant="secondary" icon="bolt" onClick={onQueuePrompt}>提取</Button>
      </div>

      <section className="space-y-3 border-b border-[hsl(var(--workspace-border))] pb-5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-semibold">标签</h3>
          <Button size="sm" variant="ghost" icon={showTagEditor ? 'check' : 'plus'} onClick={onToggleTagEditor}>
            {showTagEditor ? '完成' : '添加'}
          </Button>
        </div>
        <div className="flex min-h-8 flex-wrap gap-1.5">
          {docTags.map(tag => (
            <span
              key={tag.id}
              className="inline-flex items-center gap-1 rounded-md border border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] px-2 py-1 text-xs"
              style={tag.color ? { borderColor: tag.color, color: tag.color } : undefined}
            >
              {tag.name}
              <button type="button" onClick={() => onRemoveTag(tag.id)} className="text-[hsl(var(--workspace-muted))] hover:text-destructive">
                <Icon name="close" className="h-3 w-3" />
              </button>
            </span>
          ))}
          {docTags.length === 0 && !showTagEditor && <span className="text-xs text-[hsl(var(--workspace-muted))]">暂无标签</span>}
        </div>
        {showTagEditor && (
          <div className="space-y-2">
            {availableTagsToAdd.length > 0 && (
              <SelectInput
                defaultValue=""
                onChange={event => { onAddTag(event.target.value); event.target.value = '' }}
                className="h-8 w-full border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] text-xs"
              >
                <option value="" disabled>选择标签</option>
                {availableTagsToAdd.map(tag => <option key={tag.id} value={tag.id}>{tag.name}</option>)}
              </SelectInput>
            )}
            <div className="flex gap-2">
              <TextInput
                value={newTagName}
                onChange={event => onNewTagName(event.target.value)}
                onKeyDown={event => { if (event.key === 'Enter') onCreateTag() }}
                placeholder={allTags.length > 0 ? '新标签' : '创建第一个标签'}
                className="h-8 border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] text-xs"
              />
              <Button size="sm" variant="primary" disabled={!newTagName.trim()} onClick={onCreateTag}>创建</Button>
            </div>
          </div>
        )}
      </section>

      <section className="space-y-2 border-b border-[hsl(var(--workspace-border))] py-5 text-xs text-[hsl(var(--workspace-muted))]">
        <InfoRow label="Version" value={String(meta.version ?? currentVersion)} mono />
        <InfoRow label="Size" value={meta.size_bytes != null ? formatSize(meta.size_bytes) : '-'} mono />
        <InfoRow label="Words" value={meta.word_count != null ? String(meta.word_count) : '-'} mono />
        <InfoRow label="Lines" value={meta.line_count != null ? String(meta.line_count) : '-'} mono />
        {meta.review_status && <InfoRow label="Review" value={meta.review_status} mono highlight={meta.review_status === 'reviewed'} />}
        {meta.updated_at && <InfoRow label="Updated" value={formatDateTime(meta.updated_at)} />}
        {meta.original_path && <InfoRow label="Source" value={meta.original_path.split('/').pop() || meta.original_path} mono title={meta.original_path} />}
        {isArchived && <Badge tone="warning">Archived</Badge>}
      </section>

      <section className="space-y-3 border-b border-[hsl(var(--workspace-border))] py-5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-semibold">关键版本</h3>
          <Badge tone="neutral">{versions.length} 个快照</Badge>
        </div>
        <div className="flex gap-2">
          <TextInput
            value={snapshotNote}
            onChange={event => onSnapshotNoteChange(event.target.value)}
            placeholder="备注，例如：发布前确认"
            className="h-8 border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] text-xs"
          />
          <Button size="sm" variant="primary" icon="check" onClick={onCreateSnapshot} disabled={snapshotDisabled}>固定当前版本</Button>
        </div>
        <div className="space-y-1.5">
          {versionsLoading ? (
            <div className="h-9 animate-pulse rounded-md bg-[hsl(var(--workspace-panel-soft))]" />
          ) : versions.length === 0 ? (
            <p className="text-xs text-[hsl(var(--workspace-muted))]">当前策略没有自动保留历史，必要时可手动固定关键版本。</p>
          ) : (
            versions.slice(0, 6).map(version => (
              <div key={version.version} className="flex items-center gap-2 rounded-md border border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] px-2.5 py-2 text-xs">
                <span className="font-mono text-[hsl(var(--workspace-text))]">v{version.version}</span>
                <span className="min-w-0 flex-1 truncate text-[hsl(var(--workspace-muted))]" title={version.note || formatDateTime(version.created_at)}>
                  {version.note || formatDateTime(version.created_at)}
                </span>
                {version.pinned && <Badge tone="success">固定</Badge>}
                <span className="font-mono text-[hsl(var(--workspace-muted))]">{formatSize(version.size_bytes)}</span>
                <Button size="sm" variant="ghost" disabled={snapshotDisabled || version.version === currentVersion} onClick={() => onRestoreVersion(version.version)}>恢复</Button>
              </div>
            ))
          )}
        </div>
      </section>

      <section className="flex gap-2 pt-5">
        <Button size="sm" variant="secondary" icon="archive" onClick={onArchive}>{isArchived ? '恢复' : '归档'}</Button>
        <Button size="sm" variant="danger" icon="trash" onClick={onDelete}>{deleteRequested ? '确认删除' : '删除'}</Button>
      </section>
    </div>
  )
}

function InfoRow({ label, value, mono, highlight, title }: { label: string; value: string; mono?: boolean; highlight?: boolean; title?: string }) {
  return (
    <div className="flex items-start justify-between gap-3">
      <span>{label}</span>
      <span
        title={title}
        className={cn(
          'min-w-0 truncate text-right text-[hsl(var(--workspace-text))]',
          mono && 'font-mono',
          highlight && 'text-emerald-600 dark:text-emerald-300',
        )}
      >
        {value}
      </span>
    </div>
  )
}
