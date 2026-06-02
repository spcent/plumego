import { useEffect, useRef, useState } from 'react'
import type { ColumnInfo } from '../api'
import { useI18n } from '../i18nContext'
import { XIcon } from './Icons'

interface RowDrawerProps {
  columns: ColumnInfo[]
  pkCols: string[]
  mode: 'insert' | 'edit'
  initialValues?: Record<string, unknown>
  saving?: boolean
  onSave: (values: Record<string, unknown>) => Promise<void>
  onClose: () => void
}

type FieldState = {
  value: string
  isNull: boolean
  include: boolean // insert only: whether to send this field at all
}

function initFields(
  columns: ColumnInfo[],
  pkCols: string[],
  mode: 'insert' | 'edit',
  initialValues?: Record<string, unknown>,
): Record<string, FieldState> {
  const fields: Record<string, FieldState> = {}
  for (const col of columns) {
    const isPK = pkCols.includes(col.name)
    const raw = initialValues?.[col.name]
    const isNull = raw === null || raw === undefined
    // In insert mode, auto-increment PKs are excluded by default
    const include = mode === 'insert' ? !(isPK && col.auto_increment) : true
    fields[col.name] = {
      value: raw === null || raw === undefined ? '' : String(raw),
      isNull: mode === 'edit' ? isNull : false,
      include,
    }
  }
  return fields
}

function inputType(col: ColumnInfo): 'number' | 'checkbox' | 'textarea' | 'datetime' | 'date' | 'time' | 'text' {
  const dt = (col.data_type || '').toLowerCase()
  const ft = (col.full_type || '').toLowerCase()
  if (ft === 'tinyint(1)' || dt === 'boolean' || dt === 'bool') return 'checkbox'
  if (dt === 'tinyint' || dt === 'smallint' || dt === 'mediumint' || dt === 'int' ||
      dt === 'integer' || dt === 'bigint' || dt === 'float' || dt === 'double' ||
      dt === 'decimal' || dt === 'numeric' || dt === 'real') return 'number'
  if (dt === 'text' || dt === 'mediumtext' || dt === 'longtext' ||
      dt === 'tinytext' || dt === 'blob' || dt === 'clob') return 'textarea'
  if (dt === 'datetime' || dt === 'timestamp') return 'datetime'
  if (dt === 'date') return 'date'
  if (dt === 'time') return 'time'
  return 'text'
}

function isChanged(col: string, fields: Record<string, FieldState>, initialValues?: Record<string, unknown>): boolean {
  if (!initialValues) return false
  const f = fields[col]
  const orig = initialValues[col]
  if (f.isNull) return orig !== null && orig !== undefined
  const origStr = orig === null || orig === undefined ? '' : String(orig)
  return f.value !== origStr
}

