import { useEffect, useState, useCallback, useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { api, type RowsResponse, type TableStructure, type FilterCondition, type FilterOperator } from '../api'
import CellRenderer from '../components/CellRenderer'
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

function cellTitle(v: unknown): string {
  if (v === null || v === undefined) return 'NULL'
  if (v === '') return '(empty string)'
  return String(v)
}

function colWidth(colName: string, structure: TableStructure | null): number {
  if (!structure) return 160
  const col = structure.columns.find(c => c.name === colName)
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

  useEffect(() => {
    setSelectedRows(new Set())
  }, [data.rows])

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

  const copyBtnClass = 'text-[12px] px-2 py-0.5 rounded border hover:opacity-80 transition-opacity'
  const copyBtnStyle = { borderColor: 'var(--border-strong)', color: 'var(--text-muted)' }

  const colType = (colName: string) =>
    structure?.columns.find(c => c.name === colName)?.data_type

  return (
    <div
      className="flex flex-col h-full"
      style={{ background: 'var(--bg-surface)' }}
    >
      {/* ── Title / toolbar ──────────────────────────────── */}
      <div
        className="flex items-center justify-between px-4 py-2 shrink-0"
        style={{ borderBottom: '1px solid var(--border-subtle)' }}
      >
        <h1
          className="text-sm font-semibold flex items-center gap-2 min-w-0"
          style={{ color: 'var(--text-strong)' }}
        >
          <span className="truncate font-mono">{tableName}</span>
          <span className="text-xs font-normal shrink-0" style={{ color: 'var(--text-muted)' }}>
            ({data.total.toLocaleString()} rows)
          </span>
          {!hasPK && structure && (
            <span className="text-[11px] shrink-0" style={{ color: 'var(--warning)' }} title={noPKTitle}>
              no PK
            </span>
          )}
          {loading && (
            <span className="text-[11px] shrink-0 animate-pulse" style={{ color: 'var(--accent)' }}>
              loading…
            </span>
          )}
        </h1>

        <div className="flex gap-1.5 flex-wrap justify-end items-center ml-4 shrink-0">
          {selectedRows.size > 0 && (
            <div className="flex items-center gap-1">
              <span className="text-[11px]" style={{ color: 'var(--text-muted)' }}>
                {t('copy.rows_selected', { n: selectedRows.size })}
              </span>
              <button onClick={() => copySelectedRows('json')} className={copyBtnClass} style={copyBtnStyle}>{t('copy.json')}</button>
              <button onClick={() => copySelectedRows('csv')} className={copyBtnClass} style={copyBtnStyle}>{t('copy.csv')}</button>
              <button onClick={() => copySelectedRows('insert')} className={copyBtnClass} style={copyBtnStyle}>{t('copy.insert')}</button>
            </div>
          )}

          <button
            onClick={load}
            className="text-[12px] px-2.5 py-1 rounded hover:opacity-80 transition-opacity"
            style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)', border: '1px solid var(--border-subtle)' }}
          >
            {t('data.refresh')}
          </button>
          <button
            onClick={() => setShowFilters(v => !v)}
            className="text-[12px] px-2.5 py-1 rounded transition-colors"
            style={showFilters
              ? { background: 'var(--accent)', color: '#fff', border: '1px solid var(--accent)' }
              : { background: 'var(--bg-muted)', color: 'var(--text-muted)', border: '1px solid var(--border-subtle)' }
            }
          >
            {t('data.filter.show')}{filters.length > 0 && ` (${filters.length})`}
          </button>
          {!isReadonly && (
            <button
              onClick={() => { setDrawerInitial({}); setDrawerMode('insert') }}
              className="text-[12px] px-2.5 py-1 rounded hover:opacity-80 transition-opacity"
              style={{ background: 'var(--success)', color: '#fff' }}
            >
              {t('data.insert')}
            </button>
          )}
          <button
            onClick={() => setExportOpen(true)}
            className="text-[12px] px-2.5 py-1 rounded hover:opacity-80 transition-opacity"
            style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)', border: '1px solid var(--border-subtle)' }}
          >
            ↓ {t('export.title')}
          </button>
        </div>
      </div>

      {/* ── Filter panel ─────────────────────────────────── */}
      {showFilters && (
        <div
          className="px-4 py-2 shrink-0 space-y-2"
          style={{
            background: 'var(--bg-muted)',
            borderBottom: '1px solid var(--border-subtle)',
          }}
        >
          {filters.length > 0 && (
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
                    className="ml-1 hover:opacity-60"
                  >
                    ×
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
          <div className="flex gap-2 flex-wrap items-center">
            <select
              value={addCol}
              onChange={e => setAddCol(e.target.value)}
              className="rounded px-2 py-1 text-[12px] min-w-28"
              style={{
                border: '1px solid var(--border-strong)',
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
              }}
            >
              <option value="">{t('data.filter.col_placeholder')}</option>
              {data.columns.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
            <select
              value={addOp}
              onChange={e => setAddOp(e.target.value as FilterOperator)}
              className="rounded px-2 py-1 text-[12px]"
              style={{
                border: '1px solid var(--border-strong)',
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
              }}
            >
              {OPERATORS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
            </select>
            {needsValue(addOp) && (
              <input
                value={addVal}
                onChange={e => setAddVal(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleAddFilter()}
                placeholder={t('data.filter.val_placeholder')}
                className="rounded px-2 py-1 text-[12px] w-32"
                style={{
                  border: '1px solid var(--border-strong)',
                  background: 'var(--bg-surface)',
                  color: 'var(--text-default)',
                }}
              />
            )}
            <button
              onClick={handleAddFilter}
              disabled={!addCol}
              className="text-[12px] px-3 py-1 rounded disabled:opacity-40"
              style={{ background: 'var(--accent)', color: '#fff' }}
            >
              {t('data.filter.add')}
            </button>
          </div>
        </div>
      )}

      {/* ── Table ────────────────────────────────────────── */}
      <div
        className="flex-1 overflow-auto min-h-0"
        style={{ borderBottom: '1px solid var(--border-subtle)' }}
      >
        <table
          className="border-collapse"
          style={{ minWidth: 'max-content', width: '100%', fontSize: 12 }}
        >
          <thead style={{ background: 'var(--bg-muted)', position: 'sticky', top: 0, zIndex: 10 }}>
            <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
              {/* Checkbox col */}
              <th style={{ width: 36, padding: '0 8px' }}>
                <input
                  ref={selectAllRef}
                  type="checkbox"
                  checked={data.rows.length > 0 && selectedRows.size === data.rows.length}
                  onChange={toggleAllRows}
                  className="cursor-pointer"
                  disabled={data.rows.length === 0}
                />
              </th>

              {data.columns.map(col => {
                const structCol = structure?.columns.find(c => c.name === col)
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
                        <span className="shrink-0 text-[10px]">
                          {sortDirection === 'asc' ? '▲' : '▼'}
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
              {data.columns.length > 0 && (
                <th style={{ width: 120, padding: '6px 8px' }} />
              )}
            </tr>
          </thead>

          <tbody>
            {data.rows.map((row, ri) => (
              <tr
                key={ri}
                style={{
                  height: 34,
                  background: selectedRows.has(ri)
                    ? 'var(--bg-selected)'
                    : 'transparent',
                  borderBottom: '1px solid var(--border-subtle)',
                }}
                className="hover:bg-[var(--bg-hover)]"
              >
                <td style={{ width: 36, padding: '0 8px' }} onClick={e => e.stopPropagation()}>
                  <input
                    type="checkbox"
                    checked={selectedRows.has(ri)}
                    onChange={() => toggleRow(ri)}
                    className="cursor-pointer"
                  />
                </td>

                {data.columns.map(col => {
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

                {data.columns.length > 0 && (
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
                        className="text-[11px] px-1.5 py-0.5 rounded transition-colors"
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
                                showToast(e instanceof Error ? e.message : 'Delete failed')
                              }
                            },
                          })
                        }}
                        disabled={!hasPK || isReadonly}
                        title={isReadonly ? t('readonly.badge') : hasPK ? undefined : noPKTitle}
                        className="text-[11px] px-1.5 py-0.5 rounded transition-colors"
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

            {data.rows.length === 0 && (
              <tr>
                <td
                  colSpan={data.columns.length + 2}
                  className="py-12 text-center text-[12px]"
                  style={{ color: 'var(--text-subtle)' }}
                >
                  {t('data.no_rows')}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* ── Pagination bar ───────────────────────────────── */}
      <div
        className="flex items-center justify-between px-4 py-2 shrink-0 text-[12px]"
        style={{
          background: 'var(--bg-muted)',
          borderTop: '1px solid var(--border-subtle)',
          color: 'var(--text-muted)',
        }}
      >
        <div>Page {page} of {pageCount}</div>
        <div className="flex gap-1.5">
          <button
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="px-2.5 py-1 rounded disabled:opacity-40 hover:opacity-80 transition-opacity"
            style={{ border: '1px solid var(--border-subtle)', background: 'var(--bg-surface)' }}
          >
            {t('data.prev')}
          </button>
          <button
            onClick={() => setPage(p => Math.min(pageCount, p + 1))}
            disabled={page >= pageCount}
            className="px-2.5 py-1 rounded disabled:opacity-40 hover:opacity-80 transition-opacity"
            style={{ border: '1px solid var(--border-subtle)', background: 'var(--bg-surface)' }}
          >
            {t('data.next')}
          </button>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={pageSize}
            onChange={e => { setPageSize(Number(e.target.value)); setPage(1) }}
            className="rounded px-2 py-1 text-[12px]"
            style={{
              border: '1px solid var(--border-subtle)',
              background: 'var(--bg-surface)',
              color: 'var(--text-default)',
            }}
          >
            {[50, 100, 200, 500].map(n => <option key={n} value={n}>{n} / page</option>)}
          </select>
          {data.executionTimeMs > 0 && (
            <span style={{ color: 'var(--text-subtle)' }}>{data.executionTimeMs}ms</span>
          )}
        </div>
      </div>

      {/* ── Row drawer ──────────────────────────────────── */}
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
          <div
            className="rounded-lg shadow-xl p-6 w-full max-w-sm mx-4"
            style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
          >
            <h2
              className="font-bold mb-4 text-sm"
              style={{ color: 'var(--text-strong)' }}
            >
              {t('export.title')}
            </h2>
            <div className="mb-4">
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
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setExportOpen(false)}
                className="px-4 py-1.5 text-[12px] hover:opacity-70 transition-opacity"
                style={{ color: 'var(--text-muted)' }}
              >
                {t('data.cancel')}
              </button>
              <button
                onClick={handleExportDownload}
                className="px-4 py-1.5 text-[12px] rounded hover:opacity-80 transition-opacity"
                style={{ background: 'var(--accent)', color: '#fff' }}
              >
                ↓ {t('export.download')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
