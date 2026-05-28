import { useEffect, useState } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { api, type TableStructure } from '../api'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'
import { toMarkdownSchema, toJSONSchemaDesc } from '../utils/copyFormats'

type TabKey = 'columns' | 'indexes' | 'fks' | 'ddl'
const VALID_TABS: TabKey[] = ['columns', 'indexes', 'fks', 'ddl']

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

  useEffect(() => {
    if (!connId || !dbName || !tableName) return
    api.tableStructure(connId, dbName, tableName)
      .then(setStructure)
      .catch(e => showToast(e.message))
  }, [connId, dbName, tableName])

  function copyText(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast(t('copy.cell_success'), 'success'),
      () => showToast('Copy failed'),
    )
  }

  if (!structure) return <div className="p-6 text-gray-400">Loading…</div>

  const tabLabel: Record<TabKey, string> = {
    columns: t('structure.tab.columns'),
    indexes: t('structure.tab.indexes'),
    fks: t('structure.tab.fks'),
    ddl: t('structure.tab.ddl'),
  }

  const copyBtnClass = 'text-xs px-2.5 py-1 rounded border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-5 pb-0 border-b border-gray-100 dark:border-gray-700 shrink-0">
        <div className="flex items-start justify-between mb-3">
          <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100">
            Structure — <span className="font-mono">{tableName}</span>
          </h1>
          {/* Schema copy buttons */}
          {tab === 'columns' && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-gray-400">{t('copy.as')}</span>
              <button
                onClick={() => copyText(toMarkdownSchema(tableName!, structure.columns))}
                className={copyBtnClass}
              >{t('copy.markdown')}</button>
              <button
                onClick={() => copyText(toJSONSchemaDesc(tableName!, structure.columns))}
                className={copyBtnClass}
              >{t('copy.json_schema')}</button>
              {structure.ddl && (
                <button
                  onClick={() => copyText(structure.ddl!)}
                  className={copyBtnClass}
                >{t('copy.ddl')}</button>
              )}
            </div>
          )}
          {tab === 'ddl' && structure.ddl && (
            <button
              onClick={() => copyText(structure.ddl!)}
              className={copyBtnClass}
            >{t('copy.ddl')}</button>
          )}
        </div>
        <div className="flex gap-1">
          {VALID_TABS.map(tk => (
            <button
              key={tk}
              onClick={() => setTab(tk)}
              className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                tab === tk
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
              }`}
            >
              {tabLabel[tk]}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        {tab === 'columns' && (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10 border-b border-gray-200 dark:border-gray-700">
              <tr>
                {['#', 'Name', 'Type', 'Nullable', 'Default', 'Key', 'Extra'].map(h => (
                  <th key={h} className="text-left px-4 py-2 font-medium text-gray-600 dark:text-gray-400">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {(structure.columns ?? []).map(c => (
                <tr key={c.name} className={c.primary_key ? 'bg-yellow-50 dark:bg-yellow-900/20' : 'hover:bg-gray-50 dark:hover:bg-gray-800'}>
                  <td className="px-4 py-2 text-gray-400 tabular-nums">{c.position}</td>
                  <td className="px-4 py-2 font-mono font-medium text-gray-800 dark:text-gray-200">{c.name}</td>
                  <td className="px-4 py-2 text-blue-700 dark:text-blue-400 font-mono text-xs">{c.full_type}</td>
                  <td className="px-4 py-2 text-center text-gray-500">{c.nullable ? '✓' : ''}</td>
                  <td className="px-4 py-2 text-gray-400 text-xs font-mono">{c.default ?? ''}</td>
                  <td className="px-4 py-2 text-xs">{c.primary_key ? <span className="text-amber-600 font-medium">PK</span> : ''}</td>
                  <td className="px-4 py-2 text-gray-400 text-xs">{c.auto_increment ? 'AUTO_INCREMENT' : ''}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {tab === 'indexes' && (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10 border-b border-gray-200 dark:border-gray-700">
              <tr>
                {['Name', 'Unique', 'Columns', 'Type'].map(h => (
                  <th key={h} className="text-left px-4 py-2 font-medium text-gray-600 dark:text-gray-400">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {(structure.indexes ?? []).map(idx => (
                <tr key={idx.name} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-2 font-mono text-sm text-gray-800 dark:text-gray-200">{idx.name}</td>
                  <td className="px-4 py-2 text-center">
                    {idx.unique ? <span className="text-green-600 font-medium text-xs">UNIQUE</span> : ''}
                  </td>
                  <td className="px-4 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{(idx.columns ?? []).join(', ')}</td>
                  <td className="px-4 py-2 text-gray-400 text-xs">{idx.type || ''}</td>
                </tr>
              ))}
              {(structure.indexes ?? []).length === 0 && (
                <tr><td colSpan={4} className="px-4 py-10 text-center text-gray-400">{t('structure.no_indexes')}</td></tr>
              )}
            </tbody>
          </table>
        )}

        {tab === 'fks' && (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10 border-b border-gray-200 dark:border-gray-700">
              <tr>
                {['Name', 'Column', 'References', 'On Delete', 'On Update'].map(h => (
                  <th key={h} className="text-left px-4 py-2 font-medium text-gray-600 dark:text-gray-400">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {(structure.foreign_keys ?? []).map(fk => (
                <tr key={fk.name} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-4 py-2 font-mono text-xs text-gray-800 dark:text-gray-200">{fk.name}</td>
                  <td className="px-4 py-2 font-mono text-gray-800 dark:text-gray-200">{fk.column}</td>
                  <td className="px-4 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{fk.ref_table}.{fk.ref_column}</td>
                  <td className="px-4 py-2 text-gray-400 text-xs">{fk.on_delete}</td>
                  <td className="px-4 py-2 text-gray-400 text-xs">{fk.on_update}</td>
                </tr>
              ))}
              {(structure.foreign_keys ?? []).length === 0 && (
                <tr><td colSpan={5} className="px-4 py-10 text-center text-gray-400">{t('structure.no_fks')}</td></tr>
              )}
            </tbody>
          </table>
        )}

        {tab === 'ddl' && (
          <pre className="m-4 bg-gray-900 text-green-400 p-4 rounded-lg text-xs overflow-auto font-mono whitespace-pre-wrap">
            {structure.ddl || '-- DDL not available'}
          </pre>
        )}
      </div>
    </div>
  )
}
