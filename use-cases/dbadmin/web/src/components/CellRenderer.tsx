interface CellRendererProps {
  value: unknown
  colType?: string // SQL column type hint, e.g. "int", "varchar(255)"
}

function isNumericType(colType?: string): boolean {
  if (!colType) return false
  return /int|float|double|decimal|numeric|real|bigint|smallint|tinyint/i.test(colType)
}

function isJsonLike(s: string): boolean {
  const t = s.trim()
  return (t.startsWith('{') && t.endsWith('}')) || (t.startsWith('[') && t.endsWith(']'))
}

export default function CellRenderer({ value, colType }: CellRendererProps) {
  if (value === null || value === undefined) {
    return (
      <span
        className="font-mono italic select-none text-[11px]"
        style={{ color: 'var(--text-subtle)' }}
      >
        NULL
      </span>
    )
  }

  if (value === '') {
    return (
      <span
        className="italic select-none text-[11px]"
        style={{ color: 'var(--text-subtle)' }}
      >
        (empty)
      </span>
    )
  }

  const s = String(value)

  if (s.startsWith('<BLOB ')) {
    const meta = s.slice(6, s.endsWith('>') ? s.length - 1 : s.length)
    return (
      <span className="inline-flex items-center gap-1.5 min-w-0">
        <span className="shrink-0 text-[10px] bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 px-1 py-px rounded font-mono leading-none">
          BLOB
        </span>
        <span
          className="font-mono text-[11px] truncate"
          style={{ color: 'var(--text-muted)' }}
        >
          {meta}
        </span>
      </span>
    )
  }

  if (isJsonLike(s)) {
    const display = s.length > 120 ? s.slice(0, 120) + '…' : s
    return (
      <span className="inline-flex items-center gap-1.5 min-w-0">
        <span className="shrink-0 text-[10px] bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 px-1 py-px rounded font-mono leading-none">
          JSON
        </span>
        <span
          className="font-mono text-[11px] truncate"
          style={{ color: 'var(--text-muted)' }}
        >
          {display}
        </span>
      </span>
    )
  }

  const numeric = isNumericType(colType) || typeof value === 'number'

  if (s.length > 120) {
    return (
      <span className={numeric ? 'font-mono tabular-nums text-right block' : ''}>
        {s.slice(0, 120)}
        <span style={{ color: 'var(--text-subtle)' }}>…</span>
      </span>
    )
  }

  if (numeric) {
    return (
      <span className="font-mono tabular-nums block text-right">{s}</span>
    )
  }

  return <>{s}</>
}
