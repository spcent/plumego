import { useEffect, useState, type CSSProperties } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { api, type TableStructure } from '../api'
import WorkbenchHeader from '../components/WorkbenchHeader'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { useCurrentConn } from '../context/connections'
import { toMarkdownSchema, toJSONSchemaDesc } from '../utils/copyFormats'
import { CodePanel, PageBody, PageShell, PageToolbar } from '../components/workbench'

type TabKey = 'columns' | 'indexes' | 'fks' | 'ddl'
const VALID_TABS: TabKey[] = ['columns', 'indexes', 'fks', 'ddl']

const th: CSSProperties = {
  textAlign: 'left',
  padding: '6px 16px',
  fontWeight: 500,
  fontSize: 12,
  color: 'var(--text-muted)',
  background: 'var(--bg-muted)',
  borderBottom: '1px solid var(--border-subtle)',
  position: 'sticky',
  top: 0,
  zIndex: 10,
}

const td: CSSProperties = {
  padding: '6px 16px',
  fontSize: 13,
  borderBottom: '1px solid var(--border-subtle)',
  color: 'var(--text-default)',
}

export default function StructurePage() {
  const { connId, dbName, tableName } = useParams<{ connId: string; dbName: string; tableName: string }>()
  const [searchParams] = useSearchParams()
  const initialTab = VALID_TABS.includes(searchParams.get('tab') as TabKey)
    ? (searchParams.get('tab') as TabKey)
    : 'columns'
  const [structure, setStructure] = useState<TableStructure | null>(null)
  const [tab, setTab] = useState<TabKey>(initialTab)
  const { showToast } = useToast()
  const { t } = useI18n()
  const currentConn = useCurrentConn(connId)

  useEffect(() => {
    if (!connId || !dbName || !tableName) return
    api.tableStructure(connId, dbName, tableName)
      .then(setStructure)
      .catch(e => showToast(e.message))
  }, [connId, dbName, tableName, showToast])

  function copyText(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast(t('copy.cell_success'), 'success'),
      () => showToast('Copy failed'),
    )
  }

  if (!structure) {
    return (
      <PageShell>
        <PageBody scroll>
          <div className="p-6 text-sm" style={{ color: 'var(--text-muted)' }}>
            Loading…
          </div>
        </PageBody>
      </PageShell>
    )
  }

  const tabLabel: Record<TabKey, string> = {
    columns: t('structure.tab.columns'),
    indexes: t('structure.tab.indexes'),
    fks: t('structure.tab.fks'),
    ddl: t('structure.tab.ddl'),
  }

  return (
    <PageShell>
      <WorkbenchHeader
        connectionName={currentConn?.name}
        resourcePath={dbName && tableName ? [dbName, tableName] : []}
        datasourceType={currentConn?.driver ?? 'mysql'}
      />
      <PageToolbar
        leading={
          <div className="flex min-w-0 flex-wrap items-center gap-1">
            {VALID_TABS.map(tk => (
              <button
                key={tk}
                onClick={() => setTab(tk)}
                className="tab-btn"
                data-active={tab === tk}
              >
                {tabLabel[tk]}
              </button>
            ))}
          </div>
        }
        trailing={
          <>
          {tab === 'columns' && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>{t('copy.as')}</span>
              <button className="btn btn-ghost h-7 px-2 text-xs" onClick={() => copyText(toMarkdownSchema(tableName!, structure.columns))}>
                {t('copy.markdown')}
              </button>
              <button className="btn btn-ghost h-7 px-2 text-xs" onClick={() => copyText(toJSONSchemaDesc(tableName!, structure.columns))}>
                {t('copy.json_schema')}
              </button>
              {structure.ddl && (
                <button className="btn btn-ghost h-7 px-2 text-xs" onClick={() => copyText(structure.ddl!)}>
                  {t('copy.ddl')}
                </button>
              )}
            </div>
          )}
          {tab === 'ddl' && structure.ddl && (
            <button className="btn btn-ghost h-7 px-2 text-xs" onClick={() => copyText(structure.ddl!)}>
              {t('copy.ddl')}
            </button>
          )}
          </>
        }
      />

      <PageBody>
        <div className="data-table-wrap h-full">
        {tab === 'columns' && (
          <table className="data-table" style={{ fontSize: 13 }}>
            <thead>
              <tr>
                {['#', 'Name', 'Type', 'Nullable', 'Default', 'Key', 'Extra'].map(h => (
                  <th key={h} style={th}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(structure.columns ?? []).map(c => (
                <tr
                  key={c.name}
                  style={{ background: c.primary_key ? 'var(--bg-selected)' : 'transparent' }}
                  className="hover:bg-[var(--bg-hover)]"
                >
                  <td style={{ ...td, color: 'var(--text-subtle)', fontVariantNumeric: 'tabular-nums' }}>{c.position}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontWeight: 500, color: 'var(--text-strong)' }}>{c.name}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 11, color: 'var(--accent)' }}>{c.full_type}</td>
                  <td style={{ ...td, textAlign: 'center', color: 'var(--text-muted)' }}>
                    {c.nullable ? <span className="status-dot mx-auto" style={{ background: 'var(--success)' }} /> : ''}
                  </td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 11, color: 'var(--text-subtle)' }}>{c.default ?? ''}</td>
                  <td style={{ ...td, fontSize: 11 }}>
                    {c.primary_key ? <span style={{ color: 'var(--warning)', fontWeight: 600 }}>PK</span> : ''}
                  </td>
                  <td style={{ ...td, fontSize: 11, color: 'var(--text-subtle)' }}>{c.auto_increment ? 'AUTO_INCREMENT' : ''}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {tab === 'indexes' && (
          <table className="data-table" style={{ fontSize: 13 }}>
            <thead>
              <tr>
                {['Name', 'Unique', 'Columns', 'Type'].map(h => (
                  <th key={h} style={th}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(structure.indexes ?? []).map(idx => (
                <tr key={idx.name} className="hover:bg-[var(--bg-hover)]">
                  <td style={{ ...td, fontFamily: 'monospace' }}>{idx.name}</td>
                  <td style={{ ...td, textAlign: 'center' }}>
                    {idx.unique ? <span style={{ color: 'var(--success)', fontWeight: 600, fontSize: 11 }}>UNIQUE</span> : ''}
                  </td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 11 }}>{(idx.columns ?? []).join(', ')}</td>
                  <td style={{ ...td, fontSize: 11, color: 'var(--text-subtle)' }}>{idx.type || ''}</td>
                </tr>
              ))}
              {(structure.indexes ?? []).length === 0 && (
                <tr>
                  <td colSpan={4} style={{ padding: '40px 16px', textAlign: 'center', color: 'var(--text-subtle)' }}>
                    {t('structure.no_indexes')}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}

        {tab === 'fks' && (
          <table className="data-table" style={{ fontSize: 13 }}>
            <thead>
              <tr>
                {['Name', 'Column', 'References', 'On Delete', 'On Update'].map(h => (
                  <th key={h} style={th}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(structure.foreign_keys ?? []).map(fk => (
                <tr key={fk.name} className="hover:bg-[var(--bg-hover)]">
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 11 }}>{fk.name}</td>
                  <td style={{ ...td, fontFamily: 'monospace' }}>{fk.column}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 11 }}>{fk.ref_table}.{fk.ref_column}</td>
                  <td style={{ ...td, fontSize: 11, color: 'var(--text-subtle)' }}>{fk.on_delete}</td>
                  <td style={{ ...td, fontSize: 11, color: 'var(--text-subtle)' }}>{fk.on_update}</td>
                </tr>
              ))}
              {(structure.foreign_keys ?? []).length === 0 && (
                <tr>
                  <td colSpan={5} style={{ padding: '40px 16px', textAlign: 'center', color: 'var(--text-subtle)' }}>
                    {t('structure.no_fks')}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}

        {tab === 'ddl' && (
          <div className="h-full p-4">
            <CodePanel height="100%">
              <pre
                className="h-full overflow-auto p-4 text-xs"
                style={{
                  fontFamily: 'monospace',
                  color: '#56d364',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {structure.ddl || '-- DDL not available'}
              </pre>
            </CodePanel>
          </div>
        )}
        </div>
      </PageBody>
    </PageShell>
  )
}
