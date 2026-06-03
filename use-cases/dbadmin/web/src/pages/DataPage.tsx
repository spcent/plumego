import { useEffect, useState, useCallback, useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { api, errorMessage, type RowsResponse, type TableStructure, type FilterCondition, type FilterOperator } from '../api'
import CellRenderer from '../components/CellRenderer'
import CellViewer from '../components/CellViewer'
import ConfirmDialog from '../components/ConfirmDialog'
import RowDrawer from '../components/RowDrawer'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import { toInsertSQL, toCSV, toRowJSON } from '../utils/copyFormats'
import { XIcon } from '../components/Icons'
import { EmptyStatePanel, ErrorStatePanel, LoadingState, ModalShell, PageBody, PageShell, PageStatusBar, PageToolbar } from '../components/workbench'

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

const DEFAULT_EXPORT_LIMIT = 10000
const MAX_EXPORT_LIMIT = 100000

function cellTitle(v: unknown): string {
  if (v === null || v === undefined) return 'NULL'
  if (v === '') return '(empty string)'
  return String(v)
}

function colWidth(colName: string, structure: TableStructure | null): number {
  if (!structure) return 160
  const col = structure.columns?.find(c => c.name === colName)
  if (!col) return 160
  if (col.primary_key) return 120
  const t = (col.data_type ?? col.full_type ?? '').toLowerCase()
  if (/int|float|double|decimal|numeric|real|bigint|smallint|tinyint/i.test(t)) return 110
  if (/blob|binary|bytea/i.test(t)) return 120
  if (/text|clob/i.test(t)) return 240
  if (/varchar|char/i.test(t)) return 200
  return 160
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
  const [showColumns, setShowColumns] = useState(false)
  const [hiddenColumns, setHiddenColumns] = useState<Set<string>>(new Set())
  const [addCol, setAddCol] = useState('')
  const [addOp, setAddOp] = useState<FilterOperator>('like')
  const [addVal, setAddVal] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [drawerMode, setDrawerMode] = useState<'insert' | 'edit' | null>(null)
  const [drawerInitial, setDrawerInitial] = useState<Record<string, unknown>>({})
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null)
  const [exportOpen, setExportOpen] = useState(false)
  const [exportFormat, setExportFormat] = useState<'csv' | 'sql'>('csv')
  const [exportSchema, setExportSchema] = useState(true)
  const [exportData, setExportData] = useState(true)
  const [exportLimit, setExportLimit] = useState(DEFAULT_EXPORT_LIMIT)
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set())
  const [cellDetail, setCellDetail] = useState<{ column: string; value: unknown } | null>(null)
  const selectAllRef = useRef<HTMLInputElement>(null)
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)
  const isReadonly = currentConn?.readonly ?? false

  const structureColumns = useMemo(() => structure?.columns ?? [], [structure])
  const rowColumns = useMemo(() => data.columns ?? [], [data.columns])
  const dataRows = useMemo(() => data.rows ?? [], [data.rows])
  const totalRows = data.total ?? dataRows.length
  const pkCols = useMemo(
    () => structureColumns.filter(c => c.primary_key).map(c => c.name),
    [structureColumns]
  )
  const hasPK = pkCols.length > 0
  const visibleColumns = useMemo(
    () => rowColumns.filter(col => !hiddenColumns.has(col)),
    [rowColumns, hiddenColumns],
  )

  const load = useCallback(async () => {
    if (!connId || !dbName || !tableName) return
    setLoading(true)
    setLoadError(null)
    try {
      const r = await api.listRows(connId, dbName, tableName, {
        page,
        pageSize,
        sortColumn: sortColumn || undefined,
        sortDirection: sortColumn ? sortDirection : undefined,
        filters: filters.length ? filters : undefined,
      })
      setData(r)
      setSelectedRows(new Set())
    } catch (e) {
      const message = errorMessage(e, t('data.load_failed'))
      setLoadError(message)
      showToast(message)
    } finally {
      setLoading(false)
    }
  }, [connId, dbName, tableName, page, pageSize, sortColumn, sortDirection, filters, showToast, t])

  useEffect(() => {
    if (!connId || !dbName || !tableName) return
    api.tableStructure(connId, dbName, tableName).then(setStructure).catch(() => {})
  }, [connId, dbName, tableName])

  useEffect(() => {
    const id = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(id)
  }, [load])

  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate =
        selectedRows.size > 0 && selectedRows.size < dataRows.length
    }
  }, [selectedRows.size, dataRows.length])

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
      prev.size === dataRows.length
        ? new Set()
        : new Set(dataRows.map((_, i) => i))
    )
  }

  function copySelectedRows(format: 'json' | 'csv' | 'insert') {
    const rows = [...selectedRows].sort((a, b) => a - b).map(i => dataRows[i])
    const columns = visibleColumns.length > 0 ? visibleColumns : rowColumns
    let text: string
    if (format === 'json') text = toRowJSON(columns, rows)
    else if (format === 'csv') text = toCSV(columns, rows)
    else text = toInsertSQL(tableName!, columns, rows)
    navigator.clipboard.writeText(text).then(
      () => showToast(t('copy.cell_success'), 'success'),
      () => showToast(t('data.copy_failed')),
    )
  }

  const noPKTitle = t('data.no_pk_hint')
  const pageCount = Math.ceil(totalRows / pageSize) || 1

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
      showToast(drawerMode === 'insert' ? t('data.insert.success') : t('data.edit.success'), 'success')
    } catch (e) {
      showToast(errorMessage(e, t('data.save_failed')))
    } finally {
      setSaving(false)
    }
  }

  function handleExportDownload() {
    const url = api.exportURL(connId!, dbName!, tableName!, exportFormat, {
      includeSchema: exportSchema,
      includeData: exportData,
      limit: exportLimit,
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

  const colType = (colName: string) =>
    structureColumns.find(c => c.name === colName)?.data_type

  function toggleColumn(col: string) {
    setHiddenColumns(prev => {
      const next = new Set(prev)
      if (next.has(col)) {
        next.delete(col)
        return next
      }
      if (rowColumns.length - next.size <= 1) return next
      next.add(col)
      return next
    })
  }

  return (
    <PageShell>
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={dbName && tableName ? [dbName, tableName] : []}
        datasourceType={currentConn?.driver ?? 'mysql'}
        readonly={isReadonly}
        onRefresh={load}
        meta={{ rowCount: totalRows }}
      />
      <PageToolbar
        leading={
          <>
            {loading && (
              <span className="badge" style={{ color: 'var(--accent)' }}>
                <span className="status-dot animate-pulse" style={{ background: 'var(--accent)' }} />
                {t('data.loading')}
              </span>
            )}
          </>
        }
        trailing={
          <>
          <button
            onClick={load}
            className="btn btn-ghost h-8 px-3 text-xs"
          >
            {t('data.refresh')}
          </button>
          <button
            onClick={() => setShowFilters(v => !v)}
            className={`btn h-8 px-3 text-xs ${showFilters ? 'btn-primary' : 'btn-ghost'}`}
          >
            {t('data.filter.show')}{filters.length > 0 && ` (${filters.length})`}
          </button>
          <button
            onClick={() => setShowColumns(v => !v)}
            className={`btn h-8 px-3 text-xs ${showColumns ? 'btn-primary' : 'btn-ghost'}`}
          >
            {t('data.columns.show')} ({visibleColumns.length})
          </button>
          {!isReadonly && (
            <button
              onClick={() => { setDrawerInitial({}); setDrawerMode('insert') }}
              className="btn h-8 px-3 text-xs"
              style={{ background: 'var(--success)', color: '#fff', borderColor: 'transparent' }}
            >
              {t('data.insert')}
            </button>
          )}
          <button
            onClick={() => setExportOpen(true)}
            className="btn btn-ghost h-8 px-3 text-xs"
          >
            {t('export.title')}
          </button>
          </>
        }
      />

      <PageBody>
        <div className="flex h-full flex-col">
      {/* ── Filter panel ─────────────────────────────────── */}
      {(showFilters || showColumns) && (
        <div
          className="shrink-0 space-y-2 px-4 py-3"
          style={{
            background: 'var(--bg-muted)',
            borderBottom: '1px solid var(--border-subtle)',
          }}
        >
          {showFilters && filters.length > 0 && (
            <div className="flex flex-wrap gap-1 items-center">
              {filters.map((f, i) => (
                <span
                  key={i}
                  className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px]"
                  style={{ background: 'var(--bg-selected)', color: 'var(--accent)', border: '1px solid var(--accent)44' }}
                >
                  <span className="font-mono">{f.column}</span>
                  <span>{OPERATORS.find(o => o.value === f.operator)?.label ?? f.operator}</span>
                  {f.value && <span className="font-mono">{f.value}</span>}
                  <button
                    onClick={() => handleRemoveFilter(i)}
                    className="ml-1 grid h-4 w-4 place-items-center rounded hover:bg-[var(--bg-hover)]"
                    aria-label={t('data.filter.remove')}
                  >
                    <XIcon className="h-3 w-3" />
                  </button>
                </span>
              ))}
              <button
                onClick={() => { setFilters([]); setPage(1) }}
                className="text-[11px] hover:opacity-60"
                style={{ color: 'var(--danger)' }}
              >
                {t('data.filter.clear_all')}
              </button>
            </div>
          )}
          {showFilters && (
          <div className="flex gap-2 flex-wrap items-center">
            <select
              value={addCol}
              onChange={e => setAddCol(e.target.value)}
              className="select min-w-32 text-xs"
            >
              <option value="">{t('data.filter.col_placeholder')}</option>
              {rowColumns.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
            <select
              value={addOp}
              onChange={e => setAddOp(e.target.value as FilterOperator)}
              className="select w-auto min-w-28 text-xs"
            >
              {OPERATORS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
            </select>
            {needsValue(addOp) && (
              <input
                value={addVal}
                onChange={e => setAddVal(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleAddFilter()}
                placeholder={t('data.filter.val_placeholder')}
                className="input w-40 text-xs"
              />
            )}
            <button
              onClick={handleAddFilter}
              disabled={!addCol}
              className="btn btn-primary h-8 px-3 text-xs disabled:opacity-40"
            >
              {t('data.filter.add')}
            </button>
          </div>
          )}

          {showColumns && (
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-[11px] font-medium uppercase tracking-wide" style={{ color: 'var(--text-muted)' }}>
                {t('data.columns.visible')}
              </span>
              {rowColumns.map(col => (
                <label
                  key={col}
                  className="inline-flex h-7 cursor-pointer items-center gap-1.5 rounded-md border px-2 text-[11px]"
                  style={{
                    borderColor: hiddenColumns.has(col) ? 'var(--border-subtle)' : 'var(--accent)',
                    color: hiddenColumns.has(col) ? 'var(--text-muted)' : 'var(--accent)',
                    background: hiddenColumns.has(col) ? 'transparent' : 'var(--bg-selected)',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={!hiddenColumns.has(col)}
                    onChange={() => toggleColumn(col)}
                  />
                  <span className="font-mono">{col}</span>
                </label>
              ))}
              {hiddenColumns.size > 0 && (
                <button
                  type="button"
                  onClick={() => setHiddenColumns(new Set())}
                  className="btn btn-ghost h-7 px-2 text-xs"
                >
                  {t('data.columns.reset')}
                </button>
              )}
            </div>
          )}
        </div>
      )}

      {/* ── Table ────────────────────────────────────────── */}
      {loading && dataRows.length === 0 ? (
        <LoadingState title={t('data.loading')} detail={tableName} />
      ) : loadError ? (
        <ErrorStatePanel
          title={t('data.load_failed')}
          message={loadError}
          action={
            <button type="button" onClick={() => void load()} className="btn btn-ghost h-8 px-3 text-xs">
              {t('resource.retry')}
            </button>
          }
        />
      ) : dataRows.length === 0 ? (
        <EmptyStatePanel title={t('data.no_rows')} detail={filters.length > 0 ? t('data.empty.filtered') : tableName} />
      ) : (
      <div className="data-table-wrap flex-1">
        <table className="data-table">
          <thead>
            <tr>
              {/* Checkbox col */}
              <th style={{ width: 36, padding: '0 8px' }}>
                <input
                  ref={selectAllRef}
                  type="checkbox"
                  checked={dataRows.length > 0 && selectedRows.size === dataRows.length}
                  onChange={toggleAllRows}
                  className="cursor-pointer"
                  disabled={dataRows.length === 0}
                />
              </th>

              {visibleColumns.map(col => {
                const structCol = structureColumns.find(c => c.name === col)
                const isPK = structCol?.primary_key ?? false
                const typeLabel = structCol?.data_type ?? structCol?.full_type ?? ''
                const isSorted = sortColumn === col
                const w = colWidth(col, structure)

                return (
                  <th
                    key={col}
                    onClick={() => handleSort(col)}
                    className="text-left cursor-pointer select-none whitespace-nowrap"
                    style={{
                      width: w,
                      minWidth: w,
                      padding: '6px 8px',
                      color: 'var(--text-muted)',
                      fontWeight: 500,
                    }}
                  >
                    <div className="flex items-center gap-1">
                      {isPK && (
                        <span
                          className="shrink-0 text-[9px] font-mono px-0.5 rounded leading-none"
                          style={{ background: 'var(--warning)22', color: 'var(--warning)' }}
                        >
                          PK
                        </span>
                      )}
                      <span>{col}</span>
                      {isSorted && (
                        <span className="badge h-4 min-h-0 px-1 text-[9px] uppercase">
                          {sortDirection}
                        </span>
                      )}
                    </div>
                    {typeLabel && (
                      <div
                        className="text-[10px] font-normal font-mono mt-px truncate"
                        style={{ color: 'var(--text-subtle)', maxWidth: w - 16 }}
                      >
                        {typeLabel}
                      </div>
                    )}
                  </th>
                )
              })}

              {/* Actions col */}
              {visibleColumns.length > 0 && (
                <th style={{ width: 120, padding: '6px 8px' }} />
              )}
            </tr>
          </thead>

          <tbody>
            {dataRows.map((row, ri) => (
              <tr
                key={ri}
                style={{
                  height: 34,
                  background: selectedRows.has(ri)
                    ? 'var(--bg-selected)'
                    : 'transparent',
                }}
              >
                <td style={{ width: 36, padding: '0 8px' }} onClick={e => e.stopPropagation()}>
                  <input
                    type="checkbox"
                    checked={selectedRows.has(ri)}
                    onChange={() => toggleRow(ri)}
                    className="cursor-pointer"
                  />
                </td>

                {visibleColumns.map(col => {
                  const w = colWidth(col, structure)
                  return (
                    <td
                      key={col}
                      title={cellTitle(row[col])}
                      onClick={() => setCellDetail({ column: col, value: row[col] })}
                      className="cursor-pointer overflow-hidden"
                      style={{
                        width: w,
                        maxWidth: w,
                        padding: '6px 8px',
                        color: 'var(--text-default)',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      <CellRenderer value={row[col]} colType={colType(col)} />
                    </td>
                  )
                })}

                {visibleColumns.length > 0 && (
                  <td style={{ width: 120, padding: '6px 8px' }}>
                    <div className="flex gap-1 justify-end">
                      <button
                        onClick={() => {
                          if (!hasPK || isReadonly) return
                          setDrawerInitial({ ...row })
                          setDrawerMode('edit')
                        }}
                        disabled={!hasPK || isReadonly}
                        title={isReadonly ? t('readonly.badge') : hasPK ? undefined : noPKTitle}
                        className="btn btn-ghost h-6 px-1.5 text-[11px]"
                        style={hasPK && !isReadonly
                          ? { color: 'var(--accent)' }
                          : { color: 'var(--text-subtle)', cursor: 'not-allowed' }
                        }
                      >
                        {t('data.edit')}
                      </button>
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
                                showToast(errorMessage(e, t('data.delete_failed')))
                              }
                            },
                          })
                        }}
                        disabled={!hasPK || isReadonly}
                        title={isReadonly ? t('readonly.badge') : hasPK ? undefined : noPKTitle}
                        className="btn btn-ghost h-6 px-1.5 text-[11px]"
                        style={hasPK && !isReadonly
                          ? { color: 'var(--danger)' }
                          : { color: 'var(--text-subtle)', cursor: 'not-allowed' }
                        }
                      >
                        {t('data.delete')}
                      </button>
                    </div>
                  </td>
                )}
              </tr>
            ))}

          </tbody>
        </table>
      </div>
      )}
      {selectedRows.size > 0 && (
        <div className="shrink-0 border-t px-4 py-2" style={{ borderColor: 'var(--border-subtle)', background: 'var(--bg-surface)' }}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="text-[12px] font-medium" style={{ color: 'var(--text-default)' }}>
              {t('copy.rows_selected', { n: selectedRows.size })}
            </div>
            <div className="flex flex-wrap items-center gap-1">
              <button onClick={() => copySelectedRows('json')} className="btn btn-ghost h-7 px-2 text-xs">{t('copy.json')}</button>
              <button onClick={() => copySelectedRows('csv')} className="btn btn-ghost h-7 px-2 text-xs">{t('copy.csv')}</button>
              <button onClick={() => copySelectedRows('insert')} className="btn btn-ghost h-7 px-2 text-xs">{t('copy.insert')}</button>
              <button onClick={() => setSelectedRows(new Set())} className="btn btn-ghost h-7 px-2 text-xs">{t('data.selection.clear')}</button>
            </div>
          </div>
        </div>
      )}
        </div>
      </PageBody>

      {/* ── Pagination bar ───────────────────────────────── */}
      <PageStatusBar
        left={t('data.pagination', { page, total: pageCount })}
        center={
          <div className="flex gap-1.5">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="btn btn-ghost h-7 px-2 text-xs disabled:opacity-40"
            >
              {t('data.prev')}
            </button>
            <button
              onClick={() => setPage(p => Math.min(pageCount, p + 1))}
              disabled={page >= pageCount}
              className="btn btn-ghost h-7 px-2 text-xs disabled:opacity-40"
            >
              {t('data.next')}
            </button>
          </div>
        }
        right={
          <div className="inline-flex items-center gap-3">
            <select
              value={pageSize}
              onChange={e => { setPageSize(Number(e.target.value)); setPage(1) }}
              className="select h-7 min-h-0 w-auto py-0 text-xs"
            >
              {[50, 100, 200, 500].map(n => <option key={n} value={n}>{t('data.page_size', { n })}</option>)}
            </select>
            {data.executionTimeMs > 0 && <span>{data.executionTimeMs}ms</span>}
          </div>
        }
      />

      {/* ── Row drawer ──────────────────────────────────── */}
      {drawerMode && structure && (
        <RowDrawer
          key={`${drawerMode}:${structureColumns.map(c => c.name).join('|')}:${pkCols.map(k => String(drawerInitial[k] ?? '')).join('|')}`}
          columns={structureColumns}
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
        <ModalShell
          title={t('export.title')}
          widthClass="max-w-sm"
          onClose={() => setExportOpen(false)}
          footer={
            <>
              <button
                onClick={() => setExportOpen(false)}
                className="btn btn-ghost h-8 px-3 text-xs"
              >
                {t('data.cancel')}
              </button>
              <button
                onClick={handleExportDownload}
                className="btn btn-primary h-8 px-3 text-xs"
              >
                {t('export.download')}
              </button>
            </>
          }
        >
            <div>
              <div className="text-[12px] font-medium mb-1" style={{ color: 'var(--text-muted)' }}>
                {t('export.format')}
              </div>
              <div className="flex gap-3">
                {(['csv', 'sql'] as const).map(fmt => (
                  <label key={fmt} className="flex items-center gap-1.5 cursor-pointer text-[12px]" style={{ color: 'var(--text-default)' }}>
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
                <label className="flex items-center gap-2 cursor-pointer text-[12px]" style={{ color: 'var(--text-default)' }}>
                  <input type="checkbox" checked={exportSchema} onChange={e => setExportSchema(e.target.checked)} />
                  {t('export.include_schema')}
                </label>
                <label className="flex items-center gap-2 cursor-pointer text-[12px]" style={{ color: 'var(--text-default)' }}>
                  <input type="checkbox" checked={exportData} onChange={e => setExportData(e.target.checked)} />
                  {t('export.include_data')}
                </label>
              </div>
            )}
            <div className="mt-4">
              <label className="mb-1 block text-[12px] font-medium" style={{ color: 'var(--text-muted)' }}>
                {t('export.limit')}
              </label>
              <input
                type="number"
                min={1}
                max={MAX_EXPORT_LIMIT}
                value={exportLimit}
                onChange={e => setExportLimit(Math.min(MAX_EXPORT_LIMIT, Math.max(1, Number(e.target.value) || DEFAULT_EXPORT_LIMIT)))}
                className="input h-8 text-xs"
              />
              <p className="mt-1 text-[11px]" style={{ color: 'var(--text-subtle)' }}>
                {t('export.limit_hint', { max: MAX_EXPORT_LIMIT })}
              </p>
            </div>
        </ModalShell>
      )}
    </PageShell>
  )
}
