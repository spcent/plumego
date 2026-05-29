import { useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, type ResourceNode, type DangerousStatement, ApiError } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'

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
  }, [connId, dbName])

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

  if (!dbName) return <div className="p-6 text-gray-400">Select a database from the sidebar.</div>

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-5 pb-4 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between shrink-0">
        <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100">
          Tables — <span className="text-blue-600">{dbName}</span>
        </h1>
        <div className="flex gap-2">
          <button
            onClick={handleCopySchema}
            disabled={copyingSchema || tables.length === 0}
            className="bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-2 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50"
          >
            {copyingSchema ? t('tables.copy_schema.loading') : t('tables.copy_schema')}
          </button>
          <Link
            to={`/conn/${connId}/query`}
            className="bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-2 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            {t('tables.sql_console')}
          </Link>
          {!isReadonly && (
            <button
              onClick={() => setImportOpen(true)}
              className="bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-2 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600"
            >
              {t('tables.import')}
            </button>
          )}
          {!isReadonly && (
            <button
              onClick={() => setCreating(true)}
              className="bg-blue-600 text-white px-3 py-2 rounded text-sm hover:bg-blue-700"
            >
              {t('tables.new_table')}
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10 border-b border-gray-200 dark:border-gray-700">
            <tr>
              {[t('tables.col.name'), t('tables.col.type'), t('tables.col.engine'), t('tables.col.rows'), t('tables.col.comment'), ''].map(h => (
                <th key={h} className={`px-4 py-2 font-medium text-gray-600 dark:text-gray-400 ${h === t('tables.col.rows') ? 'text-right' : 'text-left'}`}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {tables.map(node => {
              const tableType = (node.meta?.table_type as string) || ''
              const engine    = (node.meta?.engine    as string) || ''
              const rows      =  node.meta?.rows      as number | undefined
              const comment   = (node.meta?.comment   as string) || ''
              return (
                <tr key={node.name} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-2">
                    <Link
                      to={`/conn/${connId}/db/${dbName}/tables/${node.name}/data`}
                      className="text-blue-600 hover:underline font-mono font-medium"
                    >
                      {node.name}
                    </Link>
                  </td>
                  <td className="px-4 py-2 text-gray-500 dark:text-gray-400 text-xs">{tableType}</td>
                  <td className="px-4 py-2 text-gray-500 dark:text-gray-400 text-xs">{engine || '-'}</td>
                  <td className="px-4 py-2 text-gray-500 dark:text-gray-400 text-right tabular-nums text-xs">
                    {rows?.toLocaleString() || '-'}
                  </td>
                  <td className="px-4 py-2 text-gray-400 text-xs truncate max-w-xs">{comment}</td>
                  <td className="px-4 py-2 text-right whitespace-nowrap">
                    <Link
                      to={`/conn/${connId}/db/${dbName}/tables/${node.name}/data`}
                      className="inline-flex items-center text-xs text-gray-500 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/30 px-2 py-1 rounded mr-1"
                    >
                      {t('tables.action.data')}
                    </Link>
                    <Link
                      to={`/conn/${connId}/db/${dbName}/tables/${node.name}/structure?tab=columns`}
                      className="inline-flex items-center text-xs text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 px-2 py-1 rounded mr-1"
                    >
                      {t('tables.action.fields')}
                    </Link>
                    <Link
                      to={`/conn/${connId}/db/${dbName}/tables/${node.name}/structure?tab=indexes`}
                      className="inline-flex items-center text-xs text-gray-500 hover:text-violet-600 hover:bg-violet-50 dark:hover:bg-violet-900/30 px-2 py-1 rounded mr-1"
                    >
                      {t('tables.action.indexes')}
                    </Link>
                    {!isReadonly && (
                      <button
                        onClick={() => setDropTarget(node.name)}
                        className="text-xs text-red-400 hover:text-red-600 px-2 py-1 rounded hover:bg-red-50 dark:hover:bg-red-900/30"
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
                <td colSpan={6} className="px-4 py-8 text-center text-gray-400">{t('tables.empty')}</td>
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
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-lg mx-4 flex flex-col gap-4">
            <h2 className="font-bold text-gray-900 dark:text-gray-100">{t('import.title')}</h2>
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm text-gray-600 dark:text-gray-400">{t('import.paste')}</span>
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="text-xs text-blue-600 hover:underline"
                >{t('import.file')}</button>
              </div>
              <input ref={fileInputRef} type="file" accept=".sql,.txt" className="hidden" onChange={handleFileChange} />
              <textarea
                value={importSql}
                onChange={e => setImportSql(e.target.value)}
                rows={10}
                className="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-xs font-mono bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 resize-none"
                placeholder="-- Paste SQL here or upload a file above"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => { setImportOpen(false); setImportSql('') }}
                className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800"
              >{t('data.cancel')}</button>
              <button
                onClick={() => handleImport(false)}
                disabled={importing || !importSql.trim()}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
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
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-sm mx-4">
            <h2 className="font-bold mb-4 text-gray-900 dark:text-gray-100">{t('tables.create.title')}</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              {t('tables.create.hint')}
            </p>
            <input
              value={newTableName}
              onChange={e => setNewTableName(e.target.value)}
              className="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-sm mb-4 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              placeholder="table_name"
            />
            <div className="flex justify-end gap-2">
              <button onClick={() => setCreating(false)} className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">{t('connections.form.cancel')}</button>
              <Link
                to={`/conn/${connId}/query`}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                onClick={() => setCreating(false)}
              >
                {t('tables.create.open_console')}
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