export default function RowDrawer({ columns, pkCols, mode, initialValues, saving, onSave, onClose }: RowDrawerProps) {
  const [fields, setFields] = useState<Record<string, FieldState>>(() =>
    initFields(columns, pkCols, mode, initialValues)
  )
  const [error, setError] = useState('')
  const { t } = useI18n()
  const firstInputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  useEffect(() => {
    firstInputRef.current?.focus()
  }, [])

  function setFieldProp(col: string, prop: keyof FieldState, val: unknown) {
    setFields(prev => ({ ...prev, [col]: { ...prev[col], [prop]: val } }))
  }

  async function handleSubmit() {
    setError('')
    const values: Record<string, unknown> = {}
    for (const col of columns) {
      const f = fields[col.name]
      if (mode === 'insert' && !f.include) continue
      if (f.isNull) {
        values[col.name] = null
      } else {
        const kind = inputType(col)
        if (kind === 'number') {
          values[col.name] = f.value === '' ? null : Number(f.value)
        } else if (kind === 'checkbox') {
          values[col.name] = f.value === 'true' || f.value === '1'
        } else {
          values[col.name] = f.value
        }
      }
    }
    try {
      await onSave(values)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Save failed')
    }
  }

  const inputCls = 'input'
  const disabledCls = 'opacity-40 pointer-events-none'

  const title = mode === 'insert' ? t('data.insert.title') : t('data.edit.title')
  const visibleCols = mode === 'insert'
    ? columns.filter(c => !(pkCols.includes(c.name) && c.auto_increment))
    : columns

  return (
    <div className="fixed inset-0 z-50 flex">
      {/* Backdrop */}
      <div className="flex-1 bg-black/40" onClick={onClose} />

      {/* Drawer panel */}
      <div className="flex w-96 flex-col overflow-hidden border-l" style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)', boxShadow: 'var(--shadow-lg)' }}>
        {/* Header */}
        <div className="flex shrink-0 items-center justify-between border-b px-4 py-3" style={{ borderColor: 'var(--border-subtle)' }}>
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>{title}</h2>
          <button onClick={onClose} className="icon-btn" aria-label="Close">
            <XIcon className="h-4 w-4" />
          </button>
        </div>

        {/* Fields */}
        <div className="flex-1 overflow-y-auto px-4 py-3 space-y-3">
          {error && (
            <div
              className="rounded-md border p-2 text-xs"
              style={{ background: 'var(--danger-soft)', borderColor: 'var(--danger-border)', color: 'var(--danger)' }}
            >
              {error}
            </div>
          )}

          {visibleCols.map(col => {
            const f = fields[col.name]
            if (!f) return null
            const isPK = pkCols.includes(col.name)
            const kind = inputType(col)
            const changed = mode === 'edit' && isChanged(col.name, fields, initialValues)
            const fieldDisabled = f.isNull

            return (
              <div
                key={col.name}
                className={`${changed ? 'border-l-2 border-amber-400 pl-2' : ''}`}
              >
                {/* Label row */}
                <div className="flex items-center justify-between mb-0.5">
                  <label className="flex-1 truncate text-xs font-medium" style={{ color: 'var(--text-default)' }}>
                    {col.name}
                    {isPK && <span className="ml-1 text-amber-500 text-xs">PK</span>}
                    {col.auto_increment && <span className="ml-1 text-xs" style={{ color: 'var(--text-subtle)' }}>AUTO</span>}
                    <span className="ml-1 text-xs font-normal" style={{ color: 'var(--text-subtle)' }}>{col.full_type}</span>
                  </label>
                  <div className="flex items-center gap-2 ml-2 shrink-0">
                    {/* Insert-mode include toggle (non-PK fields only) */}
                    {mode === 'insert' && (
                      <label className="flex cursor-pointer items-center gap-1 text-xs" style={{ color: 'var(--text-muted)' }}>
                        <input
                          type="checkbox"
                          checked={f.include}
                          onChange={e => setFieldProp(col.name, 'include', e.target.checked)}
                          className="rounded"
                        />
                        include
                      </label>
                    )}
                    {/* NULL checkbox for nullable fields */}
                    {col.nullable && (
                      <label className="flex cursor-pointer items-center gap-1 text-xs" style={{ color: 'var(--text-muted)' }}>
                        <input
                          type="checkbox"
                          checked={f.isNull}
                          onChange={e => setFieldProp(col.name, 'isNull', e.target.checked)}
                          className="rounded"
                        />
                        NULL
                      </label>
                    )}
                  </div>
                </div>

                {/* Input area - hidden for insert when !include */}
                {(mode === 'edit' || f.include) && (
                  <>
                    {kind === 'textarea' ? (
                      <textarea
                        value={f.value}
                        onChange={e => setFieldProp(col.name, 'value', e.target.value)}
                        disabled={fieldDisabled}
                        rows={3}
                        className={`${inputCls} resize-y ${fieldDisabled ? disabledCls : ''}`}
                      />
                    ) : kind === 'checkbox' ? (
                      <div className="flex items-center gap-2 py-1">
                        <input
                          type="checkbox"
                          checked={f.value === 'true' || f.value === '1'}
                          onChange={e => setFieldProp(col.name, 'value', e.target.checked ? 'true' : 'false')}
                          disabled={fieldDisabled}
                          className={`rounded ${fieldDisabled ? disabledCls : ''}`}
                        />
                        <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{f.value === 'true' || f.value === '1' ? 'true' : 'false'}</span>
                      </div>
                    ) : (
                      <input
                        ref={visibleCols.indexOf(col) === 0 ? firstInputRef : undefined}
                        type={kind === 'number' ? 'number' : 'text'}
                        step={kind === 'number' ? 'any' : undefined}
                        value={f.value}
                        onChange={e => setFieldProp(col.name, 'value', e.target.value)}
                        disabled={fieldDisabled}
                        placeholder={
                          kind === 'datetime' ? 'YYYY-MM-DD HH:MM:SS' :
                          kind === 'date' ? 'YYYY-MM-DD' :
                          kind === 'time' ? 'HH:MM:SS' :
                          undefined
                        }
                        className={`${inputCls} ${fieldDisabled ? disabledCls : ''}`}
                      />
                    )}
                    {/* Diff hint in edit mode */}
                    {changed && initialValues && (
                      <div className="mt-0.5 truncate text-xs" style={{ color: 'var(--warning)' }}>
                        was: {initialValues[col.name] === null ? 'NULL' : String(initialValues[col.name])}
                      </div>
                    )}
                  </>
                )}
              </div>
            )
          })}
        </div>

        {/* Footer */}
        <div className="flex shrink-0 justify-end gap-2 border-t px-4 py-3" style={{ borderColor: 'var(--border-subtle)' }}>
          <button
            onClick={onClose}
            className="btn btn-ghost"
          >{t('data.cancel')}</button>
          <button
            onClick={handleSubmit}
            disabled={saving}
            className="btn btn-primary disabled:opacity-50"
          >{saving ? '…' : t('data.save')}</button>
        </div>
      </div>
    </div>
  )
}
