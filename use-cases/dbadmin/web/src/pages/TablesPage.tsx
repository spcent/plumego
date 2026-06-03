import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, type ResourceNode, type DangerousStatement, ApiError, errorMessage } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import { ChevronRightIcon, TableIcon } from '../components/Icons'
import { EmptyStatePanel, ErrorStatePanel, LoadingState, ModalShell, PageBody, PageShell, PageToolbar } from '../components/workbench'

export default function TablesPage() {
  const { connId, dbName } = useParams<{ connId: string; dbName: string }>()
  const [tables, setTables] = useState<ResourceNode[]>([])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [dropTarget, setDropTarget] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)
  const [newTableName, setNewTableName] = useState('')
  const [importOpen, setImportOpen] = useState(false)
  const [importSql, setImportSql] = useState('')
  const [importing, setImporting] = useState(false)
  const [importConfirm, setImportConfirm] = useState<{ dangerous: DangerousStatement[] } | null>(null)
  const [copyingSchema, setCopyingSchema] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const isReadonly = currentConn?.readonly ?? false

  const loadTables = useCallback(async () => {
    if (!connId || !dbName) return
    setLoading(true)
    setLoadError(null)
    try {
      setTables(await api.resources(connId, dbName))
    } catch (e) {
      const message = errorMessage(e, t('resource.error'))
      setLoadError(message)
      showToast(message)
    } finally {
      setLoading(false)
    }
  }, [connId, dbName, showToast, t])

  useEffect(() => {
    const id = window.setTimeout(() => { void loadTables() }, 0)
    return () => window.clearTimeout(id)
  }, [loadTables])

  async function handleDropTable() {
    if (!dropTarget) return
    try {
      await api.dropTable(connId!, dbName!, dropTarget)
      setTables(ts => ts.filter(x => x.name !== dropTarget))
      setDropTarget(null)
    } catch (e) {
      showToast(errorMessage(e, 'Drop failed'))
    }
  }

  async function handleCopySchema() {
    if (!connId || !dbName) return
    setCopyingSchema(true)
    try {
      const r = await api.schemaDoc(connId, dbName)
      await navigator.clipboard.writeText(r.markdown)
      showToast(t('tables.copy_schema.success'), 'success')
    } catch (e) {
      showToast(errorMessage(e, 'Copy failed'))
    } finally {
      setCopyingSchema(false)
    }
  }

  async function handleImport(confirmDangerous: boolean) {
    if (!connId || !dbName || !importSql.trim()) return
    setImporting(true)
    try {
      const r = await api.importSQL(connId, dbName, importSql, confirmDangerous)
      showToast(`${r.statements_executed} statement(s) executed`, 'success')
      if (r.errors > 0) showToast(`${r.errors} statement(s) failed`)
      setImportOpen(false)
      setImportSql('')
    } catch (e) {
      if (e instanceof ApiError && e.details?.confirm_required) {
        setImportConfirm({ dangerous: e.details.dangerous_statements as DangerousStatement[] })
      } else {
        showToast(errorMessage(e, 'Import failed'))
      }
    } finally {
      setImporting(false)
    }
  }

  function handleFileChange(ev: React.ChangeEvent<HTMLInputElement>) {
    const file = ev.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = e => setImportSql(String(e.target?.result ?? ''))
    reader.readAsText(file)
  }

  if (!dbName) return <div className="p-6" style={{ color: 'var(--text-muted)' }}>Select a database from the sidebar.</div>

  return (
    <PageShell>
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={dbName ? [dbName] : []}
        datasourceType={currentConn?.driver ?? 'mysql'}
        readonly={isReadonly}
      />
      <PageToolbar
        leading={
          <>
          <button
            onClick={handleCopySchema}
            disabled={copyingSchema || loading || tables.length === 0}
            className="btn btn-ghost h-8 px-3 text-xs disabled:opacity-50"
          >
            {copyingSchema ? t('tables.copy_schema.loading') : t('tables.copy_schema')}
          </button>
          <Link
            to={`/conn/${connId}/query`}
            className="btn btn-ghost h-8 px-3 text-xs"
          >
            {t('tables.sql_console')}
          </Link>
          {!isReadonly && (
            <button
              onClick={() => setImportOpen(true)}
              className="btn btn-ghost h-8 px-3 text-xs"
            >
              {t('tables.import')}
            </button>
          )}
          </>
        }
        trailing={
          !isReadonly && (
            <button
              onClick={() => setCreating(true)}
              className="btn btn-primary h-8 px-3 text-xs"
            >
              {t('tables.new_table')}
            </button>
          )
        }
      />

      <PageBody>
        {loading && tables.length === 0 ? (
          <LoadingState title={t('resource.loading')} detail={dbName} />
        ) : loadError ? (
          <ErrorStatePanel
            title={t('resource.error')}
            message={loadError}
            action={
              <button type="button" onClick={() => void loadTables()} className="btn btn-ghost h-8 px-3 text-xs">
                {t('resource.retry')}
              </button>
            }
          />
        ) : tables.length === 0 ? (
          <EmptyStatePanel title={t('tables.empty')} detail={dbName} />
        ) : (
          <div className="data-table-wrap h-full">
          <table className="data-table">
            <thead>
              <tr>
                {[t('tables.col.name'), t('tables.col.type'), t('tables.col.engine'), t('tables.col.rows'), t('tables.col.comment'), ''].map(h => (
                  <th key={h} className={`px-4 py-2 ${h === t('tables.col.rows') ? 'text-right' : 'text-left'}`}>
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {tables.map(node => {
                const tableType = (node.meta?.table_type as string) || ''
                const engine    = (node.meta?.engine    as string) || ''
                const rows      =  node.meta?.rows      as number | undefined
                const comment   = (node.meta?.comment   as string) || ''
                return (
                  <tr key={node.name}>
                    <td className="px-4 py-2">
                      <Link
                        to={`/conn/${connId}/db/${dbName}/tables/${node.name}/data`}
                        className="inline-flex items-center gap-2 font-mono font-medium hover:underline"
                        style={{ color: 'var(--accent)' }}
                      >
                        <TableIcon className="h-3.5 w-3.5" />
                        {node.name}
                      </Link>
                    </td>
                    <td className="px-4 py-2 text-xs" style={{ color: 'var(--text-muted)' }}>{tableType}</td>
                    <td className="px-4 py-2 text-xs" style={{ color: 'var(--text-muted)' }}>{engine || '-'}</td>
                    <td className="px-4 py-2 text-right text-xs tabular-nums" style={{ color: 'var(--text-muted)' }}>
                      {rows?.toLocaleString() || '-'}
                    </td>
                    <td className="max-w-xs truncate px-4 py-2 text-xs" style={{ color: 'var(--text-subtle)' }}>{comment}</td>
                    <td className="px-4 py-2 text-right whitespace-nowrap">
                      <Link
                        to={`/conn/${connId}/db/${dbName}/tables/${node.name}/data`}
                        className="btn btn-ghost mr-1 h-7 px-2 text-xs"
                      >
                        {t('tables.action.data')}
                      </Link>
                      <Link
                        to={`/conn/${connId}/db/${dbName}/tables/${node.name}/structure?tab=columns`}
                        className="btn btn-ghost mr-1 h-7 px-2 text-xs"
                      >
                        {t('tables.action.fields')}
                      </Link>
                      <Link
                        to={`/conn/${connId}/db/${dbName}/tables/${node.name}/structure?tab=indexes`}
                        className="btn btn-ghost mr-1 h-7 px-2 text-xs"
                      >
                        {t('tables.action.indexes')}
                      </Link>
                      {!isReadonly && (
                        <button
                          onClick={() => setDropTarget(node.name)}
                          className="btn btn-danger h-7 px-2 text-xs"
                        >
                          {t('tables.action.drop')}
                        </button>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          </div>
        )}
      </PageBody>

      <ConfirmDialog
        open={!!dropTarget}
        title={t('tables.drop.title')}
        message={t('tables.drop.message', { table: dropTarget ?? '' })}
        confirmLabel={t('confirm.drop')}
        dangerous
        onConfirm={handleDropTable}
        onCancel={() => setDropTarget(null)}
      />

      {importOpen && (
        <ModalShell
          title={t('import.title')}
          widthClass="max-w-lg"
          onClose={() => { setImportOpen(false); setImportSql('') }}
          footer={
            <>
              <button
                onClick={() => { setImportOpen(false); setImportSql('') }}
                className="btn btn-ghost h-8 px-3 text-xs"
              >{t('data.cancel')}</button>
              <button
                onClick={() => handleImport(false)}
                disabled={importing || !importSql.trim()}
                className="btn btn-primary h-8 px-3 text-xs disabled:opacity-50"
              >{importing ? t('import.running') : t('import.run')}</button>
            </>
          }
        >
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm" style={{ color: 'var(--text-muted)' }}>{t('import.paste')}</span>
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="btn btn-ghost h-7 px-2 text-xs"
                >{t('import.file')}</button>
              </div>
              <input ref={fileInputRef} type="file" accept=".sql,.txt" className="hidden" onChange={handleFileChange} />
              <textarea
                value={importSql}
                onChange={e => setImportSql(e.target.value)}
                rows={10}
                className="textarea min-h-48 font-mono text-xs"
                placeholder="-- Paste SQL here or upload a file above"
              />
            </div>
        </ModalShell>
      )}

      <ConfirmDialog
        open={!!importConfirm}
        title={t('import.confirm_danger')}
        message={`${importConfirm?.dangerous.length ?? 0} dangerous statement(s): ${
          importConfirm?.dangerous.map(d => d.snippet).join(' | ') ?? ''
        }`}
        confirmLabel={t('import.confirm_proceed')}
        dangerous
        onConfirm={() => { setImportConfirm(null); handleImport(true) }}
        onCancel={() => setImportConfirm(null)}
      />

      {creating && (
        <ModalShell
          title={t('tables.create.title')}
          widthClass="max-w-sm"
          onClose={() => setCreating(false)}
          footer={
            <>
              <button onClick={() => setCreating(false)} className="btn btn-ghost h-8 px-3 text-xs">{t('connections.form.cancel')}</button>
              <Link
                to={`/conn/${connId}/query`}
                className="btn btn-primary h-8 px-3 text-xs"
                onClick={() => setCreating(false)}
              >
                <ChevronRightIcon className="h-4 w-4" />
                {t('tables.create.open_console')}
              </Link>
            </>
          }
        >
            <p className="mb-4 text-sm" style={{ color: 'var(--text-muted)' }}>
              {t('tables.create.hint')}
            </p>
            <input
              value={newTableName}
              onChange={e => setNewTableName(e.target.value)}
              className="input mb-4"
              placeholder="table_name"
            />
        </ModalShell>
      )}
    </PageShell>
  )
}
