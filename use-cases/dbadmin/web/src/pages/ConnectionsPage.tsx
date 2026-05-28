import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useToast } from '../components/Toast'
import { useI18n } from '../i18n'

const emptyConn: Partial<Connection> = {
  driver: 'mysql',
  host: 'localhost',
  port: 3306,
  database: '',
  username: 'root',
  password: '',
  name: '',
}

export default function ConnectionsPage() {
  const [conns, setConns] = useState<Connection[]>([])
  const [editing, setEditing] = useState<Partial<Connection> | null>(null)
  const [testing, setTesting] = useState<Record<string, string>>({})
  const [deleteTarget, setDeleteTarget] = useState<Connection | null>(null)
  const [deleteFile, setDeleteFile] = useState(false)
  const navigate = useNavigate()
  const { showToast } = useToast()
  const { t } = useI18n()

  useEffect(() => { reload() }, [])

  function reload() {
    api.listConnections().then(setConns).catch(e => showToast(e.message))
  }

  async function handleSave() {
    if (!editing) return
    try {
      if (editing.id) {
        await api.updateConnection(editing.id, editing)
      } else {
        await api.createConnection(editing)
      }
      setEditing(null)
      reload()
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Save failed')
    }
  }

  function openDeleteDialog(c: Connection) {
    setDeleteTarget(c)
    setDeleteFile(false)
  }

  async function handleDelete() {
    if (!deleteTarget) return
    try {
      await api.deleteConnection(deleteTarget.id, deleteFile)
      setDeleteTarget(null)
      reload()
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  async function handleTest(id: string) {
    setTesting(p => ({ ...p, [id]: 'testing…' }))
    try {
      const r = await api.testConnection(id)
      setTesting(p => ({ ...p, [id]: r.ok ? '✓ Connected' : `✗ ${r.error}` }))
    } catch (e) {
      setTesting(p => ({ ...p, [id]: `✗ ${e instanceof Error ? e.message : 'failed'}` }))
    }
  }

  return (
    <div className="p-6 max-w-4xl">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100">{t('connections.title')}</h1>
        <button
          onClick={() => setEditing({ ...emptyConn })}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {t('connections.add')}
        </button>
      </div>

      <div className="space-y-2">
        {conns.map(c => (
          <div key={c.id} className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 flex items-center gap-4">
            <div className="font-mono text-xs bg-gray-100 dark:bg-gray-700 dark:text-gray-300 px-2 py-1 rounded uppercase">
              {c.driver}
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium text-gray-800 dark:text-gray-100 flex items-center gap-2">
                {c.name}
                {c.readonly && (
                  <span className="text-xs bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 px-1.5 py-0.5 rounded font-mono">
                    {t('readonly.badge')}
                  </span>
                )}
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-400 truncate">
                {c.driver === 'sqlite'
                  ? (c.uploaded_file ? (c.original_filename || c.file_path) : c.file_path)
                  : `${c.host}:${c.port}/${c.database}`}
              </div>
            </div>
            {testing[c.id] && (
              <span className={`text-xs shrink-0 ${testing[c.id].startsWith('✓') ? 'text-green-600' : 'text-red-500'}`}>
                {testing[c.id]}
              </span>
            )}
            <div className="flex gap-2 shrink-0 flex-wrap justify-end">
              <button onClick={() => navigate(`/conn/${c.id}/query`)} className="text-xs text-green-600 hover:text-green-800 px-2 py-1 border border-green-200 rounded">{t('connections.open')}</button>
              {c.driver === 'sqlite' && c.uploaded_file && (
                <a
                  href={api.sqliteDownloadURL(c.id)}
                  className="text-xs text-indigo-600 hover:text-indigo-800 px-2 py-1 border border-indigo-200 rounded"
                >
                  {t('sqlite.download')}
                </a>
              )}
              <button onClick={() => handleTest(c.id)} className="text-xs text-blue-600 hover:text-blue-800 px-2 py-1 border border-blue-200 rounded">{t('connections.test')}</button>
              <button onClick={() => setEditing({ ...c })} className="text-xs text-gray-600 hover:text-gray-800 px-2 py-1 border border-gray-200 dark:border-gray-600 rounded">{t('connections.edit')}</button>
              <button onClick={() => openDeleteDialog(c)} className="text-xs text-red-600 hover:text-red-800 px-2 py-1 border border-red-200 rounded">{t('connections.delete')}</button>
            </div>
          </div>
        ))}
        {conns.length === 0 && (
          <div className="text-center text-gray-400 dark:text-gray-500 py-12">
            {t('connections.empty')}
          </div>
        )}
      </div>

      {editing && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-lg font-bold mb-4 text-gray-900 dark:text-gray-100">
              {editing.id ? t('connections.edit_title') : t('connections.add_title')}
            </h2>
            <ConnectionForm conn={editing} onChange={setEditing} />
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setEditing(null)} className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800">{t('connections.form.cancel')}</button>
              <button onClick={handleSave} className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700">{t('connections.form.save')}</button>
            </div>
          </div>
        </div>
      )}

      {deleteTarget && (
        <DeleteDialog
          conn={deleteTarget}
          deleteFile={deleteFile}
          onDeleteFileChange={setDeleteFile}
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  )
}

function DeleteDialog({
  conn, deleteFile, onDeleteFileChange, onConfirm, onCancel,
}: {
  conn: Connection
  deleteFile: boolean
  onDeleteFileChange: (v: boolean) => void
  onConfirm: () => void
  onCancel: () => void
}) {
  const { t } = useI18n()

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onCancel() }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onCancel])

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onCancel}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-sm mx-4" onClick={e => e.stopPropagation()}>
        <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2">
          {t('connections.delete.title')}
        </h2>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          {t('connections.delete.message', { name: conn.name })}
        </p>
        {conn.uploaded_file && (
          <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 mb-4 cursor-pointer">
            <input
              type="checkbox"
              checked={deleteFile}
              onChange={e => onDeleteFileChange(e.target.checked)}
            />
            {t('sqlite.delete_file')}
          </label>
        )}
        <div className="flex justify-end gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            {t('confirm.cancel')}
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 text-sm text-white rounded font-medium bg-red-600 hover:bg-red-700"
          >
            {t('confirm.delete')}
          </button>
        </div>
      </div>
    </div>
  )
}

