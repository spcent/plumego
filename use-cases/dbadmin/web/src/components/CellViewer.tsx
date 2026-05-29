import { useEffect, useState } from 'react'
import { useToast } from './Toast'
import { useI18n } from '../i18n'

type ContentType = 'null' | 'empty' | 'blob' | 'json' | 'text'

interface Parsed {
  type: ContentType
  raw: string
  prettyJson?: string
  minifiedJson?: string
  blobSize?: number
  hexPreview?: string
}

function parseValue(value: unknown): Parsed {
  if (value === null || value === undefined) return { type: 'null', raw: 'NULL' }
  if (value === '') return { type: 'empty', raw: '' }
  const raw = String(value)
  if (raw.startsWith('<BLOB ')) {
    const pipeIdx = raw.indexOf('|')
    const sizeMatch = raw.match(/^<BLOB (\d+) bytes/)
    const blobSize = sizeMatch ? parseInt(sizeMatch[1], 10) : 0
    const hexPreview = pipeIdx !== -1 ? raw.slice(pipeIdx + 1, -1) : ''
    return { type: 'blob', raw, blobSize, hexPreview }
  }
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

function formatHexDump(hex: string): string {
  const bytes: string[] = []
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(hex.slice(i, i + 2))
  }
  const lines: string[] = []
  for (let i = 0; i < bytes.length; i += 16) {
    const chunk = bytes.slice(i, i + 16)
    const hexPart = chunk.map((b, j) => (j === 8 ? ' ' : '') + b).join(' ')
    const asciiPart = chunk.map(b => {
      const code = parseInt(b, 16)
      return code >= 32 && code < 127 ? String.fromCharCode(code) : '.'
    }).join('')
    const offset = i.toString(16).padStart(4, '0')
    lines.push(`${offset}  ${hexPart.padEnd(49)}  ${asciiPart}`)
  }
  return lines.join('\n')
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
      <span className="text-xs px-1.5 py-0.5 rounded font-mono"
        style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
        NULL
      </span>
    ),
    empty: (
      <span className="text-xs px-1.5 py-0.5 rounded font-mono"
        style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
        {t('cell.empty')}
      </span>
    ),
    blob: (
      <span className="text-xs px-1.5 py-0.5 rounded font-mono"
        style={{ background: 'var(--warning)22', color: 'var(--warning)' }}>
        {t('cell.blob')}
      </span>
    ),
    json: (
      <span className="text-xs px-1.5 py-0.5 rounded font-mono"
        style={{ background: 'var(--accent)22', color: 'var(--accent)' }}>
        JSON
      </span>
    ),
    text: null,
  }

  function renderContent() {
    if (parsed.type === 'null') {
      return (
        <div className="flex items-center justify-center h-24 italic text-2xl font-mono select-none"
          style={{ color: 'var(--text-subtle)' }}>
          NULL
        </div>
      )
    }
    if (parsed.type === 'empty') {
      return (
        <div className="flex items-center justify-center h-24 italic gap-2"
          style={{ color: 'var(--text-muted)' }}>
          <span>{t('cell.empty')}</span>
          <span className="font-mono">&quot;&quot;</span>
        </div>
      )
    }
    if (parsed.type === 'blob') {
      const hexDump = parsed.hexPreview ? formatHexDump(parsed.hexPreview) : ''
      const previewBytes = parsed.hexPreview ? parsed.hexPreview.length / 2 : 0
      return (
        <div className="p-4 space-y-3">
          <div className="flex items-center gap-3">
            <span className="font-mono text-sm" style={{ color: 'var(--warning)' }}>
              BLOB — {parsed.blobSize?.toLocaleString()} bytes
            </span>
            {parsed.blobSize !== undefined && previewBytes < parsed.blobSize && (
              <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
                (showing first {previewBytes} bytes)
              </span>
            )}
          </div>
          {hexDump && (
            <pre
              className="rounded p-3 text-xs font-mono leading-relaxed overflow-auto"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-default)', border: '1px solid var(--border-subtle)' }}
            >
              {hexDump}
            </pre>
          )}
          {!hexDump && (
            <div className="text-xs italic" style={{ color: 'var(--text-subtle)' }}>{t('cell.blob_note')}</div>
          )}
        </div>
      )
    }
    const content = parsed.type === 'json' && activeTab === 'pretty'
      ? parsed.prettyJson!
      : parsed.raw
    return (
      <pre className="p-4 text-xs font-mono whitespace-pre-wrap break-all leading-relaxed"
        style={{ color: 'var(--text-default)' }}>
        {content}
      </pre>
    )
  }

  const canCopyRaw = parsed.type !== 'null' && parsed.type !== 'empty'

  const tabStyle = (active: boolean): React.CSSProperties => ({
    padding: '8px 16px',
    fontSize: 13,
    fontWeight: active ? 500 : 400,
    borderBottom: active ? '2px solid var(--accent)' : '2px solid transparent',
    color: active ? 'var(--accent)' : 'var(--text-muted)',
    background: 'transparent',
    marginBottom: -1,
  })

  const btnStyle: React.CSSProperties = {
    border: '1px solid var(--border-strong)',
    color: 'var(--text-default)',
    background: 'var(--bg-muted)',
  }

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={e => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="rounded-lg shadow-2xl w-full max-w-2xl flex flex-col"
        style={{ maxHeight: '80vh', background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}>
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 shrink-0"
          style={{ borderBottom: '1px solid var(--border-subtle)' }}>
          <div className="flex items-center gap-2 min-w-0">
            <span className="font-mono font-semibold truncate" style={{ color: 'var(--text-strong)' }}>{column}</span>
            {typeBadge[parsed.type]}
          </div>
          <button
            onClick={onClose}
            className="ml-3 shrink-0 text-xl leading-none"
            style={{ color: 'var(--text-subtle)' }}
          >✕</button>
        </div>

        {/* Tabs — only for JSON */}
        {parsed.type === 'json' && (
          <div className="flex px-4 shrink-0" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
            <button style={tabStyle(activeTab === 'pretty')} onClick={() => setActiveTab('pretty')}>
              {t('cell.tab.pretty')}
            </button>
            <button style={tabStyle(activeTab === 'raw')} onClick={() => setActiveTab('raw')}>
              {t('cell.tab.raw')}
            </button>
          </div>
        )}

        {/* Content */}
        <div className="flex-1 overflow-auto min-h-0">
          {renderContent()}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 px-4 py-3 shrink-0"
          style={{ borderTop: '1px solid var(--border-subtle)' }}>
          {parsed.type === 'json' && (
            <>
              <button
                onClick={() => copy(parsed.minifiedJson!)}
                className="text-sm px-3 py-1.5 rounded"
                style={btnStyle}
              >{t('cell.copy_minified')}</button>
              <button
                onClick={() => copy(parsed.prettyJson!)}
                className="text-sm px-3 py-1.5 rounded"
                style={btnStyle}
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
              className="text-sm px-3 py-1.5 rounded"
              style={btnStyle}
            >{t('confirm.cancel')}</button>
          )}
        </div>
      </div>
    </div>
  )
}
