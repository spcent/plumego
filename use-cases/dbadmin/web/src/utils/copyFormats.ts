import type { ColumnInfo } from '../api'

function sqlEscapeString(s: string): string {
  return s
    .replace(/\\/g, '\\\\')
    .replace(/'/g, "\\'")
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')
    .replace(/\0/g, '\\0')
}

function sqlValue(v: unknown): string {
  if (v === null || v === undefined) return 'NULL'
  if (typeof v === 'boolean') return v ? '1' : '0'
  if (typeof v === 'number') return String(v)
  return `'${sqlEscapeString(String(v))}'`
}

function sqlIdent(name: string): string {
  return '`' + name.replace(/`/g, '``') + '`'
}

export function toInsertSQL(
  tableName: string,
  columns: string[],
  rows: Record<string, unknown>[],
): string {
  if (rows.length === 0) return ''
  const cols = columns.map(sqlIdent).join(', ')
  const valueClauses = rows.map(row => {
    const vals = columns.map(c => sqlValue(row[c])).join(', ')
    return `  (${vals})`
  })
  return `INSERT INTO ${sqlIdent(tableName)} (${cols}) VALUES\n${valueClauses.join(',\n')};`
}

function csvCell(v: unknown): string {
  if (v === null || v === undefined) return 'NULL'
  const s = String(v)
  if (s.includes(',') || s.includes('"') || s.includes('\n') || s.includes('\r')) {
    return '"' + s.replace(/"/g, '""') + '"'
  }
  return s
}

export function toCSV(columns: string[], rows: Record<string, unknown>[]): string {
  const header = columns.map(c => csvCell(c)).join(',')
  const body = rows.map(row => columns.map(c => csvCell(row[c])).join(','))
  return [header, ...body].join('\r\n')
}

export function toRowJSON(columns: string[], rows: Record<string, unknown>[]): string {
  const objs = rows.map(row => {
    const obj: Record<string, unknown> = {}
    for (const c of columns) {
      obj[c] = row[c] !== undefined ? row[c] : null
    }
    return obj
  })
  return JSON.stringify(objs.length === 1 ? objs[0] : objs, null, 2)
}

export function toMarkdownSchema(tableName: string, columns: ColumnInfo[]): string {
  const lines = [
    `## \`${tableName}\``,
    '',
    '| # | Column | Type | Nullable | Default | Key |',
    '|---|--------|------|----------|---------|-----|',
    ...columns.map(c =>
      `| ${c.position} | \`${c.name}\` | \`${c.full_type}\` | ${c.nullable ? 'YES' : 'NO'} | ${c.default ?? ''} | ${c.primary_key ? 'PK' : ''} |`
    ),
  ]
  return lines.join('\n')
}

export function toJSONSchemaDesc(tableName: string, columns: ColumnInfo[]): string {
  return JSON.stringify(
    {
      table: tableName,
      columns: columns.map(c => ({
        name: c.name,
        type: c.full_type,
        nullable: c.nullable,
        ...(c.primary_key ? { primary_key: true } : {}),
        ...(c.default !== undefined && c.default !== null ? { default: c.default } : {}),
        ...(c.auto_increment ? { auto_increment: true } : {}),
      })),
    },
    null,
    2,
  )
}
