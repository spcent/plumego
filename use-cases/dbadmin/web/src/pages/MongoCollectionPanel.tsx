import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useToast } from '../components/Toast'
import { api } from '../api'
import type {
  MongoDocQuery,
  MongoDocsResponse,
  MongoIndexInfo,
  MongoAggregateResponse,
  MongoSchemaResponse,
  MongoStatsResponse,
  MongoPipelineEntry,
  MongoObjectIdInfo,
} from '../api'
import ErrorState from '../components/ErrorState'
import EmptyState from '../components/EmptyState'
import ConfirmDialog from '../components/ConfirmDialog'

type ViewMode = 'table' | 'json'
type ActiveTab = 'documents' | 'aggregation' | 'schema' | 'stats'

export default function MongoCollectionPanel() {
  const { t } = useI18n()
  const { showToast } = useToast()
  const { connectionId, database, collection } = useParams<{
    connectionId: string
    database: string
    collection: string
  }>()

  // Tab state
  const [activeTab, setActiveTab] = useState<ActiveTab>('documents')

  // Query state (Documents tab)
  const [filter, setFilter] = useState('')
  const [projection, setProjection] = useState('')
  const [sort, setSort] = useState('')
  const [limit, setLimit] = useState(50)
  const [skip, setSkip] = useState(0)

  // Results state (Documents tab)
  const [results, setResults] = useState<MongoDocsResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // UI state (Documents tab)
  const [viewMode, setViewMode] = useState<ViewMode>('table')
  const [selectedDoc, setSelectedDoc] = useState<Record<string, unknown> | null>(null)
  const [showInsertModal, setShowInsertModal] = useState(false)
  const [showIndexesModal, setShowIndexesModal] = useState(false)
  const [docToDelete, setDocToDelete] = useState<Record<string, unknown> | null>(null)

  // Aggregation state
  const [pipeline, setPipeline] = useState('[]')
  const [aggResult, setAggResult] = useState<MongoAggregateResponse | null>(null)
  const [aggLoading, setAggLoading] = useState(false)
  const [aggError, setAggError] = useState<string | null>(null)
  const [aggHistory, setAggHistory] = useState<MongoPipelineEntry[]>([])
  const [showHistory, setShowHistory] = useState(false)

  // Schema state
  const [schema, setSchema] = useState<MongoSchemaResponse | null>(null)
  const [schemaLoading, setSchemaLoading] = useState(false)
  const [schemaError, setSchemaError] = useState<string | null>(null)

  // Stats state
  const [stats, setStats] = useState<MongoStatsResponse | null>(null)
  const [statsLoading, setStatsLoading] = useState(false)
  const [statsError, setStatsError] = useState<string | null>(null)

  // Export/Import state
  const [showExportModal, setShowExportModal] = useState(false)
  const [showImportModal, setShowImportModal] = useState(false)

  // ObjectId parser state
  const [showObjectIdModal, setShowObjectIdModal] = useState(false)
  const [objectIdInfo, setObjectIdInfo] = useState<MongoObjectIdInfo | null>(null)

  // Extract ID from ObjectId or plain string
  const extractId = (id: unknown): string => {
    if (typeof id === 'object' && id !== null && '$oid' in id) {
      return (id as { $oid: string }).$oid
    }
    return String(id)
  }

  // Run aggregation pipeline
  const runAggregation = useCallback(async () => {
    if (!connectionId || !database || !collection) return

    try {
      const parsed = JSON.parse(pipeline)
      if (!Array.isArray(parsed)) {
        setAggError(t('mongodb.p1.aggregation.pipeline_array_required'))
        return
      }
    } catch (e) {
      setAggError(t('mongodb.p1.aggregation.invalid_json') + ': ' + (e as Error).message)
      return
    }

    setAggLoading(true)
    setAggError(null)

    try {
      const result = await api.mongoAggregate(connectionId, database, collection, pipeline)
      setAggResult(result)
      // Refresh history
      const history = await api.mongoListHistory(connectionId)
      setAggHistory(history)
    } catch (err) {
      setAggError(err instanceof Error ? err.message : 'Aggregation failed')
    } finally {
      setAggLoading(false)
    }
  }, [connectionId, database, collection, pipeline, t])

  // Load aggregation history
  const loadHistory = useCallback(async () => {
    if (!connectionId) return
    try {
      const history = await api.mongoListHistory(connectionId)
      setAggHistory(history)
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to load history', 'error')
    }
  }, [connectionId, showToast])

  // Run schema analysis
  const analyzeSchema = useCallback(async () => {
    if (!connectionId || !database || !collection) return

    setSchemaLoading(true)
    setSchemaError(null)

    try {
      const result = await api.mongoSchema(connectionId, database, collection, 100)
      setSchema(result)
    } catch (err) {
      setSchemaError(err instanceof Error ? err.message : 'Schema analysis failed')
    } finally {
      setSchemaLoading(false)
    }
  }, [connectionId, database, collection])

  // Load collection stats
  const loadStats = useCallback(async () => {
    if (!connectionId || !database || !collection) return

    setStatsLoading(true)
    setStatsError(null)

    try {
      const result = await api.mongoStats(connectionId, database, collection)
      setStats(result)
    } catch (err) {
      setStatsError(err instanceof Error ? err.message : 'Failed to load stats')
    } finally {
      setStatsLoading(false)
    }
  }, [connectionId, database, collection])

  // Parse ObjectId
  const parseObjectId = useCallback(async (objectId: string) => {
    try {
      const result = await api.mongoParseObjectId(objectId)
      setObjectIdInfo(result)
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to parse ObjectId', 'error')
    }
  }, [showToast])

  // Execute query
  const executeQuery = useCallback(async () => {
    if (!connectionId || !database || !collection) return

    setLoading(true)
    setError(null)

    const query: MongoDocQuery = {
      database,
      collection,
      filter: filter.trim() || undefined,
      projection: projection.trim() || undefined,
      sort: sort.trim() || undefined,
      limit,
      skip,
    }

    try {
      const data = await api.mongoQueryDocuments(connectionId, query)
      setResults(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Query failed')
    } finally {
      setLoading(false)
    }
  }, [connectionId, database, collection, filter, projection, sort, limit, skip])

  // Insert document
  const handleInsert = async (docJson: string) => {
    if (!connectionId || !database || !collection) return

    try {
      await api.mongoInsertDocument(connectionId, database, collection, docJson)
      showToast(t('mongodb.document.inserted'), 'success')
      setShowInsertModal(false)
      executeQuery()
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.document.insert_error'), 'error')
    }
  }

  // Update document
  const handleUpdate = async (docJson: string) => {
    if (!connectionId || !database || !collection || !selectedDoc) return

    const id = extractId(selectedDoc._id)
    try {
      await api.mongoUpdateDocument(connectionId, database, collection, id, docJson)
      showToast(t('mongodb.document.saved'), 'success')
      setSelectedDoc(null)
      executeQuery()
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.document.save_error'), 'error')
    }
  }

  // Delete document
  const handleDelete = async () => {
    if (!connectionId || !database || !collection || !docToDelete) return

    const id = extractId(docToDelete._id)
    try {
      await api.mongoDeleteDocument(connectionId, database, collection, id)
      showToast(t('mongodb.document.deleted'), 'success')
      setDocToDelete(null)
      executeQuery()
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.document.delete_error'), 'error')
    }
  }

  // Format value for display
  const formatValue = (value: unknown): string => {
    if (value === null || value === undefined) return 'null'
    if (typeof value === 'object' && '$oid' in value) {
      return `ObjectId("${(value as { $oid: string }).$oid}")`
    }
    if (typeof value === 'object') {
      return JSON.stringify(value)
    }
    return String(value)
  }

  // Get document keys
  const getDocumentKeys = (): string[] => {
    if (!results?.documents || results.documents.length === 0) return []
    const keys = new Set<string>()
    results.documents.forEach(doc => {
      Object.keys(doc).forEach(key => keys.add(key))
    })
    return ['_id', ...Array.from(keys).filter(k => k !== '_id')]
  }

  // Load data when tab changes
  useEffect(() => {
    if (activeTab === 'documents') {
      executeQuery()
    } else if (activeTab === 'aggregation' && aggHistory.length === 0) {
      loadHistory()
    } else if (activeTab === 'schema' && !schema && !schemaLoading) {
      analyzeSchema()
    } else if (activeTab === 'stats' && !stats && !statsLoading) {
      loadStats()
    }
  }, [activeTab, aggHistory.length, schema, schemaLoading, stats, statsLoading, loadHistory, analyzeSchema, loadStats, executeQuery])

  if (!connectionId || !database || !collection) {
    return (
      <ErrorState
        title={t('mongodb.invalidRoute')}
        message={t('mongodb.invalidRouteMessage')}
      />
    )
  }

  return (
    <div className="flex flex-col h-full bg-[var(--bg-primary)]">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[var(--border-primary)] bg-[var(--bg-secondary)]">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-[var(--text-primary)]">
              {collection}
            </h1>
            <p className="text-sm text-[var(--text-secondary)] mt-1">
              {database} / {collection}
            </p>
          </div>
          <button
            onClick={() => setShowObjectIdModal(true)}
            className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.p1.objectid.title')}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="px-6 py-3 border-b border-[var(--border-primary)] bg-[var(--bg-secondary)]">
        <div className="flex gap-2">
          <button
            onClick={() => setActiveTab('documents')}
            className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
              activeTab === 'documents'
                ? 'bg-[var(--accent-primary)] text-white'
                : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
            }`}
          >
            {t('mongodb.p1.tab.documents')}
          </button>
          <button
            onClick={() => setActiveTab('aggregation')}
            className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
              activeTab === 'aggregation'
                ? 'bg-[var(--accent-primary)] text-white'
                : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
            }`}
          >
            {t('mongodb.p1.tab.aggregation')}
          </button>
          <button
            onClick={() => setActiveTab('schema')}
            className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
              activeTab === 'schema'
                ? 'bg-[var(--accent-primary)] text-white'
                : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
            }`}
          >
            {t('mongodb.p1.tab.schema')}
          </button>
          <button
            onClick={() => setActiveTab('stats')}
            className={`px-4 py-2 text-sm font-medium rounded transition-colors ${
              activeTab === 'stats'
                ? 'bg-[var(--accent-primary)] text-white'
                : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
            }`}
          >
            {t('mongodb.p1.tab.stats')}
          </button>
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-auto">
        {activeTab === 'documents' && (
          <DocumentsTab
            connectionId={connectionId}
            database={database}
            collection={collection}
            filter={filter}
            setFilter={setFilter}
            projection={projection}
            setProjection={setProjection}
            sort={sort}
            setSort={setSort}
            limit={limit}
            setLimit={setLimit}
            skip={skip}
            setSkip={setSkip}
            results={results}
            loading={loading}
            error={error}
            viewMode={viewMode}
            setViewMode={setViewMode}
            selectedDoc={selectedDoc}
            setSelectedDoc={setSelectedDoc}
            showInsertModal={showInsertModal}
            setShowInsertModal={setShowInsertModal}
            showIndexesModal={showIndexesModal}
            setShowIndexesModal={setShowIndexesModal}
            docToDelete={docToDelete}
            setDocToDelete={setDocToDelete}
            extractId={extractId}
            executeQuery={executeQuery}
            handleInsert={handleInsert}
            handleUpdate={handleUpdate}
            handleDelete={handleDelete}
            formatValue={formatValue}
            getDocumentKeys={getDocumentKeys}
            showExportModal={showExportModal}
            setShowExportModal={setShowExportModal}
            showImportModal={showImportModal}
            setShowImportModal={setShowImportModal}
          />
        )}

        {activeTab === 'aggregation' && (
          <AggregationTab
            pipeline={pipeline}
            setPipeline={setPipeline}
            aggResult={aggResult}
            aggLoading={aggLoading}
            aggError={aggError}
            aggHistory={aggHistory}
            showHistory={showHistory}
            setShowHistory={setShowHistory}
            runAggregation={runAggregation}
            loadHistory={loadHistory}
            connectionId={connectionId}
          />
        )}

        {activeTab === 'schema' && (
          <SchemaTab
            schema={schema}
            schemaLoading={schemaLoading}
            schemaError={schemaError}
            analyzeSchema={analyzeSchema}
          />
        )}

        {activeTab === 'stats' && (
          <StatsTab
            stats={stats}
            statsLoading={statsLoading}
            statsError={statsError}
            loadStats={loadStats}
          />
        )}
      </div>

      {/* Modals */}
      {selectedDoc && (
        <DocumentViewerModal
          document={selectedDoc}
          onClose={() => setSelectedDoc(null)}
          onSave={(docJson) => handleUpdate(docJson)}
        />
      )}

      {showInsertModal && (
        <InsertDocumentModal
          onClose={() => setShowInsertModal(false)}
          onInsert={handleInsert}
        />
      )}

      {showIndexesModal && (
        <IndexesModal
          connectionId={connectionId}
          database={database}
          collection={collection}
          onClose={() => setShowIndexesModal(false)}
        />
      )}

      {docToDelete && (
        <ConfirmDialog
          open={true}
          title={t('mongodb.document.delete')}
          message={t('mongodb.document.delete_confirm')}
          onConfirm={handleDelete}
          onCancel={() => setDocToDelete(null)}
        />
      )}

      {showExportModal && (
        <ExportModal
          connectionId={connectionId}
          database={database}
          collection={collection}
          filter={filter}
          onClose={() => setShowExportModal(false)}
        />
      )}

      {showImportModal && (
        <ImportModal
          connectionId={connectionId}
          database={database}
          collection={collection}
          onClose={() => setShowImportModal(false)}
          onSuccess={() => {
            setShowImportModal(false)
            executeQuery()
          }}
        />
      )}

      {showHistory && (
        <HistoryModal
          history={aggHistory}
          connectionId={connectionId}
          onClose={() => setShowHistory(false)}
          onSelect={(entry: MongoPipelineEntry) => {
            setPipeline(entry.pipeline)
            setShowHistory(false)
          }}
          loadHistory={loadHistory}
        />
      )}

      {showObjectIdModal && (
        <ObjectIdModal
          onClose={() => {
            setShowObjectIdModal(false)
            setObjectIdInfo(null)
          }}
          onParse={parseObjectId}
          objectIdInfo={objectIdInfo}
        />
      )}
    </div>
  )
}

// Documents Tab Component
function DocumentsTab(props: {
  connectionId: string
  database: string
  collection: string
  filter: string
  setFilter: (v: string) => void
  projection: string
  setProjection: (v: string) => void
  sort: string
  setSort: (v: string) => void
  limit: number
  setLimit: (v: number) => void
  skip: number
  setSkip: (v: number) => void
  results: MongoDocsResponse | null
  loading: boolean
  error: string | null
  viewMode: ViewMode
  setViewMode: (v: ViewMode) => void
  selectedDoc: Record<string, unknown> | null
  setSelectedDoc: (v: Record<string, unknown> | null) => void
  showInsertModal: boolean
  setShowInsertModal: (v: boolean) => void
  showIndexesModal: boolean
  setShowIndexesModal: (v: boolean) => void
  docToDelete: Record<string, unknown> | null
  setDocToDelete: (v: Record<string, unknown> | null) => void
  extractId: (id: unknown) => string
  executeQuery: () => void
  handleInsert: (docJson: string) => void
  handleUpdate: (docJson: string) => void
  handleDelete: () => void
  formatValue: (value: unknown) => string
  getDocumentKeys: () => string[]
  showExportModal: boolean
  setShowExportModal: (v: boolean) => void
  showImportModal: boolean
  setShowImportModal: (v: boolean) => void
}) {
  const { t } = useI18n()
  const {
    filter, setFilter, projection, setProjection, sort, setSort, limit, setLimit, skip, setSkip,
    results, loading, error, viewMode, setViewMode, setSelectedDoc, setShowInsertModal, setShowIndexesModal,
    setDocToDelete, executeQuery, formatValue, getDocumentKeys, setShowExportModal, setShowImportModal
  } = props

  return (
    <div className="flex flex-col h-full">
      {/* Query Controls */}
      <div className="px-6 py-4 border-b border-[var(--border-primary)] space-y-3">
        <div className="grid grid-cols-1 gap-3">
          <div>
            <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1">
              {t('mongodb.browser.filter')}
            </label>
            <input
              type="text"
              value={filter}
              onChange={e => setFilter(e.target.value)}
              placeholder={t('mongodb.browser.filter_placeholder')}
              className="w-full px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1">
                {t('mongodb.browser.projection')}
              </label>
              <input
                type="text"
                value={projection}
                onChange={e => setProjection(e.target.value)}
                placeholder={t('mongodb.browser.projection_placeholder')}
                className="w-full px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1">
                {t('mongodb.browser.sort')}
              </label>
              <input
                type="text"
                value={sort}
                onChange={e => setSort(e.target.value)}
                placeholder={t('mongodb.browser.sort_placeholder')}
                className="w-full px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1">
                {t('mongodb.browser.limit')}
              </label>
              <input
                type="number"
                value={limit}
                onChange={e => setLimit(Math.min(500, Math.max(1, Number(e.target.value))))}
                min={1}
                max={500}
                className="w-full px-3 py-2 text-sm rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-[var(--text-secondary)] mb-1">
                {t('mongodb.browser.skip')}
              </label>
              <input
                type="number"
                value={skip}
                onChange={e => setSkip(Math.max(0, Number(e.target.value)))}
                min={0}
                className="w-full px-3 py-2 text-sm rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
              />
            </div>
          </div>
        </div>

        <div className="flex gap-2">
          <button
            onClick={executeQuery}
            disabled={loading}
            className="px-4 py-2 text-sm font-medium text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? t('mongodb.loading') : t('mongodb.browser.run_query')}
          </button>

          <button
            onClick={() => setShowInsertModal(true)}
            className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.browser.insert')}
          </button>

          <button
            onClick={() => setShowIndexesModal(true)}
            className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.browser.view_indexes')}
          </button>

          <button
            onClick={() => setShowExportModal(true)}
            className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.p1.export.title')}
          </button>

          <button
            onClick={() => setShowImportModal(true)}
            className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.p1.import.title')}
          </button>
        </div>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-auto">
        {error && (
          <ErrorState
            title={t('mongodb.browser.query_error')}
            message={error}
            onRetry={executeQuery}
          />
        )}

        {!error && loading && (
          <div className="flex items-center justify-center py-16">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--accent-primary)]" />
          </div>
        )}

        {!error && !loading && results && results.documents.length === 0 && (
          <EmptyState message={t('mongodb.browser.no_documents')} />
        )}

        {!error && !loading && results && results.documents.length > 0 && (
          <>
            {/* View Mode Toggle */}
            <div className="px-6 py-3 border-b border-[var(--border-primary)] flex items-center justify-between">
              <div className="flex gap-2">
                <button
                  onClick={() => setViewMode('table')}
                  className={`px-3 py-1.5 text-sm rounded ${
                    viewMode === 'table'
                      ? 'bg-[var(--accent-primary)] text-white'
                      : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
                  }`}
                >
                  Table
                </button>
                <button
                  onClick={() => setViewMode('json')}
                  className={`px-3 py-1.5 text-sm rounded ${
                    viewMode === 'json'
                      ? 'bg-[var(--accent-primary)] text-white'
                      : 'bg-[var(--bg-tertiary)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)]'
                  }`}
                >
                  JSON
                </button>
              </div>

              <div className="text-sm text-[var(--text-secondary)]">
                {t('mongodb.browser.total_documents', { count: results.total })}
              </div>
            </div>

            {/* Table View */}
            {viewMode === 'table' && (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="bg-[var(--bg-secondary)] sticky top-0">
                    <tr>
                      {getDocumentKeys().map((key: string) => (
                        <th
                          key={key}
                          className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]"
                        >
                          {key}
                        </th>
                      ))}
                      <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                        {t('common.actions')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.documents.map((doc: Record<string, unknown>, idx: number) => (
                      <tr
                        key={idx}
                        className="border-b border-[var(--border-primary)] hover:bg-[var(--bg-hover)]"
                      >
                        {getDocumentKeys().map((key: string) => (
                          <td
                            key={key}
                            className="px-4 py-3 text-[var(--text-primary)] font-mono text-xs max-w-xs truncate"
                            title={formatValue(doc[key])}
                          >
                            {formatValue(doc[key])}
                          </td>
                        ))}
                        <td className="px-4 py-3">
                          <div className="flex gap-2">
                            <button
                              onClick={() => setSelectedDoc(doc)}
                              className="text-xs text-[var(--accent-primary)] hover:underline"
                            >
                              {t('mongodb.view')}
                            </button>
                            <button
                              onClick={() => setDocToDelete(doc)}
                              className="text-xs text-[var(--danger)] hover:underline"
                            >
                              {t('mongodb.delete')}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {/* JSON View */}
            {viewMode === 'json' && (
              <div className="p-6 space-y-3">
                {results.documents.map((doc: Record<string, unknown>, idx: number) => (
                  <div
                    key={idx}
                    className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]"
                  >
                    <pre className="text-xs font-mono text-[var(--text-primary)] whitespace-pre-wrap">
                      {JSON.stringify(doc, null, 2)}
                    </pre>
                    <div className="mt-3 flex gap-2">
                      <button
                        onClick={() => setSelectedDoc(doc)}
                        className="text-xs text-[var(--accent-primary)] hover:underline"
                      >
                        {t('mongodb.view')}
                      </button>
                      <button
                        onClick={() => setDocToDelete(doc)}
                        className="text-xs text-[var(--danger)] hover:underline"
                      >
                        {t('mongodb.delete')}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

// Aggregation Tab Component
function AggregationTab(props: {
  pipeline: string
  setPipeline: (v: string) => void
  aggResult: MongoAggregateResponse | null
  aggLoading: boolean
  aggError: string | null
  aggHistory: MongoPipelineEntry[]
  showHistory: boolean
  setShowHistory: (v: boolean) => void
  runAggregation: () => void
  loadHistory: () => void
  connectionId: string
}) {
  const { t } = useI18n()
  const {
    pipeline, setPipeline, aggResult, aggLoading, aggError,
    setShowHistory, runAggregation
  } = props

  return (
    <div className="p-6 space-y-4">
      <div>
        <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
          {t('mongodb.p1.aggregation.title')}
        </label>
        <textarea
          value={pipeline}
          onChange={e => setPipeline(e.target.value)}
          placeholder={t('mongodb.p1.aggregation.pipeline_placeholder')}
          className="w-full h-48 px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
        />
      </div>

      <div className="flex gap-2">
        <button
          onClick={runAggregation}
          disabled={aggLoading}
          className="px-4 py-2 text-sm font-medium text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {aggLoading ? t('mongodb.p1.aggregation.running') : t('mongodb.p1.aggregation.run')}
        </button>
        <button
          onClick={() => setShowHistory(true)}
          className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
        >
          {t('mongodb.p1.history.title')}
        </button>
      </div>

      {aggError && (
        <ErrorState
          title={t('mongodb.p1.aggregation.error')}
          message={aggError}
        />
      )}

      {aggResult && !aggError && (
        <div className="space-y-3">
          <div className="text-sm text-[var(--text-secondary)]">
            {t('mongodb.p1.aggregation.result_count', { count: aggResult.count })} · {t('mongodb.p1.aggregation.duration', { ms: aggResult.duration_ms })}
          </div>

          {aggResult.documents.length === 0 ? (
            <EmptyState message={t('mongodb.p1.aggregation.no_results')} />
          ) : (
            <div className="space-y-2">
              {aggResult.documents.map((doc: Record<string, unknown>, idx: number) => (
                <div
                  key={idx}
                  className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]"
                >
                  <pre className="text-xs font-mono text-[var(--text-primary)] whitespace-pre-wrap">
                    {JSON.stringify(doc, null, 2)}
                  </pre>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Schema Tab Component
function SchemaTab(props: {
  schema: MongoSchemaResponse | null
  schemaLoading: boolean
  schemaError: string | null
  analyzeSchema: () => void
}) {
  const { t } = useI18n()
  const { schema, schemaLoading, schemaError, analyzeSchema } = props

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.p1.schema.title')}
          </h2>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            {t('mongodb.p1.schema.description')}
          </p>
        </div>
        <button
          onClick={analyzeSchema}
          disabled={schemaLoading}
          className="px-4 py-2 text-sm font-medium text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {schemaLoading ? t('mongodb.p1.schema.analyzing') : t('mongodb.p1.schema.analyze')}
        </button>
      </div>

      {schemaError && (
        <ErrorState
          title={t('mongodb.p1.schema.error')}
          message={schemaError}
          onRetry={analyzeSchema}
        />
      )}

      {schemaLoading && (
        <div className="flex items-center justify-center py-16">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--accent-primary)]" />
        </div>
      )}

      {!schemaLoading && !schemaError && schema && (
        <div className="space-y-3">
          <div className="text-sm text-[var(--text-secondary)]">
            {t('mongodb.p1.schema.sampled_docs', { count: schema.sampled_count })}
          </div>

          {schema.fields.length === 0 ? (
            <EmptyState message={t('mongodb.p1.schema.no_fields')} />
          ) : (
            <div className="border border-[var(--border-primary)] rounded overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-[var(--bg-secondary)]">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                      {t('mongodb.p1.schema.field_name')}
                    </th>
                    <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                      {t('mongodb.p1.schema.field_type')}
                    </th>
                    <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                      {t('mongodb.p1.schema.field_count')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {schema.fields.map((field, idx) => (
                    <tr key={idx} className="border-b border-[var(--border-primary)]">
                      <td className="px-4 py-3 text-[var(--text-primary)] font-mono text-xs">
                        {field.name}
                      </td>
                      <td className="px-4 py-3 text-[var(--text-primary)] text-xs">
                        {field.types.map((t, i) => (
                          <span key={i} className="inline-block px-2 py-1 mr-1 mb-1 rounded bg-[var(--bg-tertiary)] text-xs">
                            {t.type}
                          </span>
                        ))}
                      </td>
                      <td className="px-4 py-3 text-[var(--text-primary)] text-xs">
                        {field.types.reduce((sum, t) => sum + t.count, 0)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Stats Tab Component
function StatsTab(props: {
  stats: MongoStatsResponse | null
  statsLoading: boolean
  statsError: string | null
  loadStats: () => void
}) {
  const { t } = useI18n()
  const { stats, statsLoading, statsError, loadStats } = props

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--text-primary)]">
          {t('mongodb.p1.stats.title')}
        </h2>
        <button
          onClick={loadStats}
          disabled={statsLoading}
          className="px-4 py-2 text-sm font-medium text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {statsLoading ? t('mongodb.p1.stats.loading') : t('mongodb.p1.stats.load')}
        </button>
      </div>

      {statsError && (
        <ErrorState
          title={t('mongodb.p1.stats.error')}
          message={statsError}
          onRetry={loadStats}
        />
      )}

      {statsLoading && (
        <div className="flex items-center justify-center py-16">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--accent-primary)]" />
        </div>
      )}

      {!statsLoading && !statsError && stats && (
        <div className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.document_count')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {stats.count.toLocaleString()}
              </div>
            </div>

            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.collection_size')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {formatBytes(stats.size)}
              </div>
            </div>

            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.avg_doc_size')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {formatBytes(stats.avgObjSize)}
              </div>
            </div>

            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.storage_size')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {formatBytes(stats.storageSize)}
              </div>
            </div>

            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.index_count')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {stats.indexes.length}
              </div>
            </div>

            <div className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
              <div className="text-sm text-[var(--text-secondary)]">{t('mongodb.p1.stats.index_size')}</div>
              <div className="text-2xl font-semibold text-[var(--text-primary)] mt-1">
                {formatBytes(stats.totalIndexSize)}
              </div>
            </div>
          </div>

          {stats.indexes.length > 0 && (
            <div className="border border-[var(--border-primary)] rounded overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-[var(--bg-secondary)]">
                  <tr>
                    <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                      {t('mongodb.p1.stats.index_name')}
                    </th>
                    <th className="px-4 py-3 text-left font-medium text-[var(--text-secondary)] border-b border-[var(--border-primary)]">
                      {t('mongodb.p1.stats.index_size_col')}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {stats.indexes.map((index, idx) => (
                    <tr key={idx} className="border-b border-[var(--border-primary)]">
                      <td className="px-4 py-3 text-[var(--text-primary)] font-mono text-xs">
                        {index.name}
                      </td>
                      <td className="px-4 py-3 text-[var(--text-primary)] text-xs">
                        {formatBytes(index.size)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Export Modal
function ExportModal(props: {
  connectionId: string
  database: string
  collection: string
  filter: string
  onClose: () => void
}) {
  const { t } = useI18n()
  const { connectionId, database, collection, filter, onClose } = props
  const [format, setFormat] = useState<'json' | 'ndjson' | 'csv'>('json')

  const handleExport = () => {
    const url = api.mongoExport(connectionId, database, collection, format, filter.trim() || undefined)
    window.open(url, '_blank')
    onClose()
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-md w-full">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.p1.export.title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              {t('mongodb.p1.export.format')}
            </label>
            <select
              value={format}
              onChange={e => setFormat(e.target.value as 'json' | 'ndjson' | 'csv')}
              className="w-full px-3 py-2 text-sm rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
            >
              <option value="json">{t('mongodb.p1.export.format_json')}</option>
              <option value="ndjson">{t('mongodb.p1.export.format_ndjson')}</option>
              <option value="csv">{t('mongodb.p1.export.format_csv')}</option>
            </select>
          </div>
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex gap-2 justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.cancel')}
          </button>
          <button
            onClick={handleExport}
            className="px-4 py-2 text-sm text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)]"
          >
            {t('mongodb.p1.export.export')}
          </button>
        </div>
      </div>
    </div>
  )
}

// Import Modal
function ImportModal(props: {
  connectionId: string
  database: string
  collection: string
  onClose: () => void
  onSuccess: () => void
}) {
  const { t } = useI18n()
  const { showToast } = useToast()
  const { connectionId, database, collection, onClose, onSuccess } = props
  const [format, setFormat] = useState<'json' | 'ndjson'>('json')
  const [data, setData] = useState('')
  const [loading, setLoading] = useState(false)

  const handleImport = async () => {
    if (!data.trim()) {
      showToast(t('mongodb.p1.import.empty_data'), 'error')
      return
    }

    setLoading(true)
    try {
      const result = await api.mongoImport(connectionId, database, collection, data, format)
      showToast(t('mongodb.p1.import.success', { count: result.inserted_count }), 'success')
      onSuccess()
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.p1.import.error'), 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] flex flex-col">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.p1.import.title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              {t('mongodb.p1.import.format')}
            </label>
            <select
              value={format}
              onChange={e => setFormat(e.target.value as 'json' | 'ndjson')}
              className="w-full px-3 py-2 text-sm rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
            >
              <option value="json">{t('mongodb.p1.import.format_json')}</option>
              <option value="ndjson">{t('mongodb.p1.import.format_ndjson')}</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              {t('mongodb.p1.import.data')}
            </label>
            <textarea
              value={data}
              onChange={e => setData(e.target.value)}
              placeholder={format === 'json' ? t('mongodb.p1.import.data_placeholder_json') : t('mongodb.p1.import.data_placeholder_ndjson')}
              className="w-full h-64 px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
            />
          </div>
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex gap-2 justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.cancel')}
          </button>
          <button
            onClick={handleImport}
            disabled={loading}
            className="px-4 py-2 text-sm text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? t('mongodb.p1.import.importing') : t('mongodb.p1.import.import')}
          </button>
        </div>
      </div>
    </div>
  )
}

// History Modal
function HistoryModal(props: {
  history: MongoPipelineEntry[]
  connectionId: string
  onClose: () => void
  onSelect: (entry: MongoPipelineEntry) => void
  loadHistory: () => void
}) {
  const { t } = useI18n()
  const { showToast } = useToast()
  const { history, connectionId, onClose, onSelect, loadHistory } = props

  const handleDelete = async (entryId: string) => {
    try {
      await api.mongoDeleteHistoryEntry(connectionId, entryId)
      loadHistory()
      showToast(t('mongodb.p1.history.deleted'), 'success')
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.p1.history.error'), 'error')
    }
  }

  const handleClearAll = async () => {
    if (!confirm(t('mongodb.p1.history.clear_confirm'))) return
    try {
      await api.mongoClearHistory(connectionId)
      loadHistory()
      showToast(t('mongodb.p1.history.cleared'), 'success')
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('mongodb.p1.history.error'), 'error')
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.p1.history.title')}
          </h2>
          <div className="flex gap-2">
            {history.length > 0 && (
              <button
                onClick={handleClearAll}
                className="px-3 py-1.5 text-xs font-medium text-[var(--danger)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
              >
                {t('mongodb.p1.history.clear_all')}
              </button>
            )}
            <button
              onClick={onClose}
              className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
            >
              ✕
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-auto p-6">
          {history.length === 0 ? (
            <EmptyState message={t('mongodb.p1.history.no_history')} />
          ) : (
            <div className="space-y-3">
              {history.map(entry => (
                <div
                  key={entry.id}
                  className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] cursor-pointer"
                  onClick={() => onSelect(entry)}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <pre className="text-xs font-mono text-[var(--text-primary)] whitespace-pre-wrap truncate">
                        {entry.pipeline}
                      </pre>
                      <div className="flex gap-3 mt-2 text-xs text-[var(--text-secondary)]">
                        <span>{t('mongodb.p1.history.result_count')}: {entry.result_count}</span>
                        <span>{t('mongodb.p1.history.duration')}: {entry.duration_ms}ms</span>
                        <span>{new Date(entry.created_at).toLocaleString()}</span>
                      </div>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDelete(entry.id)
                      }}
                      className="text-xs text-[var(--danger)] hover:underline"
                    >
                      {t('mongodb.p1.history.delete')}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.close')}
          </button>
        </div>
      </div>
    </div>
  )
}

// ObjectId Modal
function ObjectIdModal(props: {
  onClose: () => void
  onParse: (objectId: string) => void
  objectIdInfo: MongoObjectIdInfo | null
}) {
  const { t } = useI18n()
  const { onClose, onParse, objectIdInfo } = props
  const [objectId, setObjectId] = useState('')

  const handleParse = () => {
    if (objectId.trim()) {
      onParse(objectId.trim())
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-md w-full">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.p1.objectid.title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--text-secondary)] mb-2">
              {t('mongodb.p1.objectid.input_placeholder')}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={objectId}
                onChange={e => setObjectId(e.target.value)}
                placeholder="507f1f77bcf86cd799439011"
                className="flex-1 px-3 py-2 text-sm font-mono rounded border border-[var(--border-primary)] bg-[var(--bg-primary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
              />
              <button
                onClick={handleParse}
                className="px-4 py-2 text-sm font-medium text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)]"
              >
                {t('mongodb.p1.objectid.parse')}
              </button>
            </div>
          </div>

          {objectIdInfo && (
            <div className="space-y-3">
              <div className="p-3 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
                <div className="text-xs text-[var(--text-secondary)] mb-1">{t('mongodb.p1.objectid.hex')}</div>
                <div className="text-sm font-mono text-[var(--text-primary)]">{objectIdInfo.hex}</div>
              </div>
              <div className="p-3 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
                <div className="text-xs text-[var(--text-secondary)] mb-1">{t('mongodb.p1.objectid.timestamp')}</div>
                <div className="text-sm text-[var(--text-primary)]">{objectIdInfo.timestamp}</div>
              </div>
              <div className="p-3 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]">
                <div className="text-xs text-[var(--text-secondary)] mb-1">{t('mongodb.p1.objectid.counter')}</div>
                <div className="text-sm text-[var(--text-primary)]">{objectIdInfo.counter}</div>
              </div>
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.close')}
          </button>
        </div>
      </div>
    </div>
  )
}

// Document Viewer Modal
function DocumentViewerModal({
  document,
  onClose,
  onSave,
}: {
  document: Record<string, unknown>
  onClose: () => void
  onSave: (docJson: string) => void
}) {
  const { t } = useI18n()
  const [editing, setEditing] = useState(false)
  const [editJson, setEditJson] = useState(JSON.stringify(document, null, 2))

  const handleSave = () => {
    try {
      JSON.parse(editJson)
      onSave(editJson)
    } catch {
      alert(t('mongodb.document.invalid_json'))
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {editing ? t('mongodb.document.editor_title') : t('mongodb.document.viewer_title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-auto p-6">
          {editing ? (
            <textarea
              value={editJson}
              onChange={e => setEditJson(e.target.value)}
              className="w-full h-96 p-3 font-mono text-xs rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
            />
          ) : (
            <pre className="p-4 rounded bg-[var(--bg-secondary)] text-xs font-mono text-[var(--text-primary)] whitespace-pre-wrap">
              {JSON.stringify(document, null, 2)}
            </pre>
          )}
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex gap-2 justify-end">
          {editing ? (
            <>
              <button
                onClick={() => setEditing(false)}
                className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
              >
                {t('mongodb.cancel')}
              </button>
              <button
                onClick={handleSave}
                className="px-4 py-2 text-sm text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)]"
              >
                {t('mongodb.document.save')}
              </button>
            </>
          ) : (
            <>
              <button
                onClick={onClose}
                className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
              >
                {t('mongodb.close')}
              </button>
              <button
                onClick={() => setEditing(true)}
                className="px-4 py-2 text-sm text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)]"
              >
                {t('mongodb.edit')}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

// Insert Document Modal
function InsertDocumentModal({
  onClose,
  onInsert,
}: {
  onClose: () => void
  onInsert: (docJson: string) => void
}) {
  const { t } = useI18n()
  const [docJson, setDocJson] = useState('{\n  \n}')

  const handleInsert = () => {
    try {
      JSON.parse(docJson)
      onInsert(docJson)
    } catch {
      alert(t('mongodb.document.invalid_json'))
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] flex flex-col">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.document.insert_title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-auto p-6">
          <textarea
            value={docJson}
            onChange={e => setDocJson(e.target.value)}
            className="w-full h-96 p-3 font-mono text-xs rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)] text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)]"
          />
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex gap-2 justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.cancel')}
          </button>
          <button
            onClick={handleInsert}
            className="px-4 py-2 text-sm text-white bg-[var(--accent-primary)] rounded hover:bg-[var(--accent-hover)]"
          >
            {t('mongodb.document.insert')}
          </button>
        </div>
      </div>
    </div>
  )
}

// Indexes Modal
function IndexesModal({
  connectionId,
  database,
  collection,
  onClose,
}: {
  connectionId: string
  database: string
  collection: string
  onClose: () => void
}) {
  const { t } = useI18n()
  const [indexes, setIndexes] = useState<MongoIndexInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadIndexes = async () => {
      try {
        const response = await api.mongoListIndexes(connectionId, database, collection)
        setIndexes(response.indexes)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load indexes')
      } finally {
        setLoading(false)
      }
    }
    loadIndexes()
  }, [connectionId, database, collection])

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[var(--bg-primary)] rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
        <div className="px-6 py-4 border-b border-[var(--border-primary)] flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">
            {t('mongodb.indexes.title')}
          </h2>
          <button
            onClick={onClose}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-auto p-6">
          {loading && (
            <div className="flex items-center justify-center py-16">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--accent-primary)]" />
            </div>
          )}

          {error && (
            <ErrorState
              title={t('mongodb.indexes.load_error')}
              message={error}
            />
          )}

          {!loading && !error && indexes.length === 0 && (
            <EmptyState message={t('mongodb.indexes.no_indexes')} />
          )}

          {!loading && !error && indexes.length > 0 && (
            <div className="space-y-4">
              {indexes.map((index, idx) => (
                <div
                  key={idx}
                  className="p-4 rounded border border-[var(--border-primary)] bg-[var(--bg-secondary)]"
                >
                  <div className="font-semibold text-[var(--text-primary)] mb-2">
                    {index.name}
                  </div>
                  <div className="text-xs text-[var(--text-secondary)] mb-2">
                    {index.unique && <span className="mr-3">✓ {t('mongodb.indexes.unique')}</span>}
                    {index.sparse && <span>✓ {t('mongodb.indexes.sparse')}</span>}
                  </div>
                  <pre className="text-xs font-mono text-[var(--text-primary)] bg-[var(--bg-primary)] p-3 rounded">
                    {JSON.stringify(index.key, null, 2)}
                  </pre>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-[var(--border-primary)] flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--text-primary)] bg-[var(--bg-tertiary)] rounded hover:bg-[var(--bg-hover)]"
          >
            {t('mongodb.close')}
          </button>
        </div>
      </div>
    </div>
  )
}
