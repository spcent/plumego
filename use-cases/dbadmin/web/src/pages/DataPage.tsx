import { useEffect, useState, useCallback, useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { api, type RowsResponse, type TableStructure, type FilterCondition, type FilterOperator } from '../api'
import CellViewer from '../components/CellViewer'
import ConfirmDialog from '../components/ConfirmDialog'
import RowDrawer from '../components/RowDrawer'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'
import { toInsertSQL, toCSV, toRowJSON } from '../utils/copyFormats'

interface ConfirmState {
  title: string
  message: string
  onConfirm: () => void
}

const OPERATORS: { value: FilterOperator; label: string }[] = [
  { value: 'eq', label: '=' },
  { value: 'ne', label: '!=' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
  { value: 'like', label: 'LIKE' },
  { value: 'not_like', label: 'NOT LIKE' },
  { value: 'is_null', label: 'IS NULL' },
  { value: 'is_not_null', label: 'IS NOT NULL' },
]

function renderCell(v: unknown) {
  if (v === null || v === undefined) {
    return <span className="text-gray-300 dark:text-gray-600 italic select-none">NULL</span>
  }
  if (v === '') {
    return <span className="text-gray-400 dark:text-gray-500 font-mono italic select-none text-xs">''</span>
  }
  const s = String(v)
  if (s.startsWith('<BLOB ')) {
    return <span className="text-orange-500 dark:text-orange-400 font-mono text-xs">{s}</span>
  }
  // Quick JSON heuristic: starts with { or [ and ends with } or ]
  const t = s.trim()
  if ((t.startsWith('{') && t.endsWith('}')) || (t.startsWith('[') && t.endsWith(']'))) {
    return (
      <>
        <span className="mr-1.5 inline-block align-middle text-[10px] bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 px-1 rounded font-mono leading-snug">JSON</span>
        {s}
      </>
    )
  }
  return s
}

function cellTitle(v: unknown): string {
  if (v === null || v === undefined) return 'NULL'
  if (v === '') return '(empty string)'
  return String(v)
}

export default function DataPage() {
  const { connId, dbName, tableName } = useParams<{ connId: string; dbName: string; tableName: string }>()
  const [data, setData] = useState<RowsResponse>({ rows: [], total: 0, page: 1, pageSize: 50, columns: [], executionTimeMs: 0 })
  const [structure, setStructure] = useState<TableStructure | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [sortColumn, setSortColumn] = useState('')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
  const [filters, setFilters] = useState<FilterCondition[]>([])
  const [showFilters, setShowFilters] = useState(false)
  const [addCol, setAddCol] = useState('')
  const [addOp, setAddOp] = useState<FilterOperator>('like')
  const [addVal, setAddVal] = useState('')
  const [loading, setLoading] = useState(false)
  const [drawerMode, setDrawerMode] = useState<'insert' | 'edit' | null>(null)
  const [drawerInitial, setDrawerInitial] = useState<Record<string, unknown>>({})
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)
  const [exportOpen, setExportOpen] = useState(false)
  const [exportFormat, setExportFormat] = useState<'csv' | 'sql'>('csv')
  const [exportSchema, setExportSchema] = useState(true)
  const [exportData, setExportData] = useState(true)
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set())
  const [cellDetail, setCellDetail] = useState<{ column: string; value: unknown } | null>(null)
  const selectAllRef = useRef<HTMLInputElement>(null)
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const isReadonly = currentConn?.readonly ?? false

  const pkCols = useMemo(
    () => structure?.columns.filter(c => c.primary_key).map(c => c.name) ?? [],
    [structure]
  )
  const hasPK = pkCols.length > 0

  const load = useCallback(async () => {
    if (!connId || !dbName || !tableName) return
    setLoading(true)
    try {
      const r = await api.listRows(connId, dbName, tableName, {
        page,
        pageSize,
        sortColumn: sortColumn || undefined,
        sortDirection: sortColumn ? sortDirection : undefined,
        filters: filters.length ? filters : undefined,
      })
      setData(r)
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Load failed')
    } finally {
      setLoading(false)
    }
  }, [connId, dbName, tableName, page, pageSize, sortColumn, sortDirection, filters])

  useEffect(() => {
    if (!connId || !dbName || !tableName) return
    api.tableStructure(connId, dbName, tableName).then(setStructure).catch(() => {})
  }, [connId, dbName, tableName])

  useEffect(() => {
    load()
  }, [load])

  // Reset selection when page data changes.
  useEffect(() => {
    setSelectedRows(new Set())
  }, [data.rows])

  // Sync indeterminate state on the select-all checkbox.
  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate =
        selectedRows.size > 0 && selectedRows.size < data.rows.length
    }
  }, [selectedRows.size, data.rows.length])

  function handleSort(col: string) {
    if (sortColumn !== col) {
      setSortColumn(col)
      setSortDirection('asc')
    } else if (sortDirection === 'asc') {
      setSortDirection('desc')
    } else {
      setSortColumn('')
      setSortDirection('asc')
    }
    setPage(1)
  }

  function handleAddFilter() {
    if (!addCol) return
    setFilters(prev => [...prev, { column: addCol, operator: addOp, value: addVal || undefined }])
    setAddCol('')
    setAddOp('like')
    setAddVal('')
    setPage(1)
  }

  function handleRemoveFilter(i: number) {
    setFilters(prev => prev.filter((_, idx) => idx !== i))
    setPage(1)
  }

  function toggleRow(i: number) {
    setSelectedRows(prev => {
      const next = new Set(prev)
      if (next.has(i)) next.delete(i)
      else next.add(i)
      return next
    })
  }

  function toggleAllRows() {
    setSelectedRows(prev =>
      prev.size === data.rows.length
        ? new Set()
        : new Set(data.rows.map((_, i) => i))
    )
  }

  function copySelectedRows(format: 'json' | 'csv' | 'insert') {
    const rows = [...selectedRows].sort((a, b) => a - b).map(i => data.rows[i])
    let text: string
    if (format === 'json') text = toRowJSON(data.columns, rows)
    else if (format === 'csv') text = toCSV(data.columns, rows)
    else text = toInsertSQL(tableName!, data.columns, rows)
    navigator.clipboard.writeText(text).then(
      () => showToast(t('copy.cell_success'), 'success'),
      () => showToast('Copy failed'),
    )
  }

  const noPKTitle = 'No primary key — edit/delete unavailable'
  const pageCount = Math.ceil(data.total / pageSize) || 1

  async function handleSave(values: Record<string, unknown>) {
    setSaving(true)
    try {
      if (drawerMode === 'insert') {
        await api.createRow(connId!, dbName!, tableName!, values)
      } else {
        const primaryKey = Object.fromEntries(pkCols.map(k => [k, drawerInitial[k]]))
        await api.updateRow(connId!, dbName!, tableName!, primaryKey, values)
      }
      setDrawerMode(null)
      load()
      showToast(drawerMode === 'insert' ? 'Row inserted' : 'Row updated', 'success')
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  function handleExportDownload() {
    const url = api.exportURL(connId!, dbName!, tableName!, exportFormat, {
      includeSchema: exportSchema,
      includeData: exportData,
    })
    const a = document.createElement('a')
    a.href = url
    a.download = ''
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setExportOpen(false)
  }

  const needsValue = (op: FilterOperator) => op !== 'is_null' && op !== 'is_not_null'

  const copyBtnClass = 'text-xs px-2 py-0.5 rounded border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'

  return (
    <div className="p-4 flex flex-col h-full">
      {/* Title row */}
      <div className="flex items-center justify-between mb-2">
        <h1 className="text-lg font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2">
          {dbName} / <span className="font-mono">{tableName}</span>
          <span className="text-sm text-gray-400 font-normal">({data.total.toLocaleString()} rows)</span>
          {!hasPK && structure && (
            <span className="text-xs text-amber-500 font-normal" title={noPKTitle}>no PK</span>
          )}
          {loading && <span className="text-xs text-blue-400 font-normal animate-pulse">loading…</span>}
        </h1>

        {/* Action bar */}
        <div className="flex gap-2 flex-wrap justify-end items-center">
          {/* Copy selected rows */}
          {selectedRows.size > 0 && (
            <div className="flex items-center gap-1">
              <span className="text-xs text-gray-500">{t('copy.rows_selected', { n: selectedRows.size })}</span>
              <button onClick={() => copySelectedRows('json')} className={copyBtnClass}>{t('copy.json')}</button>
              <button onClick={() => copySelectedRows('csv')} className={copyBtnClass}>{t('copy.csv')}</button>
              <button onClick={() => copySelectedRows('insert')} className={copyBtnClass}>{t('copy.insert')}</button>
            </div>
          )}
          <button
            onClick={load}
            className="bg-gray-100 dark:bg-gray-700 dark:text-gray-300 text-sm px-3 py-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >{t('data.refresh')}</button>
          <button
            onClick={() => setShowFilters(v => !v)}
            className={`text-sm px-3 py-1 rounded ${showFilters
              ? 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300'
              : 'bg-gray-100 dark:bg-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}`}
          >
            {t('data.filter.show')}{filters.length > 0 && ` (${filters.length})`}
          </button>
          {!isReadonly && (
            <button
              onClick={() => { setDrawerInitial({}); setDrawerMode('insert') }}
              className="bg-green-600 text-white text-sm px-3 py-1 rounded hover:bg-green-700"
            >{t('data.insert')}</button>
          )}
          <button
            onClick={() => setExportOpen(true)}
            className="bg-gray-100 dark:bg-gray-700 dark:text-gray-300 text-sm px-3 py-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          >↓ {t('export.title')}</button>
        </div>
      </div>

      {/* Filter panel */}
      {showFilters && (
        <div className="mb-2 p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm space-y-2">
          {filters.length > 0 && (
            <div className="flex flex-wrap gap-1 items-center">
              {filters.map((f, i) => (
                <span key={i} className="inline-flex items-center gap-1 bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-200 px-2 py-0.5 rounded text-xs">
                  <span className="font-mono">{f.column}</span>
                  <span>{OPERATORS.find(o => o.value === f.operator)?.label ?? f.operator}</span>
                  {f.value && <span className="font-mono">{f.value}</span>}
                  <button onClick={() => handleRemoveFilter(i)} className="ml-1 hover:text-red-500">×</button>
                </span>
              ))}
              <button
                onClick={() => { setFilters([]); setPage(1) }}
                className="text-xs text-gray-500 hover:text-red-500 ml-1"
              >{t('data.filter.clear_all')}</button>
            </div>
          )}
          <div className="flex gap-2 flex-wrap">
            <select
              value={addCol}
              onChange={e => setAddCol(e.target.value)}
              className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-700 dark:text-gray-200 min-w-28"
            >
              <option value="">{t('data.filter.col_placeholder')}</option>
              {data.columns.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
            <select
              value={addOp}
              onChange={e => setAddOp(e.target.value as FilterOperator)}
              className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-700 dark:text-gray-200"
            >
              {OPERATORS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
            </select>
            {needsValue(addOp) && (
              <input
                value={addVal}
                onChange={e => setAddVal(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleAddFilter()}
                placeholder={t('data.filter.val_placeholder')}
                className="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-700 dark:text-gray-200 w-32"
              />
            )}
            <button
              onClick={handleAddFilter}
              disabled={!addCol}
              className="bg-blue-600 text-white text-xs px-3 py-1 rounded hover:bg-blue-700 disabled:opacity-50"
            >{t('data.filter.add')}</button>
          </div>
        </div>
      )}

      {/* Table */}
      <div className="flex-1 overflow-auto border border-gray-200 dark:border-gray-700 rounded-lg">
        <table className="w-full text-xs">
          <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th className="w-8 px-2 py-2">
                <input
                  ref={selectAllRef}
                  type="checkbox"
                  checked={data.rows.length > 0 && selectedRows.size === data.rows.length}
                  onChange={toggleAllRows}
                  className="cursor-pointer"
                  disabled={data.rows.length === 0}
                />
              </th>
              {data.columns.map(col => (
                <th
                  key={col}
                  onClick={() => handleSort(col)}
                  className="text-left px-3 py-2 font-medium text-gray-600 dark:text-gray-400 whitespace-nowrap cursor-pointer select-none hover:bg-gray-100 dark:hover:bg-gray-700"
                >
                  {col}
                  {sortColumn === col && (
                    <span className="ml-1">{sortDirection === 'asc' ? '▲' : '▼'}</span>
                  )}
                </th>
              ))}
              {data.columns.length > 0 && (
                <th className="px-3 py-2 text-right font-medium text-gray-600 dark:text-gray-400 w-20"></th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {data.rows.map((row, ri) => (
              <tr key={ri} className={`hover:bg-blue-50 dark:hover:bg-blue-900/20 ${selectedRows.has(ri) ? 'bg-blue-50/60 dark:bg-blue-900/30' : ''}`}>
                <td className="w-8 px-2" onClick={e => e.stopPropagation()}>
                  <input
                    type="checkbox"
                    checked={selectedRows.has(ri)}
                    onChange={() => toggleRow(ri)}
                    className="cursor-pointer"
                  />
                </td>
                {data.columns.map(col => (
                  <td
                    key={col}
                    title={cellTitle(row[col])}
                    onClick={() => setCellDetail({ column: col, value: row[col] })}
                    className="cursor-pointer px-3 py-1.5 max-w-xs truncate text-gray-700 dark:text-gray-300 hover:bg-blue-50 dark:hover:bg-blue-900/10"
                  >
                    {renderCell(row[col])}
                  </td>
                ))}
                {data.columns.length > 0 && (
                  <td className="px-3 py-1.5">
                    <div className="flex gap-1 justify-end">
                      <button
                        onClick={() => {
                          if (!hasPK || isReadonly) return
                          setDrawerInitial({ ...row })
                          setDrawerMode('edit')
                        }}
                        disabled={!hasPK || isReadonly}
                        title={isReadonly ? t('readonly.badge') : hasPK ? undefined : noPKTitle}
                        className={`text-xs px-2 py-0.5 rounded ${hasPK && !isReadonly
                          ? 'text-blue-500 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-900/30'
                          : 'text-gray-300 dark:text-gray-600 cursor-not-allowed'}`}
                      >{t('data.edit')}</button>
                      <button
                        onClick={() => {
                          if (!hasPK || isReadonly) return
                          const primaryKey = Object.fromEntries(pkCols.map(k => [k, row[k]]))
                          setConfirmState({
                            title: t('data.delete.title'),
                            message: t('data.delete.message', { col: pkCols[0], val: String(row[pkCols[0]]) }),
                            onConfirm: async () => {
                              setConfirmState(null)
                              try {
                                await api.deleteRow(connId!, dbName!, tableName!, primaryKey)
                                load()
                              } catch (e) {
                                showToast(e instanceof Error ? e.message : 'Delete failed')
                              }
                            },
                          })
                        }}
                        disabled={!hasPK || isReadonly}
                        title={isReadonly ? t('readonly.badge') : hasPK ? undefined : noPKTitle}
                        className={`text-xs px-2 py-0.5 rounded ${hasPK && !isReadonly
                          ? 'text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30'
                          : 'text-gray-300 dark:text-gray-600 cursor-not-allowed'}`}
                      >{t('data.delete')}</button>
                    </div>
                  </td>
                )}
              </tr>
            ))}
            {data.rows.length === 0 && (
              <tr>
                <td colSpan={data.columns.length + 2} className="px-3 py-8 text-center text-gray-400">
                  {t('data.no_rows')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination bar */}
      <div className="flex items-center justify-between mt-3 text-sm text-gray-600 dark:text-gray-400">
        <div>Page {page} of {pageCount}</div>
        <div className="flex gap-2">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="px-3 py-1 border border-gray-200 dark:border-gray-600 rounded disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700"
          >{t('data.prev')}</button>
          <button
            onClick={() => setPage(p => Math.min(pageCount, p + 1))}
            disabled={page >= pageCount}
            className="px-3 py-1 border border-gray-200 dark:border-gray-600 rounded disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700"
          >{t('data.next')}</button>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={pageSize}
            onChange={e => { setPageSize(Number(e.target.value)); setPage(1) }}
            className="border border-gray-200 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-800 dark:text-gray-300"
          >
            {[50, 100, 200, 500].map(n => <option key={n} value={n}>{n} / page</option>)}
          </select>
          {data.executionTimeMs > 0 && (
            <span className="text-xs text-gray-400">{data.executionTimeMs}ms</span>
          )}
        </div>
      </div>

      {/* Row insert/edit drawer */}
      {drawerMode && structure && (
        <RowDrawer
          columns={structure.columns}
          pkCols={pkCols}
          mode={drawerMode}
          initialValues={drawerMode === 'edit' ? drawerInitial : undefined}
          saving={saving}
          onSave={handleSave}
          onClose={() => setDrawerMode(null)}
        />
      )}

      <ConfirmDialog
        open={!!confirmState}
        title={confirmState?.title ?? ''}
        message={confirmState?.message ?? ''}
        confirmLabel={t('confirm.delete')}
        dangerous
        onConfirm={() => confirmState?.onConfirm()}
        onCancel={() => setConfirmState(null)}
      />

      {cellDetail && (
        <CellViewer
          column={cellDetail.column}
          value={cellDetail.value}
          onClose={() => setCellDetail(null)}
        />
      )}

      {exportOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-sm mx-4">
            <h2 className="font-bold mb-4 text-gray-900 dark:text-gray-100">{t('export.title')}</h2>
            <div className="mb-4">
              <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('export.format')}</div>
              <div className="flex gap-3">
                {(['csv', 'sql'] as const).map(fmt => (
                  <label key={fmt} className="flex items-center gap-1.5 cursor-pointer text-sm">
                    <input
                      type="radio"
                      name="exportFormat"
                      value={fmt}
                      checked={exportFormat === fmt}
                      onChange={() => setExportFormat(fmt)}
                    />
                    {fmt.toUpperCase()}
                  </label>
                ))}
              </div>
            </div>
            {exportFormat === 'sql' && (
              <div className="mb-4 space-y-2">
                <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700 dark:text-gray-300">
                  <input
                    type="checkbox"
                    checked={exportSchema}
                    onChange={e => setExportSchema(e.target.checked)}
                  />
                  {t('export.include_schema')}
                </label>
                <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700 dark:text-gray-300">
                  <input
                    type="checkbox"
                    checked={exportData}
                    onChange={e => setExportData(e.target.checked)}
                  />
                  {t('export.include_data')}
                </label>
              </div>
            )}
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setExportOpen(false)}
                className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800"
              >{t('data.cancel')}</button>
              <button
                onClick={handleExportDownload}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
              >↓ {t('export.download')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
