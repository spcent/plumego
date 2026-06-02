import { useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, type ResourceNode, type DangerousStatement, ApiError } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import { ChevronRightIcon, TableIcon } from '../components/Icons'

export default function TablesPage() {
  const { connId, dbName } = useParams<{ connId: string; dbName: string }>()
  const [tables, setTables] = useState<ResourceNode[]>([])
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

  useEffect(() => {
    if (!connId || !dbName) return
    api.resources(connId, dbName).then(setTables).catch(e => showToast(e.message))
  }, [connId, dbName, showToast])

  async function handleDropTable() {
    if (!dropTarget) return
    try {
      await api.dropTable(connId!, dbName!, dropTarget)
      setTables(ts => ts.filter(x => x.name !== dropTarget))
      setDropTarget(null)
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Drop failed')
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
      showToast(e instanceof Error ? e.message : 'Copy failed')
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
        showToast(e instanceof Error ? e.message : 'Import failed')
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
    <div className="flex flex-col h-full">
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={dbName ? [dbName] : []}
        datasourceType={currentConn?.driver ?? 'mysql'}
        readonly={isReadonly}
      />
      <div className="toolbar shrink-0">
        <div className="flex flex-wrap gap-2">
          <button
            onClick={handleCopySchema}
            disabled={copyingSchema || tables.length === 0}
            className="btn btn-ghost disabled:opacity-50"
          >
            {copyingSchema ? t('tables.copy_schema.loading') : t('tables.copy_schema')}
          </button>
          <Link
            to={`/conn/${connId}/query`}
            className="btn btn-ghost"
          >
            {t('tables.sql_console')}
          </Link>
          {!isReadonly && (
            <button
              onClick={() => setImportOpen(true)}
              className="btn btn-ghost"
            >
              {t('tables.import')}
            </button>
          )}
          {!isReadonly && (
            <button
              onClick={() => setCreating(true)}
              className="btn btn-primary"
            >
              {t('tables.new_table')}
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <table className="w-full text-sm">
          <thead className="sticky top-0 z-10 border-b" style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)' }}>
            <tr>
              {[t('tables.col.name'), t('tables.col.type'), t('tables.col.engine'), t('tables.col.rows'), t('tables.col.comment'), ''].map(h => (
                <th key={h} className={`px-4 py-2 font-medium ${h === t('tables.col.rows') ? 'text-right' : 'text-left'}`} style={{ color: 'var(--text-muted)' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y" style={{ borderColor: 'var(--border-subtle)' }}>
            {tables.map(node => {
              const tableType = (node.meta?.table_type as string) || ''
              const engine    = (node.meta?.engine    as string) || ''
              const rows      =  node.meta?.rows      as number | undefined
              const comment   = (node.meta?.comment   as string) || ''
              return (
                <tr key={node.name} className="transition-colors hover:bg-[var(--bg-hover)]">
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
            {tables.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center" style={{ color: 'var(--text-muted)' }}>{t('tables.empty')}</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

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
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="panel w-full max-w-lg mx-4 flex flex-col gap-4 p-6">
            <h2 className="font-semibold" style={{ color: 'var(--text-strong)' }}>{t('import.title')}</h2>
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
            <div className="flex justify-end gap-2">
              <button
                onClick={() => { setImportOpen(false); setImportSql('') }}
                className="btn btn-ghost"
              >{t('data.cancel')}</button>
              <button
                onClick={() => handleImport(false)}
                disabled={importing || !importSql.trim()}
                className="btn btn-primary disabled:opacity-50"
              >{importing ? t('import.running') : t('import.run')}</button>
            </div>
          </div>
        </div>
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
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="panel w-full max-w-sm mx-4 p-6">
            <h2 className="mb-4 font-semibold" style={{ color: 'var(--text-strong)' }}>{t('tables.create.title')}</h2>
            <p className="mb-4 text-sm" style={{ color: 'var(--text-muted)' }}>
              {t('tables.create.hint')}
            </p>
            <input
              value={newTableName}
              onChange={e => setNewTableName(e.target.value)}
              className="input mb-4"
              placeholder="table_name"
            />
            <div className="flex justify-end gap-2">
              <button onClick={() => setCreating(false)} className="btn btn-ghost">{t('connections.form.cancel')}</button>
              <Link
                to={`/conn/${connId}/query`}
                className="btn btn-primary"
                onClick={() => setCreating(false)}
              >
                <ChevronRightIcon className="h-4 w-4" />
                {t('tables.create.open_console')}
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
