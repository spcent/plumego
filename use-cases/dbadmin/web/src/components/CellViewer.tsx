import { useEffect, useState } from 'react'
import { useToast } from './Toast'
import { useI18n } from '../i18n'

type ContentType = 'null' | 'empty' | 'blob' | 'json' | 'text'

interface Parsed {
  type: ContentType
  raw: string
  prettyJson?: string
  minifiedJson?: string
}

function parseValue(value: unknown): Parsed {
  if (value === null || value === undefined) return { type: 'null', raw: 'NULL' }
  if (value === '') return { type: 'empty', raw: '' }
  const raw = String(value)
  if (raw.startsWith('<BLOB ')) return { type: 'blob', raw }
  try {
    const obj = JSON.parse(raw)
    return {
      type: 'json',
      raw,
      prettyJson: JSON.stringify(obj, null, 2),
      minifiedJson: JSON.stringify(obj),
    }
  } catch {
    return { type: 'text', raw }
  }
}

interface CellViewerProps {
  column: string
  value: unknown
  onClose: () => void
}

export default function CellViewer({ column, value, onClose }: CellViewerProps) {
  const { showToast } = useToast()
  const { t } = useI18n()
  const parsed = parseValue(value)
  const [activeTab, setActiveTab] = useState<'raw' | 'pretty'>(
    parsed.type === 'json' ? 'pretty' : 'raw'
  )

  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [onClose])

  function copy(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast(t('copy.cell_success'), 'success'),
      () => showToast('Copy failed'),
    )
  }

  const typeBadge: Record<ContentType, React.ReactNode> = {
    null: (
      <span className="text-xs bg-gray-100 dark:bg-gray-700 text-gray-500 px-1.5 py-0.5 rounded font-mono">
        NULL
      </span>
    ),
    empty: (
      <span className="text-xs bg-gray-100 dark:bg-gray-700 text-gray-500 px-1.5 py-0.5 rounded font-mono">
        {t('cell.empty')}
      </span>
    ),
    blob: (
      <span className="text-xs bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 px-1.5 py-0.5 rounded font-mono">
        {t('cell.blob')}
      </span>
    ),
    json: (
      <span className="text-xs bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 px-1.5 py-0.5 rounded font-mono">
        JSON
      </span>
    ),
    text: null,
  }

  function renderContent() {
    if (parsed.type === 'null') {
      return (
        <div className="flex items-center justify-center h-24 text-gray-300 dark:text-gray-600 italic text-2xl font-mono select-none">
          NULL
        </div>
      )
    }
    if (parsed.type === 'empty') {
      return (
        <div className="flex items-center justify-center h-24 text-gray-400 italic gap-2">
          <span>{t('cell.empty')}</span>
          <span className="font-mono text-gray-500">&quot;&quot;</span>
        </div>
      )
    }
    if (parsed.type === 'blob') {
      return (
        <div className="p-4 space-y-3">
          <div className="font-mono text-sm text-orange-600 dark:text-orange-400">{parsed.raw}</div>
          <div className="text-xs text-gray-400 italic">{t('cell.blob_note')}</div>
        </div>
      )
    }
    const content = parsed.type === 'json' && activeTab === 'pretty'
      ? parsed.prettyJson!
      : parsed.raw
    return (
      <pre className="p-4 text-xs font-mono text-gray-800 dark:text-gray-200 whitespace-pre-wrap break-all leading-relaxed">
        {content}
      </pre>
    )
  }

  const tabCls = (active: boolean) =>
    `px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
      active
        ? 'border-blue-500 text-blue-600 dark:text-blue-400'
        : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
    }`

  const canCopyRaw = parsed.type !== 'null' && parsed.type !== 'empty'

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={e => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-2xl w-full max-w-2xl flex flex-col" style={{ maxHeight: '80vh' }}>
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            <span className="font-mono font-semibold text-gray-800 dark:text-gray-100 truncate">{column}</span>
            {typeBadge[parsed.type]}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 ml-3 shrink-0 text-xl leading-none"
          >✕</button>
        </div>

        {/* Tabs — only for JSON */}
        {parsed.type === 'json' && (
          <div className="flex border-b border-gray-200 dark:border-gray-700 px-4 shrink-0">
            <button className={tabCls(activeTab === 'pretty')} onClick={() => setActiveTab('pretty')}>
              {t('cell.tab.pretty')}
            </button>
            <button className={tabCls(activeTab === 'raw')} onClick={() => setActiveTab('raw')}>
              {t('cell.tab.raw')}
            </button>
          </div>
        )}

        {/* Content */}
        <div className="flex-1 overflow-auto min-h-0">
          {renderContent()}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 px-4 py-3 border-t border-gray-200 dark:border-gray-700 shrink-0">
          {parsed.type === 'json' && (
            <>
              <button
                onClick={() => copy(parsed.minifiedJson!)}
                className="text-sm px-3 py-1.5 border border-gray-200 dark:border-gray-600 rounded text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
              >{t('cell.copy_minified')}</button>
              <button
                onClick={() => copy(parsed.prettyJson!)}
                className="text-sm px-3 py-1.5 border border-gray-200 dark:border-gray-600 rounded text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
              >{t('cell.copy_pretty')}</button>
            </>
          )}
          {canCopyRaw && (
            <button
              onClick={() => copy(parsed.raw)}
              className="text-sm px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700"
            >{t('cell.copy')}</button>
          )}
          {!canCopyRaw && (
            <button
              onClick={onClose}
              className="text-sm px-3 py-1.5 border border-gray-200 dark:border-gray-600 rounded text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
            >{t('confirm.cancel')}</button>
          )}
        </div>
      </div>
    </div>
  )
}
