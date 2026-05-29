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

type DriverFilter = 'all' | 'mysql' | 'sqlite' | 'redis'

export default function ConnectionsPage() {
  const [conns, setConns] = useState<Connection[]>([])
  const [driverFilter, setDriverFilter] = useState<DriverFilter>('all')
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

  async function handleDuplicate(c: Connection) {
    const { id: _id, name, ...rest } = c
    try {
      await api.createConnection({ ...rest, name: `${name} (copy)` })
      reload()
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Duplicate failed')
    }
  }

  const filteredConns = driverFilter === 'all' ? conns : conns.filter(c => c.driver === driverFilter)

  return (
    <div className="p-6 max-w-4xl">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold" style={{ color: 'var(--text-strong)' }}>{t('connections.title')}</h1>
        <button
          onClick={() => setEditing({ ...emptyConn })}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
        >
          {t('connections.add')}
        </button>
      </div>

      {conns.length > 0 && (
        <div className="flex gap-1 mb-3">
          {(['all', 'mysql', 'sqlite', 'redis'] as DriverFilter[]).map(f => (
            <button
              key={f}
              onClick={() => setDriverFilter(f)}
              className="px-3 py-1 rounded text-xs border"
              style={driverFilter === f
                ? { background: 'var(--accent)', color: '#fff', borderColor: 'var(--accent)' }
                : { background: 'var(--bg-muted)', color: 'var(--text-muted)', borderColor: 'var(--border-strong)' }
              }
            >
              {t(`connections.filter.${f}`)}
            </button>
          ))}
        </div>
      )}

      <div className="space-y-2">
        {filteredConns.map(c => (
          <div
            key={c.id}
            className="rounded-lg p-4 flex items-center gap-4 border"
            style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)' }}
          >
            <div
              className="font-mono text-xs px-2 py-1 rounded uppercase"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-default)' }}
            >
              {c.driver}
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium flex items-center gap-2" style={{ color: 'var(--text-strong)' }}>
                {c.name}
                {c.readonly && (
                  <span
                    className="text-xs px-1.5 py-0.5 rounded font-mono"
                    style={{ background: 'var(--warning)22', color: 'var(--warning)' }}
                  >
                    {t('readonly.badge')}
                  </span>
                )}
              </div>
              <div className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>
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
              <button
                onClick={() => setEditing({ ...c })}
                className="text-xs px-2 py-1 border rounded"
                style={{ color: 'var(--text-default)', borderColor: 'var(--border-strong)' }}
              >
                {t('connections.edit')}
              </button>
              <button onClick={() => handleDuplicate(c)} className="text-xs px-2 py-1 border rounded" style={{ color: 'var(--text-muted)', borderColor: 'var(--border-strong)' }}>{t('connections.duplicate')}</button>
              <button onClick={() => openDeleteDialog(c)} className="text-xs text-red-600 hover:text-red-800 px-2 py-1 border border-red-200 rounded">{t('connections.delete')}</button>
            </div>
          </div>
        ))}
        {filteredConns.length === 0 && conns.length > 0 && (
          <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
            No {driverFilter} connections.
          </div>
        )}
        {conns.length === 0 && (
          <div className="text-center py-12" style={{ color: 'var(--text-muted)' }}>
            {t('connections.empty')}
          </div>
        )}
      </div>

      {editing && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div
            className="rounded-lg shadow-xl p-6 w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto"
            style={{ background: 'var(--bg-surface)' }}
          >
            <h2 className="text-lg font-bold mb-4" style={{ color: 'var(--text-strong)' }}>
              {editing.id ? t('connections.edit_title') : t('connections.add_title')}
            </h2>
            <ConnectionForm conn={editing} onChange={setEditing} />
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setEditing(null)}
                className="px-4 py-2 text-sm"
                style={{ color: 'var(--text-muted)' }}
              >
                {t('connections.form.cancel')}
              </button>
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
      <div
        className="rounded-lg shadow-xl p-6 w-full max-w-sm mx-4"
        style={{ background: 'var(--bg-surface)' }}
        onClick={e => e.stopPropagation()}
      >
        <h2 className="text-base font-semibold mb-2" style={{ color: 'var(--text-strong)' }}>
          {t('connections.delete.title')}
        </h2>
        <p className="text-sm mb-4" style={{ color: 'var(--text-default)' }}>
          {t('connections.delete.message', { name: conn.name })}
        </p>
        {conn.uploaded_file && (
          <label className="flex items-center gap-2 text-sm mb-4 cursor-pointer" style={{ color: 'var(--text-default)' }}>
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
            className="px-4 py-2 text-sm border rounded"
            style={{ color: 'var(--text-default)', borderColor: 'var(--border-strong)' }}
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
  const [uploadProgress, setUploadProgress] = useState(0)

  const field = (key: keyof Connection) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      onChange({ ...conn, [key]: e.target.value })

  const inputCls = 'w-full border rounded px-2 py-1.5 text-sm'
  const inputStyle = {
    borderColor: 'var(--border-strong)',
    background: 'var(--bg-muted)',
    color: 'var(--text-strong)',
  }

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    setUploadProgress(0)
    try {
      const result = await api.uploadSQLite(file, setUploadProgress)
      onChange({ ...conn, file_path: result.file_path, uploaded_file: true, original_filename: result.original_name })
    } catch (err) {
      showToast(err instanceof Error ? err.message : t('sqlite.upload_failed'))
    } finally {
      setUploading(false)
      setUploadProgress(0)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  function switchMode(toUpload: boolean) {
    setUploadMode(toUpload)
    onChange({ ...conn, file_path: '', uploaded_file: toUpload ? undefined : undefined, original_filename: undefined })
  }

  const labelCls = 'block text-xs font-medium mb-1'
  const labelStyle = { color: 'var(--text-default)' }

  return (
    <div className="space-y-3">
      <div>
        <label className={labelCls} style={labelStyle}>{t('connections.form.name')}</label>
        <input value={conn.name || ''} onChange={field('name')} className={inputCls} style={inputStyle} placeholder="My Database" />
      </div>
      <div>
        <label className={labelCls} style={labelStyle}>{t('connections.form.driver')}</label>
        <select value={conn.driver || 'mysql'} onChange={field('driver')} className={inputCls} style={inputStyle}>
          <option value="mysql">MySQL</option>
          <option value="sqlite">SQLite</option>
          <option value="redis">Redis</option>
        </select>
      </div>
      {conn.driver === 'sqlite' ? (
        <>
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => switchMode(false)}
              className={`px-3 py-1.5 rounded border text-xs ${!uploadMode ? 'bg-blue-600 text-white border-blue-600' : ''}`}
              style={uploadMode ? { color: 'var(--text-default)', borderColor: 'var(--border-strong)' } : {}}
            >
              {t('sqlite.mode_path')}
            </button>
            <button
              type="button"
              onClick={() => switchMode(true)}
              className={`px-3 py-1.5 rounded border text-xs ${uploadMode ? 'bg-blue-600 text-white border-blue-600' : ''}`}
              style={!uploadMode ? { color: 'var(--text-default)', borderColor: 'var(--border-strong)' } : {}}
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
                <div className="flex items-center gap-2 border rounded px-3 py-2" style={{ borderColor: 'var(--border-subtle)' }}>
                  <span className="flex-1 text-xs truncate font-mono" style={{ color: 'var(--text-muted)' }}>
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
                  className="w-full border-2 border-dashed rounded px-4 py-6 text-sm hover:border-blue-400 hover:text-blue-600 disabled:opacity-50 text-center"
                  style={{ borderColor: 'var(--border-strong)', color: 'var(--text-muted)' }}
                >
                  {uploading ? t('sqlite.uploading') : t('sqlite.upload_choose')}
                </button>
              )}
              {uploading && (
                <div className="mt-2 h-1.5 rounded-full overflow-hidden" style={{ background: 'var(--border-subtle)' }}>
                  <div
                    className="h-full rounded-full transition-all duration-150"
                    style={{ width: `${uploadProgress}%`, background: 'var(--accent)' }}
                  />
                </div>
              )}
              <p className="text-xs mt-1" style={{ color: 'var(--text-subtle)' }}>{t('sqlite.upload_hint')}</p>
            </div>
          ) : (
            <div>
              <label className={labelCls} style={labelStyle}>{t('connections.form.file_path')}</label>
              <input value={conn.file_path || ''} onChange={field('file_path')} className={inputCls} style={inputStyle} placeholder="/path/to/database.db" />
            </div>
          )}
        </>
      ) : (
        <>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} style={inputStyle} placeholder="localhost" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 3306} onChange={field('port')} type="number" className={inputCls} style={inputStyle} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('connections.form.database')}</label>
            <input value={conn.database || ''} onChange={field('database')} className={inputCls} style={inputStyle} />
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.username')}</label>
              <input value={conn.username || ''} onChange={field('username')} className={inputCls} style={inputStyle} />
            </div>
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.password')}</label>
              <input value={conn.password || ''} onChange={field('password')} type="password" className={inputCls} style={inputStyle} disabled={!conn.save_password} placeholder={conn.save_password ? '' : '••••••'} />
            </div>
          </div>
          <div>
            <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
              <input
                type="checkbox"
                checked={!!conn.save_password}
                onChange={e => onChange({ ...conn, save_password: e.target.checked, password: e.target.checked ? (conn.password ?? '') : '' })}
              />
              {t('connections.form.save_password')}
            </label>
            {conn.save_password && (
              <div
                className="mt-1 px-3 py-2 text-xs rounded border"
                style={{
                  background: 'var(--warning)18',
                  borderColor: 'var(--warning)',
                  color: 'var(--warning)',
                }}
              >
                {t('connections.form.save_password_warn')}
              </div>
            )}
          </div>
        </>
      )}
      {conn.driver === 'redis' && (
        <>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} style={inputStyle} placeholder="127.0.0.1" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 6379} onChange={field('port')} type="number" className={inputCls} style={inputStyle} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('redis.form.password')}</label>
            <input
              value={conn.password || ''}
              onChange={field('password')}
              type="password"
              className={inputCls}
              style={inputStyle}
              placeholder={t('redis.form.password_placeholder')}
            />
          </div>
          <div className="flex gap-2 items-end">
            <div className="w-32">
              <label className={labelCls} style={labelStyle}>{t('redis.form.db_index')}</label>
              <input
                value={conn.redis_db_index ?? 0}
                onChange={e => onChange({ ...conn, redis_db_index: Math.max(0, Math.min(15, parseInt(e.target.value) || 0)) })}
                type="number" min="0" max="15"
                className={inputCls}
                style={inputStyle}
              />
            </div>
            <div className="pb-0.5 text-xs" style={{ color: 'var(--text-subtle)' }}>{t('redis.form.db_index_hint')}</div>
          </div>
          <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
            <input
              type="checkbox"
              checked={!!conn.tls_enabled}
              onChange={e => onChange({ ...conn, tls_enabled: e.target.checked })}
            />
            {t('redis.form.tls')}
          </label>
          <div
            className="px-3 py-2 rounded text-xs"
            style={{ background: 'var(--warning)18', border: '1px solid var(--warning)44', color: 'var(--warning)' }}
          >
            {t('redis.driver_coming_soon')}
          </div>
        </>
      )}
      <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
        <input
          type="checkbox"
          checked={!!conn.readonly}
          onChange={e => onChange({ ...conn, readonly: e.target.checked })}
        />
        {t('connections.form.readonly')}
        <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>— {t('connections.form.readonly_hint')}</span>
      </label>
    </div>
  )
}
