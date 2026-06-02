import { useParams } from 'react-router-dom'
import { useCallback, useEffect, useState } from 'react'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import { api } from '../api'
import type { ESClusterInfo, ESIndexInfo, ESSearchResponse, ESDocument } from '../api'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useToast } from '../components/toastContext'
import { useEsHistory } from '../hooks/useEsHistory'
import { XIcon } from '../components/Icons'
import { EmptyStatePanel, ErrorStatePanel, PageBody, PageShell, PageToolbar } from '../components/workbench'

type Tab = 'overview' | 'mapping' | 'settings' | 'search' | 'documents'

export default function ElasticsearchIndexPanel() {
  const { connId, esIndex } = useParams<{ connId: string; esIndex: string }>()
  const { t } = useI18n()
  const conn = useCurrentConn(connId)
  const { showToast } = useToast()

  const [activeTab, setActiveTab] = useState<Tab>('overview')
  const [clusterInfo, setClusterInfo] = useState<ESClusterInfo | null>(null)
  const [indices, setIndices] = useState<ESIndexInfo[]>([])
  const [mapping, setMapping] = useState<string>('')
  const [settings, setSettings] = useState<string>('')
  const [dsl, setDsl] = useState('{"query":{"match_all":{}},"size":50}')
  const [searchResult, setSearchResult] = useState<ESSearchResponse | null>(null)
  const [searchError, setSearchError] = useState<string>('')
  const [searching, setSearching] = useState(false)
  const esHistory = useEsHistory()
  const [docId, setDocId] = useState('')
  const [docResult, setDocResult] = useState<ESDocument | null>(null)
  const [showRawJson, setShowRawJson] = useState(false)
  const [selectedHit, setSelectedHit] = useState<Record<string, unknown> | null>(null)
  const [copied, setCopied] = useState(false)
  const [loadError, setLoadError] = useState<string>('')
  const [showImportModal, setShowImportModal] = useState(false)
  const [importData, setImportData] = useState(`[
  {"title":"Example"}
]`)
  const [importing, setImporting] = useState(false)

  const loadOverview = useCallback(async () => {
    if (!connId) return
    try {
      const [info, indicesRes] = await Promise.all([
        api.esInfo(connId),
        api.esListIndices(connId)
      ])
      setClusterInfo(info)
      setIndices(indicesRes.indices)
    } catch (err) {
      setLoadError((err as Error).message)
    }
  }, [connId])

  const loadMapping = useCallback(async () => {
    if (!connId || !esIndex) return
    try {
      const res = await api.esGetMapping(connId, esIndex)
      setMapping(JSON.stringify(res, null, 2))
    } catch (err) {
      setLoadError((err as Error).message)
    }
  }, [connId, esIndex])

  const loadSettings = useCallback(async () => {
    if (!connId || !esIndex) return
    try {
      const res = await api.esGetSettings(connId, esIndex)
      setSettings(JSON.stringify(res, null, 2))
    } catch (err) {
      setLoadError((err as Error).message)
    }
  }, [connId, esIndex])

  useEffect(() => {
    if (!connId || !esIndex) return
    queueMicrotask(() => {
      void loadOverview()
    })
  }, [connId, esIndex, loadOverview])

  useEffect(() => {
    queueMicrotask(() => {
      if (activeTab === 'mapping' && connId && esIndex) {
        void loadMapping()
      } else if (activeTab === 'settings' && connId && esIndex) {
        void loadSettings()
      }
    })
  }, [activeTab, connId, esIndex, loadMapping, loadSettings])

  const handleSearch = async () => {
    if (!connId || !esIndex) return
    const startTime = Date.now()
    try {
      const parsed = JSON.parse(dsl)
      setSearching(true)
      setSearchError('')
      const result = await api.esSearch(connId, esIndex, parsed)
      setSearchResult(result)
      // Record to localStorage history
      esHistory.addEntry({
        conn_id: connId,
        index: esIndex,
        dsl,
        duration_ms: Date.now() - startTime,
        result_count: result.hits.length,
      })
    } catch (err) {
      if (err instanceof SyntaxError) {
        setSearchError(t('elasticsearch.invalid_json'))
      } else {
        setSearchError((err as Error).message)
      }
      setSearchResult(null)
    } finally {
      setSearching(false)
    }
  }

  const handleGetDocument = async () => {
    if (!connId || !esIndex || !docId) return
    try {
      const doc = await api.esGetDocument(connId, esIndex, docId)
      setDocResult(doc)
    } catch (err) {
      setDocResult(null)
      showToast((err as Error).message, 'error')
    }
  }

  const handleDeleteDocument = async () => {
    if (!connId || !esIndex || !docId) return
    if (!confirm(t('elasticsearch.delete_document_confirm').replace('{id}', docId).replace('{index}', esIndex))) {
      return
    }
    try {
      await api.esDeleteDocument(connId, esIndex, docId, true)
      showToast(t('elasticsearch.delete_document_success'), 'success')
      setDocResult(null)
    } catch (err) {
      showToast((err as Error).message, 'error')
    }
  }

  const handleCopy = async (text: string) => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleClearHistory = () => {
    if (!confirm(t('elasticsearch.clear_history_confirm'))) return
    esHistory.clearHistory()
  }

  const handleDeleteHistoryEntry = (entryId: string) => {
    esHistory.removeEntry(entryId)
  }

  const handleExport = (format: 'json' | 'ndjson') => {
    if (!connId || !esIndex) return
    let parsed: Record<string, unknown> | undefined
    try {
      parsed = JSON.parse(dsl)
    } catch {
      parsed = undefined
    }
    window.location.href = api.esExport(connId, esIndex, format, parsed)
  }

  const handleImport = async () => {
    if (!connId || !esIndex) return
    let documents: Record<string, unknown>[]
    try {
      const parsed = JSON.parse(importData)
      if (!Array.isArray(parsed)) {
        showToast(t('elasticsearch.import_array_required'), 'error')
        return
      }
      documents = parsed
    } catch {
      showToast(t('elasticsearch.invalid_json'), 'error')
      return
    }
    setImporting(true)
    try {
      const result = await api.esImport(connId, esIndex, documents, true)
      showToast(t('elasticsearch.import_success').replace('{count}', String(result.imported_count)), 'success')
      setShowImportModal(false)
      await loadOverview()
    } catch (err) {
      showToast((err as Error).message, 'error')
    } finally {
      setImporting(false)
    }
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: 'overview', label: t('elasticsearch.tab.overview') },
    { id: 'mapping', label: t('elasticsearch.tab.mapping') },
    { id: 'settings', label: t('elasticsearch.tab.settings') },
    { id: 'search', label: t('elasticsearch.tab.search') },
    { id: 'documents', label: t('elasticsearch.tab.documents') },
  ]

  return (
    <PageShell>
      <WorkbenchHeader
        connectionName={conn?.name}
        resourcePath={esIndex ? [esIndex] : []}
        datasourceType="elasticsearch"
        readonly={conn?.readonly}
        onRefresh={loadOverview}
      />

      <PageToolbar
        leading={
          <div className="flex flex-wrap gap-1">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className="tab-btn"
                data-active={activeTab === tab.id}
              >
                {tab.label}
              </button>
            ))}
          </div>
        }
      />

      {/* Content */}
      <PageBody scroll>
        <div className="p-4">
        {loadError && (
          <ErrorStatePanel
            title={t('error.operation_failed')}
            message={loadError}
            action={
              <button
                className="button-secondary"
                onClick={() => {
                  setLoadError('')
                  if (activeTab === 'overview') void loadOverview()
                  else if (activeTab === 'mapping') void loadMapping()
                  else if (activeTab === 'settings') void loadSettings()
                }}
              >
                {t('common.retry')}
              </button>
            }
          />
        )}
        {!loadError && activeTab === 'overview' && (
          <div className="space-y-4">
            {clusterInfo && (
              <div className="rounded-lg p-4" style={{ background: 'var(--bg-muted)' }}>
                <h3 className="text-sm font-semibold mb-3" style={{ color: 'var(--text-strong)' }}>
                  {t('elasticsearch.cluster_info')}
                </h3>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <span style={{ color: 'var(--text-muted)' }}>{t('elasticsearch.cluster_name')}:</span>
                    <span className="ml-2 font-mono" style={{ color: 'var(--text-strong)' }}>{clusterInfo.cluster_name}</span>
                  </div>
                  <div>
                    <span style={{ color: 'var(--text-muted)' }}>{t('elasticsearch.version')}:</span>
                    <span className="ml-2 font-mono" style={{ color: 'var(--text-strong)' }}>{clusterInfo.version.number}</span>
                  </div>
                  <div>
                    <span style={{ color: 'var(--text-muted)' }}>{t('elasticsearch.lucene_version')}:</span>
                    <span className="ml-2 font-mono" style={{ color: 'var(--text-strong)' }}>{clusterInfo.version.lucene_version}</span>
                  </div>
                </div>
              </div>
            )}

            <div className="rounded-lg p-4 flex flex-wrap gap-2" style={{ background: 'var(--bg-muted)' }}>
              <button
                onClick={() => handleExport('json')}
                className="px-3 py-1.5 text-sm rounded"
                style={{ background: 'var(--accent)', color: 'white' }}
              >
                {t('elasticsearch.export_json')}
              </button>
              <button
                onClick={() => handleExport('ndjson')}
                className="px-3 py-1.5 text-sm rounded"
                style={{ background: 'var(--accent)', color: 'white' }}
              >
                {t('elasticsearch.export_ndjson')}
              </button>
              {!conn?.readonly && (
                <button
                  onClick={() => setShowImportModal(true)}
                  className="px-3 py-1.5 text-sm rounded"
                  style={{ background: 'var(--bg-surface)', color: 'var(--text-default)' }}
                >
                  {t('elasticsearch.import_json')}
                </button>
              )}
            </div>

            <div>
              <h3 className="text-sm font-semibold mb-3" style={{ color: 'var(--text-strong)' }}>
                {t('elasticsearch.indices')} ({indices.length})
              </h3>
              {indices.length === 0 ? (
                <EmptyStatePanel compact title={t('elasticsearch.no_indices')} />
              ) : (
                <div className="rounded-lg overflow-hidden border" style={{ borderColor: 'var(--border-subtle)' }}>
                  <table className="w-full text-sm">
                    <thead style={{ background: 'var(--bg-muted)' }}>
                      <tr>
                        <th className="px-3 py-2 text-left font-medium" style={{ color: 'var(--text-muted)' }}>
                          {t('elasticsearch.index_name')}
                        </th>
                        <th className="px-3 py-2 text-left font-medium" style={{ color: 'var(--text-muted)' }}>
                          {t('elasticsearch.health')}
                        </th>
                        <th className="px-3 py-2 text-left font-medium" style={{ color: 'var(--text-muted)' }}>
                          {t('elasticsearch.status')}
                        </th>
                        <th className="px-3 py-2 text-right font-medium" style={{ color: 'var(--text-muted)' }}>
                          {t('elasticsearch.docs_count')}
                        </th>
                        <th className="px-3 py-2 text-right font-medium" style={{ color: 'var(--text-muted)' }}>
                          {t('elasticsearch.store_size')}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {indices.map(idx => (
                        <tr key={idx.name} className="border-t" style={{ borderColor: 'var(--border-subtle)' }}>
                          <td className="px-3 py-2 font-mono" style={{ color: 'var(--text-strong)' }}>{idx.name}</td>
                          <td className="px-3 py-2">
                            <span className="inline-flex items-center gap-1">
                              <span
                                className="w-2 h-2 rounded-full"
                                style={{
                                  background: idx.health === 'green' ? '#10b981' : idx.health === 'yellow' ? '#f59e0b' : '#ef4444'
                                }}
                              />
                              <span className="capitalize" style={{ color: 'var(--text-strong)' }}>{idx.health}</span>
                            </span>
                          </td>
                          <td className="px-3 py-2 capitalize" style={{ color: 'var(--text-strong)' }}>{idx.status}</td>
                          <td className="px-3 py-2 text-right font-mono" style={{ color: 'var(--text-strong)' }}>{idx.docs_count}</td>
                          <td className="px-3 py-2 text-right font-mono" style={{ color: 'var(--text-strong)' }}>{idx.store_size}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'mapping' && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>
                {t('elasticsearch.mapping_title')}
              </h3>
              <button
                onClick={() => handleCopy(mapping)}
                className="px-3 py-1 text-xs rounded transition-colors"
                style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
              >
                {copied ? t('elasticsearch.copied') : t('elasticsearch.copy')}
              </button>
            </div>
            <pre
              className="rounded-lg p-4 overflow-auto text-xs font-mono"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)', maxHeight: '600px' }}
            >
              {mapping || t('elasticsearch.load_error')}
            </pre>
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>
                {t('elasticsearch.settings_title')}
              </h3>
              <button
                onClick={() => handleCopy(settings)}
                className="px-3 py-1 text-xs rounded transition-colors"
                style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
              >
                {copied ? t('elasticsearch.copied') : t('elasticsearch.copy')}
              </button>
            </div>
            <pre
              className="rounded-lg p-4 overflow-auto text-xs font-mono"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)', maxHeight: '600px' }}
            >
              {settings || t('elasticsearch.load_error')}
            </pre>
          </div>
        )}

        {activeTab === 'search' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-strong)' }}>
                Query DSL
              </label>
              <textarea
                value={dsl}
                onChange={(e) => setDsl(e.target.value)}
                placeholder={t('elasticsearch.search_placeholder')}
                className="w-full h-32 px-3 py-2 rounded-lg font-mono text-sm resize-none"
                style={{
                  background: 'var(--bg-muted)',
                  color: 'var(--text-strong)',
                  border: '1px solid var(--border-subtle)',
                }}
              />
              <div className="mt-2 flex gap-2">
                <button
                  onClick={handleSearch}
                  disabled={searching}
                  className="px-4 py-2 text-sm font-medium rounded transition-colors"
                  style={{
                    background: 'var(--accent)',
                    color: 'white',
                    opacity: searching ? 0.6 : 1,
                  }}
                >
                  {searching ? t('elasticsearch.executing') : t('elasticsearch.execute')}
                </button>
                {showRawJson && (
                  <button
                    onClick={() => setShowRawJson(false)}
                    className="px-4 py-2 text-sm rounded"
                    style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
                  >
                    Hide Raw JSON
                  </button>
                )}
              </div>
            </div>

            {searchError && (
              <ErrorStatePanel compact title={t('error.operation_failed')} message={searchError} />
            )}

            {searchResult && !showRawJson && (
              <div className="space-y-3">
                <div className="flex items-center gap-4 text-sm">
                  <span style={{ color: 'var(--text-muted)' }}>
                    {t('elasticsearch.took').replace('{ms}', String(searchResult.took))}
                  </span>
                  <span style={{ color: 'var(--text-muted)' }}>
                    {t('elasticsearch.total_hits').replace('{count}', String(searchResult.total.value))}
                  </span>
                  {searchResult.timed_out && (
                    <span className="px-2 py-0.5 rounded text-xs" style={{ background: 'var(--warning)18', color: 'var(--warning)' }}>
                      {t('elasticsearch.timed_out')}
                    </span>
                  )}
                  <button
                    onClick={() => setShowRawJson(true)}
                    className="ml-auto px-3 py-1 text-xs rounded"
                    style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
                  >
                    {t('elasticsearch.raw_json')}
                  </button>
                </div>

                {searchResult.hits.length === 0 ? (
                  <EmptyStatePanel compact title={t('elasticsearch.no_results')} />
                ) : (
                  <div className="rounded-lg overflow-hidden border" style={{ borderColor: 'var(--border-subtle)' }}>
                    <table className="w-full text-sm">
                      <thead style={{ background: 'var(--bg-muted)' }}>
                        <tr>
                          <th className="px-3 py-2 text-left font-medium" style={{ color: 'var(--text-muted)' }}>_id</th>
                          <th className="px-3 py-2 text-right font-medium" style={{ color: 'var(--text-muted)' }}>
                            {t('elasticsearch.score')}
                          </th>
                          <th className="px-3 py-2 text-left font-medium" style={{ color: 'var(--text-muted)' }}>_source</th>
                        </tr>
                      </thead>
                      <tbody>
                        {searchResult.hits.map(hit => (
                          <tr
                            key={hit._id}
                            className="border-t cursor-pointer hover:bg-opacity-50"
                            style={{ borderColor: 'var(--border-subtle)' }}
                            onClick={() => setSelectedHit(hit._source)}
                          >
                            <td className="px-3 py-2 font-mono text-xs" style={{ color: 'var(--text-strong)' }}>{hit._id}</td>
                            <td className="px-3 py-2 text-right font-mono text-xs" style={{ color: 'var(--text-muted)' }}>
                              {hit._score?.toFixed(2) || '-'}
                            </td>
                            <td className="px-3 py-2 font-mono text-xs truncate max-w-md" style={{ color: 'var(--text-strong)' }}>
                              {JSON.stringify(hit._source)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            )}

            {searchResult && showRawJson && (
              <pre
                className="rounded-lg p-4 overflow-auto text-xs font-mono"
                style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)', maxHeight: '600px' }}
              >
                {JSON.stringify(searchResult, null, 2)}
              </pre>
            )}

            <div className="border-t pt-4" style={{ borderColor: 'var(--border-subtle)' }}>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>
                  {t('elasticsearch.history')}
                </h3>
                {esHistory.entries.length > 0 && (
                  <button
                    onClick={handleClearHistory}
                    className="px-3 py-1 text-xs rounded"
                    style={{ background: 'var(--error)18', color: 'var(--error)' }}
                  >
                    {t('elasticsearch.clear_history')}
                  </button>
                )}
              </div>
              {esHistory.entries.length === 0 ? (
                <EmptyStatePanel compact title={t('elasticsearch.history_empty')} />
              ) : (
                <div className="space-y-2 max-h-64 overflow-auto">
                  {esHistory.entries.map(entry => (
                    <div
                      key={entry.id}
                      className="rounded-lg p-3 cursor-pointer hover:bg-opacity-70 flex items-start gap-2"
                      style={{ background: 'var(--bg-muted)' }}
                      onClick={() => setDsl(entry.dsl)}
                    >
                      <div className="flex-1 min-w-0">
                        <div className="text-xs font-mono truncate" style={{ color: 'var(--text-strong)' }}>
                          {entry.dsl}
                        </div>
                        <div className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                          {new Date(entry.created_at).toLocaleString()} · {entry.duration_ms}ms · {entry.result_count} hits
                        </div>
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteHistoryEntry(entry.id)
                        }}
                        className="icon-btn h-7 w-7 shrink-0"
                        style={{ background: 'var(--error)18', color: 'var(--error)' }}
                        title={t('elasticsearch.delete_history_entry')}
                      >
                        <XIcon className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'documents' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-strong)' }}>
                {t('elasticsearch.doc_id')}
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={docId}
                  onChange={(e) => setDocId(e.target.value)}
                  placeholder="Enter document _id"
                  className="flex-1 px-3 py-2 rounded-lg text-sm"
                  style={{
                    background: 'var(--bg-muted)',
                    color: 'var(--text-strong)',
                    border: '1px solid var(--border-subtle)',
                  }}
                />
                <button
                  onClick={handleGetDocument}
                  className="px-4 py-2 text-sm font-medium rounded"
                  style={{ background: 'var(--accent)', color: 'white' }}
                >
                  {t('elasticsearch.fetch_document')}
                </button>
              </div>
            </div>

            {docResult && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    <span className="font-mono" style={{ color: 'var(--text-strong)' }}>{docResult._id}</span>
                    <span className="ml-3">v{docResult._version}</span>
                  </div>
                  {!conn?.readonly && (
                    <button
                      onClick={handleDeleteDocument}
                      className="px-3 py-1 text-xs rounded"
                      style={{ background: 'var(--error)18', color: 'var(--error)' }}
                    >
                      {t('elasticsearch.delete_document')}
                    </button>
                  )}
                </div>
                <pre
                  className="rounded-lg p-4 overflow-auto text-xs font-mono"
                  style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)', maxHeight: '600px' }}
                >
                  {JSON.stringify(docResult._source, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
        </div>
      </PageBody>

      {/* Document detail modal */}
      {selectedHit && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
          onClick={() => setSelectedHit(null)}
        >
          <div
            className="panel max-w-2xl w-full mx-4 max-h-[80vh] overflow-auto p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>
                Document Details
              </h3>
              <button
                onClick={() => setSelectedHit(null)}
                className="icon-btn"
                aria-label="Close"
              >
                <XIcon className="h-4 w-4" />
              </button>
            </div>
            <pre
              className="rounded-lg p-4 overflow-auto text-xs font-mono"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)' }}
            >
              {JSON.stringify(selectedHit, null, 2)}
            </pre>
          </div>
        </div>
      )}
      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="panel max-w-2xl w-full mx-4 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>
                {t('elasticsearch.import_json')}
              </h3>
              <button
                onClick={() => setShowImportModal(false)}
                className="icon-btn"
                aria-label="Close"
              >
                <XIcon className="h-4 w-4" />
              </button>
            </div>
            <textarea
              value={importData}
              onChange={e => setImportData(e.target.value)}
              className="textarea h-64 font-mono text-sm"
              style={{
                background: 'var(--bg-muted)',
                color: 'var(--text-strong)',
                border: '1px solid var(--border-subtle)',
              }}
            />
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setShowImportModal(false)}
                className="px-4 py-2 text-sm rounded"
                style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleImport}
                disabled={importing}
                className="px-4 py-2 text-sm rounded"
                style={{ background: 'var(--accent)', color: 'white', opacity: importing ? 0.6 : 1 }}
              >
                {importing ? t('elasticsearch.importing') : t('elasticsearch.import_json')}
              </button>
            </div>
          </div>
        </div>
      )}
    </PageShell>
  )
}