function ConnectionForm({ conn, onChange }: { conn: Partial<Connection>; onChange: (c: Partial<Connection>) => void }) {
  const { t } = useI18n()
  const { showToast } = useToast()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadMode, setUploadMode] = useState(() => !!conn.uploaded_file)
  const [uploading, setUploading] = useState(false)

  const field = (key: keyof Connection) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      onChange({ ...conn, [key]: e.target.value })

  const inputCls = 'w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100'

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const result = await api.uploadSQLite(file)
      onChange({ ...conn, file_path: result.file_path, uploaded_file: true, original_filename: result.original_name })
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('sqlite.upload_failed'))
    } finally {
      setUploading(false)
      // Reset input so the same file can be re-selected after an error.
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  function switchMode(toUpload: boolean) {
    setUploadMode(toUpload)
    onChange({ ...conn, file_path: '', uploaded_file: toUpload ? undefined : undefined, original_filename: undefined })
  }

  return (
    <div className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.name')}</label>
        <input value={conn.name || ''} onChange={field('name')} className={inputCls} placeholder="My Database" />
      </div>
      <div>
        <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.driver')}</label>
        <select value={conn.driver || 'mysql'} onChange={field('driver')} className={inputCls}>
          <option value="mysql">MySQL</option>
          <option value="sqlite">SQLite</option>
        </select>
      </div>
      {conn.driver === 'sqlite' ? (
        <>
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => switchMode(false)}
              className={`px-3 py-1.5 rounded border text-xs ${!uploadMode ? 'bg-blue-600 text-white border-blue-600' : 'text-gray-600 dark:text-gray-400 border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'}`}
            >
              {t('sqlite.mode_path')}
            </button>
            <button
              type="button"
              onClick={() => switchMode(true)}
              className={`px-3 py-1.5 rounded border text-xs ${uploadMode ? 'bg-blue-600 text-white border-blue-600' : 'text-gray-600 dark:text-gray-400 border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'}`}
            >
              {t('sqlite.mode_upload')}
            </button>
          </div>
          {uploadMode ? (
            <div>
              <input
                ref={fileInputRef}
                type="file"
                accept=".db,.sqlite,.sqlite3"
                className="hidden"
                onChange={handleFileChange}
              />
              {conn.file_path ? (
                <div className="flex items-center gap-2 border border-gray-200 dark:border-gray-600 rounded px-3 py-2">
                  <span className="flex-1 text-xs text-gray-600 dark:text-gray-400 truncate font-mono">
                    {conn.original_filename || conn.file_path.split('/').pop()}
                  </span>
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={uploading}
                    className="text-xs text-blue-600 hover:text-blue-800 shrink-0 disabled:opacity-50"
                  >
                    {uploading ? t('sqlite.uploading') : t('sqlite.replace')}
                  </button>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploading}
                  className="w-full border-2 border-dashed border-gray-300 dark:border-gray-600 rounded px-4 py-6 text-sm text-gray-500 dark:text-gray-400 hover:border-blue-400 hover:text-blue-600 disabled:opacity-50 text-center"
                >
                  {uploading ? t('sqlite.uploading') : t('sqlite.upload_choose')}
                </button>
              )}
              <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{t('sqlite.upload_hint')}</p>
            </div>
          ) : (
            <div>
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.file_path')}</label>
              <input value={conn.file_path || ''} onChange={field('file_path')} className={inputCls} placeholder="/path/to/database.db" />
            </div>
          )}
        </>
      ) : (
        <>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} placeholder="localhost" />
            </div>
            <div className="w-24">
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.port')}</label>
              <input value={conn.port || 3306} onChange={field('port')} type="number" className={inputCls} />
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.database')}</label>
            <input value={conn.database || ''} onChange={field('database')} className={inputCls} />
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.username')}</label>
              <input value={conn.username || ''} onChange={field('username')} className={inputCls} />
            </div>
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('connections.form.password')}</label>
              <input value={conn.password || ''} onChange={field('password')} type="password" className={inputCls} disabled={!conn.save_password} placeholder={conn.save_password ? '' : '••••••'} />
            </div>
          </div>
          <div>
            <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700 dark:text-gray-300">
              <input
                type="checkbox"
                checked={!!conn.save_password}
                onChange={e => onChange({ ...conn, save_password: e.target.checked, password: e.target.checked ? (conn.password ?? '') : '' })}
              />
              {t('connections.form.save_password')}
            </label>
            {conn.save_password && (
              <div className="mt-1 px-3 py-2 bg-amber-50 dark:bg-amber-900/30 border border-amber-300 dark:border-amber-600 text-amber-700 dark:text-amber-400 text-xs rounded">
                {t('connections.form.save_password_warn')}
              </div>
            )}
          </div>
        </>
      )}
      <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700 dark:text-gray-300">
        <input
          type="checkbox"
          checked={!!conn.readonly}
          onChange={e => onChange({ ...conn, readonly: e.target.checked })}
        />
        {t('connections.form.readonly')}
        <span className="text-xs text-gray-400">— {t('connections.form.readonly_hint')}</span>
      </label>
    </div>
  )
}
