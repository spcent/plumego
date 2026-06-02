import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useToast } from '../components/toastContext'
import { useI18n } from '../i18nContext'
import { DatabaseIcon } from '../components/Icons'

const emptyConn: Partial<Connection> = {
  driver: 'mysql',
  host: 'localhost',
  port: 3306,
  database: '',
  username: 'root',
  password: '',
  name: '',
}

type DriverFilter = 'all' | 'mysql' | 'sqlite' | 'redis' | 'mongodb' | 'elasticsearch'
type TestStatus = { state: 'testing' | 'ok' | 'error'; message: string }

export default function ConnectionsPage() {
  const [conns, setConns] = useState<Connection[]>([])
  const [driverFilter, setDriverFilter] = useState<DriverFilter>('all')
  const [editing, setEditing] = useState<Partial<Connection> | null>(null)
  const [testing, setTesting] = useState<Record<string, TestStatus>>({})
  const [deleteTarget, setDeleteTarget] = useState<Connection | null>(null)
  const [deleteFile, setDeleteFile] = useState(false)
  const navigate = useNavigate()
  const { showToast } = useToast()
  const { t } = useI18n()

  const reload = useCallback(() => {
    api.listConnections().then(setConns).catch(e => showToast(e.message))
  }, [showToast])

  useEffect(() => {
    const id = window.setTimeout(reload, 0)
    return () => window.clearTimeout(id)
  }, [reload])

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
    setTesting(p => ({ ...p, [id]: { state: 'testing', message: 'Testing' } }))
    try {
      const r = await api.testConnection(id)
      setTesting(p => ({ ...p, [id]: r.ok ? { state: 'ok', message: 'Connected' } : { state: 'error', message: r.error || 'Connection failed' } }))
    } catch (e) {
      setTesting(p => ({ ...p, [id]: { state: 'error', message: e instanceof Error ? e.message : 'failed' } }))
    }
  }

  async function handleDuplicate(c: Connection) {
    const { name, ...restWithID } = c
    const rest: Partial<Connection> = { ...restWithID }
    delete rest.id
    try {
      await api.createConnection({ ...rest, name: `${name} (copy)` })
      reload()
    } catch (e) {
      showToast(e instanceof Error ? e.message : 'Duplicate failed')
    }
  }

  const filteredConns = driverFilter === 'all' ? conns : conns.filter(c => c.driver === driverFilter)

  return (
    <div className="p-6">
      <div className="mb-5 flex max-w-6xl items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold" style={{ color: 'var(--text-strong)' }}>{t('connections.title')}</h1>
          <p className="mt-1 text-sm" style={{ color: 'var(--text-muted)' }}>
            {conns.length} connection{conns.length === 1 ? '' : 's'} configured
          </p>
        </div>
        <button
          onClick={() => setEditing({ ...emptyConn })}
          className="btn btn-primary"
        >
          {t('connections.add')}
        </button>
      </div>

      {conns.length > 0 && (
        <div className="mb-4 flex max-w-6xl flex-wrap gap-1 rounded-md border p-1" style={{ borderColor: 'var(--border-subtle)', background: 'var(--bg-muted)' }}>
          {(['all', 'mysql', 'sqlite', 'redis', 'mongodb', 'elasticsearch'] as DriverFilter[]).map(f => (
            <button
              key={f}
              onClick={() => setDriverFilter(f)}
              className="rounded px-3 py-1.5 text-xs font-medium transition-colors"
              style={driverFilter === f
                ? { background: 'var(--bg-surface)', color: 'var(--accent)', boxShadow: 'var(--shadow-sm)' }
                : { color: 'var(--text-muted)' }
              }
            >
              {t(`connections.filter.${f}`)}
            </button>
          ))}
        </div>
      )}

      <div className="grid max-w-6xl gap-2">
        {filteredConns.map(c => (
          <div
            key={c.id}
            className="panel flex items-center gap-4 px-4 py-3"
          >
            <div
              className="grid h-10 w-10 shrink-0 place-items-center rounded-md border"
              style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)', color: 'var(--accent)' }}
            >
              <DatabaseIcon className="h-5 w-5" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium flex items-center gap-2" style={{ color: 'var(--text-strong)' }}>
                {c.name}
                <span className="badge font-mono uppercase">{c.driver}</span>
                {c.readonly && (
                  <span className="badge" style={{ background: 'var(--warning-soft)', color: 'var(--warning)' }}>
                    {t('readonly.badge')}
                  </span>
                )}
              </div>
              <div className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>
                {c.driver === 'sqlite'
                  ? (c.uploaded_file ? (c.original_filename || c.file_path) : c.file_path)
                  : c.driver === 'mongodb'
                    ? (c.mongo_uri || `${c.host}:${c.port}`)
                    : c.driver === 'elasticsearch'
                      ? (c.es_nodes?.join(', ') || `${c.host}:${c.port}`)
                      : `${c.host}:${c.port}/${c.database}`}
              </div>
            </div>
            {testing[c.id] && (
              <span
                className="badge max-w-[220px] shrink-0 truncate"
                style={{
                  color: testing[c.id].state === 'ok' ? 'var(--success)' : testing[c.id].state === 'error' ? 'var(--danger)' : 'var(--text-muted)',
                  background: testing[c.id].state === 'ok' ? 'var(--success-soft)' : testing[c.id].state === 'error' ? 'var(--danger-soft)' : 'var(--bg-muted)',
                  borderColor: testing[c.id].state === 'ok' ? 'var(--success-border)' : testing[c.id].state === 'error' ? 'var(--danger-border)' : 'var(--border-subtle)',
                }}
              >
                {testing[c.id].message}
              </span>
            )}
            <div className="flex gap-2 shrink-0 flex-wrap justify-end">
              <button onClick={() => navigate(`/conn/${c.id}/query`)} className="btn btn-primary h-8 px-3 text-xs">{t('connections.open')}</button>
              {c.driver === 'sqlite' && c.uploaded_file && (
                <a
                  href={api.sqliteDownloadURL(c.id)}
                  className="btn btn-ghost h-8 px-3 text-xs"
                >
                  {t('sqlite.download')}
                </a>
              )}
              <button onClick={() => handleTest(c.id)} className="btn btn-ghost h-8 px-3 text-xs">{t('connections.test')}</button>
              <button
                onClick={() => setEditing({ ...c })}
                className="btn btn-ghost h-8 px-3 text-xs"
              >
                {t('connections.edit')}
              </button>
              <button onClick={() => handleDuplicate(c)} className="btn btn-ghost h-8 px-3 text-xs">{t('connections.duplicate')}</button>
              <button onClick={() => openDeleteDialog(c)} className="btn btn-danger h-8 px-3 text-xs">{t('connections.delete')}</button>
            </div>
          </div>
        ))}
        {filteredConns.length === 0 && conns.length > 0 && (
          <div className="panel-flat max-w-6xl py-12 text-center" style={{ color: 'var(--text-muted)' }}>
            No {driverFilter} connections.
          </div>
        )}
        {conns.length === 0 && (
          <div className="panel-flat max-w-6xl py-12 text-center" style={{ color: 'var(--text-muted)' }}>
            {t('connections.empty')}
          </div>
        )}
      </div>

      {editing && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="panel w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto p-6">
            <h2 className="mb-4 text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>
              {editing.id ? t('connections.edit_title') : t('connections.add_title')}
            </h2>
            <ConnectionForm conn={editing} onChange={setEditing} />
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setEditing(null)}
                className="btn btn-ghost"
              >
                {t('connections.form.cancel')}
              </button>
              <button onClick={handleSave} className="btn btn-primary">{t('connections.form.save')}</button>
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
        className="panel w-full max-w-sm mx-4 p-6"
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
            className="btn btn-ghost"
          >
            {t('confirm.cancel')}
          </button>
          <button
            onClick={onConfirm}
            className="btn btn-danger"
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

  const inputCls = 'input'

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
        <input value={conn.name || ''} onChange={field('name')} className={inputCls} placeholder="My Database" />
      </div>
      <div>
        <label className={labelCls} style={labelStyle}>{t('connections.form.driver')}</label>
        <select value={conn.driver || 'mysql'} onChange={field('driver')} className="select">
          <option value="mysql">MySQL</option>
          <option value="sqlite">SQLite</option>
          <option value="redis">Redis</option>
          <option value="mongodb">MongoDB</option>
          <option value="elasticsearch">Elasticsearch</option>
        </select>
      </div>
      {conn.driver === 'sqlite' ? (
        <>
          <div className="flex gap-1 rounded-md border p-1" style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)' }}>
            <button
              type="button"
              onClick={() => switchMode(false)}
              className="rounded px-3 py-1.5 text-xs font-medium transition-colors"
              style={!uploadMode ? { background: 'var(--bg-surface)', color: 'var(--accent)', boxShadow: 'var(--shadow-sm)' } : { color: 'var(--text-muted)' }}
            >
              {t('sqlite.mode_path')}
            </button>
            <button
              type="button"
              onClick={() => switchMode(true)}
              className="rounded px-3 py-1.5 text-xs font-medium transition-colors"
              style={uploadMode ? { background: 'var(--bg-surface)', color: 'var(--accent)', boxShadow: 'var(--shadow-sm)' } : { color: 'var(--text-muted)' }}
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
                    className="btn btn-ghost h-7 px-2 text-xs disabled:opacity-50"
                  >
                    {uploading ? t('sqlite.uploading') : t('sqlite.replace')}
                  </button>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploading}
                  className="w-full rounded-md border-2 border-dashed px-4 py-6 text-center text-sm transition-colors hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:opacity-50"
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
              <input value={conn.file_path || ''} onChange={field('file_path')} className={inputCls} placeholder="/path/to/database.db" />
            </div>
          )}
        </>
      ) : conn.driver === 'redis' ? (
        <>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} placeholder="127.0.0.1" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 6379} onChange={field('port')} type="number" className={inputCls} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('redis.form.password')}</label>
            <input
              value={conn.password || ''}
              onChange={field('password')}
              type="password"
              className={inputCls}
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
      ) : conn.driver === 'mongodb' ? (
        <>
          <div>
            <label className={labelCls} style={labelStyle}>{t('mongodb.form.uri')}</label>
            <input
              value={conn.mongo_uri || ''}
              onChange={field('mongo_uri')}
              className={inputCls}
              placeholder="mongodb://localhost:27017 or mongodb+srv://..."
            />
            <p className="text-xs mt-1" style={{ color: 'var(--text-subtle)' }}>
              {t('mongodb.form.uri_hint')}
            </p>
          </div>
          <div
            className="text-center text-xs py-1.5"
            style={{ color: 'var(--text-subtle)' }}
          >
            {t('mongodb.form.or_basic')}
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} placeholder="localhost" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 27017} onChange={field('port')} type="number" className={inputCls} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('mongodb.form.auth_db')}</label>
            <input
              value={conn.mongo_auth_db || 'admin'}
              onChange={field('mongo_auth_db')}
              className={inputCls}
              placeholder="admin"
            />
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.username')}</label>
              <input value={conn.username || ''} onChange={field('username')} className={inputCls} />
            </div>
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.password')}</label>
              <input
                value={conn.password || ''}
                onChange={field('password')}
                type="password"
                className={inputCls}
                disabled={!conn.save_password}
                placeholder={conn.save_password ? '' : '••••••'}
              />
            </div>
          </div>
          <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
            <input
              type="checkbox"
              checked={!!conn.save_password}
              onChange={e => onChange({ ...conn, save_password: e.target.checked, password: e.target.checked ? (conn.password ?? '') : '' })}
            />
            {t('connections.form.save_password')}
          </label>
          <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
            <input
              type="checkbox"
              checked={!!conn.mongo_tls_enabled}
              onChange={e => onChange({ ...conn, mongo_tls_enabled: e.target.checked })}
            />
            {t('mongodb.form.tls')}
          </label>
          <div
            className="px-3 py-2 rounded text-xs"
            style={{ background: 'var(--warning)18', border: '1px solid var(--warning)44', color: 'var(--warning)' }}
          >
            {t('mongodb.driver_coming_soon')}
          </div>
        </>
      ) : conn.driver === 'elasticsearch' ? (
        <>
          <div>
            <label className={labelCls} style={labelStyle}>{t('elasticsearch.form.nodes')}</label>
            <input
              value={(conn.es_nodes || []).join(', ')}
              onChange={e => onChange({ ...conn, es_nodes: e.target.value.split(',').map(s => s.trim()).filter(Boolean) })}
              className={inputCls}
              placeholder="http://localhost:9200, http://node2:9200"
            />
            <p className="text-xs mt-1" style={{ color: 'var(--text-subtle)' }}>
              {t('elasticsearch.form.nodes_hint')}
            </p>
          </div>
          <div
            className="text-center text-xs py-1.5"
            style={{ color: 'var(--text-subtle)' }}
          >
            {t('elasticsearch.form.or_basic')}
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} placeholder="localhost" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 9200} onChange={field('port')} type="number" className={inputCls} />
            </div>
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.username')}</label>
              <input value={conn.username || ''} onChange={field('username')} className={inputCls} />
            </div>
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.password')}</label>
              <input value={conn.password || ''} onChange={field('password')} type="password" className={inputCls} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('elasticsearch.form.api_key')}</label>
            <input
              value={conn.es_api_key || ''}
              onChange={field('es_api_key')}
              type="password"
              className={inputCls}
              placeholder={t('elasticsearch.form.api_key_placeholder')}
            />
          </div>
          <label className="flex items-center gap-2 cursor-pointer text-sm" style={{ color: 'var(--text-default)' }}>
            <input
              type="checkbox"
              checked={!!conn.es_insecure_skip_tls}
              onChange={e => onChange({ ...conn, es_insecure_skip_tls: e.target.checked })}
            />
            {t('elasticsearch.form.insecure_skip_tls')}
          </label>
          <div
            className="px-3 py-2 rounded text-xs"
            style={{ background: 'var(--warning)18', border: '1px solid var(--warning)44', color: 'var(--warning)' }}
          >
            {t('elasticsearch.driver_coming_soon')}
          </div>
        </>
      ) : (
        <>
          {/* MySQL (default) */}
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.host')}</label>
              <input value={conn.host || ''} onChange={field('host')} className={inputCls} placeholder="localhost" />
            </div>
            <div className="w-24">
              <label className={labelCls} style={labelStyle}>{t('connections.form.port')}</label>
              <input value={conn.port || 3306} onChange={field('port')} type="number" className={inputCls} />
            </div>
          </div>
          <div>
            <label className={labelCls} style={labelStyle}>{t('connections.form.database')}</label>
            <input value={conn.database || ''} onChange={field('database')} className={inputCls} />
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.username')}</label>
              <input value={conn.username || ''} onChange={field('username')} className={inputCls} />
            </div>
            <div className="flex-1">
              <label className={labelCls} style={labelStyle}>{t('connections.form.password')}</label>
              <input value={conn.password || ''} onChange={field('password')} type="password" className={inputCls} disabled={!conn.save_password} placeholder={conn.save_password ? '' : '••••••'} />
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
