import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'
import { api, type RedisKeyEntry, type RedisKeyDetail } from '../api'
import { useToast } from '../components/Toast'

type ViewMode = 'inspector' | 'command'

// Pattern favorites stored in localStorage
interface PatternFavorite {
  name: string
  pattern: string
}

const FAVORITES_KEY = 'redis_pattern_favorites'

function loadFavorites(): PatternFavorite[] {
  try {
    const stored = localStorage.getItem(FAVORITES_KEY)
    return stored ? JSON.parse(stored) : []
  } catch {
    return []
  }
}

function saveFavorites(favorites: PatternFavorite[]) {
  localStorage.setItem(FAVORITES_KEY, JSON.stringify(favorites))
}

export default function RedisKeyPanel() {
  const { connId, redisDb } = useParams<{ connId: string; redisDb: string }>()
  const { t } = useI18n()
  const conn = useCurrentConn(connId)
  const toast = useToast()
  const dbIndex = parseInt(redisDb ?? '0', 10)

  const [keys, setKeys] = useState<RedisKeyEntry[]>([])
  const [cursor, setCursor] = useState<number>(0)
  const [done, setDone] = useState(false)
  const [loading, setLoading] = useState(false)
  const [pattern, setPattern] = useState('*')
  const [selectedKey, setSelectedKey] = useState<RedisKeyDetail | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('inspector')
  const [commandInput, setCommandInput] = useState('')
  const [commandResult, setCommandResult] = useState<unknown>(null)
  const [commandRunning, setCommandRunning] = useState(false)

  // Batch delete state
  const [batchMode, setBatchMode] = useState(false)
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set())
  const [showBatchPreview, setShowBatchPreview] = useState(false)
  const [previewKeys, setPreviewKeys] = useState<RedisKeyEntry[]>([])

  // Pattern favorites
  const [favorites, setFavorites] = useState<PatternFavorite[]>(loadFavorites())
  const [showFavorites, setShowFavorites] = useState(false)
  const [newFavoriteName, setNewFavoriteName] = useState('')

  const isReadonly = conn?.readonly ?? false

  // Load keys
  const loadKeys = useCallback(async (reset = false) => {
    if (!connId || !redisDb) return
    setLoading(true)
    try {
      const resp = await api.redisListKeys(connId, dbIndex, {
        pattern,
        cursor: reset ? 0 : cursor,
        count: 100,
      })
      if (reset) {
        setKeys(resp.keys)
      } else {
        setKeys(prev => [...prev, ...resp.keys])
      }
      setCursor(resp.nextCursor)
      setDone(resp.done)
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to load keys', 'error')
    } finally {
      setLoading(false)
    }
  }, [connId, dbIndex, pattern, cursor])

  useEffect(() => {
    loadKeys(true)
  }, [connId, dbIndex, pattern])

  // Load key details
  const loadKeyDetail = useCallback(async (key: string) => {
    if (!connId) return
    try {
      const detail = await api.redisGetKey(connId, dbIndex, key)
      setSelectedKey(detail)
      setViewMode('inspector')
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to load key details', 'error')
    }
  }, [connId, dbIndex, toast])

  // Delete key
  const handleDelete = useCallback(async (key: string) => {
    if (!connId || isReadonly) return
    if (!window.confirm(t('redis.key.delete_confirm', { key }))) return
    try {
      await api.redisDeleteKey(connId, dbIndex, key)
      setKeys(prev => prev.filter(k => k.key !== key))
      if (selectedKey?.key === key) setSelectedKey(null)
      toast.showToast('Key deleted', 'success')
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to delete key', 'error')
    }
  }, [connId, dbIndex, isReadonly, selectedKey, t, toast])

  // Set TTL
  const handleSetTTL = useCallback(async (ttl: number) => {
    if (!connId || !selectedKey || isReadonly) return
    try {
      await api.redisSetTTL(connId, dbIndex, selectedKey.key, ttl)
      setSelectedKey(prev => prev ? { ...prev, ttl } : null)
      toast.showToast('TTL updated', 'success')
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to set TTL', 'error')
    }
  }, [connId, dbIndex, selectedKey, isReadonly, toast])

  // Run command
  const handleCommand = useCallback(async () => {
    if (!connId || !commandInput.trim()) return
    setCommandRunning(true)
    try {
      const resp = await api.redisCommand(connId, dbIndex, commandInput.trim())
      setCommandResult(resp)
      // Refresh key list after write commands
      if (!isReadonly && /^(SET|DEL|EXPIRE|PERSIST|HSET|LPUSH|SADD|ZADD)/i.test(commandInput)) {
        await loadKeys(true)
      }
    } catch (err: any) {
      setCommandResult({ error: err.message })
    } finally {
      setCommandRunning(false)
    }
  }, [connId, dbIndex, commandInput, isReadonly, loadKeys])

  // Batch delete handlers
  const handleBatchPreview = useCallback(async () => {
    if (!connId) return
    try {
      const resp = await api.redisBatchPreview(connId, dbIndex, pattern, 100)
      setPreviewKeys(resp.keys)
      setShowBatchPreview(true)
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to preview keys', 'error')
    }
  }, [connId, dbIndex, pattern, toast])

  const handleBatchDelete = useCallback(async () => {
    if (!connId || isReadonly || previewKeys.length === 0) return
    const count = previewKeys.length
    if (!window.confirm(t('redis.batch.confirm_delete', { n: count }))) return

    try {
      const keysToDelete = previewKeys.map(k => k.key)
      await api.redisBatchDelete(connId, dbIndex, keysToDelete, true)
      toast.showToast(t('redis.batch.deleted', { n: count }), 'success')
      setSelectedKeys(new Set())
      setBatchMode(false)
      setShowBatchPreview(false)
      setPreviewKeys([])
      await loadKeys(true)
    } catch (err: any) {
      toast.showToast(err.message || 'Failed to delete keys', 'error')
    }
  }, [connId, dbIndex, previewKeys, isReadonly, loadKeys, t, toast])

  const toggleKeySelection = (key: string) => {
    setSelectedKeys(prev => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  const toggleSelectAll = () => {
    if (selectedKeys.size === keys.length) {
      setSelectedKeys(new Set())
    } else {
      setSelectedKeys(new Set(keys.map(k => k.key)))
    }
  }

  // Pattern favorites handlers
  const handleSaveFavorite = () => {
    if (!newFavoriteName.trim() || !pattern.trim()) return
    const newFav = { name: newFavoriteName.trim(), pattern: pattern.trim() }
    const updated = [...favorites, newFav]
    setFavorites(updated)
    saveFavorites(updated)
    setNewFavoriteName('')
    toast.showToast(t('redis.favorites.saved'), 'success')
  }

  const handleDeleteFavorite = (name: string) => {
    const updated = favorites.filter(f => f.name !== name)
    setFavorites(updated)
    saveFavorites(updated)
  }

  const handleSelectFavorite = (pattern: string) => {
    setPattern(pattern)
    setShowFavorites(false)
    loadKeys(true)
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.showToast('Copied', 'success')
  }

  const formatTTL = (ttl: number) => {
    if (ttl === -1) return t('redis.browser.ttl_none')
    if (ttl === -2) return 'N/A'
    return t('redis.browser.ttl_sec', { n: ttl })
  }

  const formatMemory = (bytes: number) => {
    if (bytes < 1024) return t('redis.memory.bytes', { n: bytes })
    if (bytes < 1024 * 1024) return t('redis.memory.kb', { n: (bytes / 1024).toFixed(1) })
    if (bytes < 1024 * 1024 * 1024) return t('redis.memory.mb', { n: (bytes / 1024 / 1024).toFixed(1) })
    return t('redis.memory.gb', { n: (bytes / 1024 / 1024 / 1024).toFixed(2) })
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-4 py-3 border-b flex items-center gap-3" style={{ borderColor: 'var(--border-subtle)' }}>
        <span className="text-lg font-mono" style={{ color: '#f87171' }}>⚡</span>
        <h2 className="text-base font-semibold" style={{ color: 'var(--text-strong)' }}>
          {t('redis.browser.db', { n: dbIndex })}
        </h2>
        {conn && (
          <span className="text-xs font-mono px-2 py-0.5 rounded"
            style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {conn.name}
          </span>
        )}
        {isReadonly && (
          <span className="text-xs px-2 py-0.5 rounded"
            style={{ background: 'var(--warning)18', color: 'var(--warning)', border: '1px solid var(--warning)44' }}>
            {t('redis.readonly')}
          </span>
        )}
      </div>

      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left: Key Browser */}
        <div className="w-80 border-r flex flex-col" style={{ borderColor: 'var(--border-subtle)' }}>
          {/* Search */}
          <div className="p-3 border-b" style={{ borderColor: 'var(--border-subtle)' }}>
            <div className="flex gap-2 mb-2">
              <input
                type="text"
                value={pattern}
                onChange={e => setPattern(e.target.value)}
                placeholder={t('redis.browser.pattern_placeholder')}
                className="flex-1 px-3 py-1.5 text-sm rounded border"
                style={{
                  background: 'var(--bg-surface)',
                  color: 'var(--text-default)',
                  borderColor: 'var(--border-subtle)',
                }}
                onKeyDown={e => e.key === 'Enter' && loadKeys(true)}
              />
              <button
                onClick={() => loadKeys(true)}
                disabled={loading}
                className="px-3 py-1.5 text-sm rounded border"
                style={{
                  background: 'var(--bg-surface)',
                  color: 'var(--text-default)',
                  borderColor: 'var(--border-subtle)',
                }}
              >
                {loading ? t('redis.browser.loading') : t('redis.browser.search')}
              </button>
            </div>

            {/* Pattern favorites and batch mode buttons */}
            <div className="flex gap-2 mb-2">
              <button
                onClick={() => setShowFavorites(!showFavorites)}
                className="flex-1 px-2 py-1 text-xs rounded border"
                style={{
                  background: 'var(--bg-surface)',
                  color: 'var(--text-default)',
                  borderColor: 'var(--border-subtle)',
                }}
                title={t('redis.favorites.title')}
              >
                ⭐ {t('redis.favorites.title')}
              </button>
              {!isReadonly && (
                <button
                  onClick={() => setBatchMode(!batchMode)}
                  className="flex-1 px-2 py-1 text-xs rounded border"
                  style={{
                    background: batchMode ? 'var(--accent)' : 'var(--bg-surface)',
                    color: batchMode ? 'white' : 'var(--text-default)',
                    borderColor: 'var(--border-subtle)',
                  }}
                >
                  {batchMode ? t('redis.batch.cancel') : t('redis.batch.select_keys')}
                </button>
              )}
            </div>

            {/* Favorites dropdown */}
            {showFavorites && (
              <div className="mb-2 p-2 rounded border" style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)' }}>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={newFavoriteName}
                    onChange={e => setNewFavoriteName(e.target.value)}
                    placeholder={t('redis.favorites.name')}
                    className="flex-1 px-2 py-1 text-xs rounded border"
                    style={{
                      background: 'var(--bg-muted)',
                      color: 'var(--text-default)',
                      borderColor: 'var(--border-subtle)',
                    }}
                  />
                  <button
                    onClick={handleSaveFavorite}
                    className="px-2 py-1 text-xs rounded"
                    style={{ background: 'var(--accent)', color: 'white' }}
                  >
                    {t('redis.favorites.save')}
                  </button>
                </div>
                {favorites.length === 0 ? (
                  <div className="text-xs text-center py-2" style={{ color: 'var(--text-muted)' }}>
                    {t('redis.favorites.no_favorites')}
                  </div>
                ) : (
                  <div className="space-y-1 max-h-32 overflow-y-auto">
                    {favorites.map(fav => (
                      <div
                        key={fav.name}
                        className="flex items-center gap-2 text-xs p-1 rounded hover:bg-opacity-50"
                        style={{ background: 'var(--bg-muted)' }}
                      >
                        <button
                          onClick={() => handleSelectFavorite(fav.pattern)}
                          className="flex-1 text-left truncate"
                          style={{ color: 'var(--text-default)' }}
                          title={fav.pattern}
                        >
                          {fav.name}
                        </button>
                        <button
                          onClick={() => handleDeleteFavorite(fav.name)}
                          className="shrink-0"
                          style={{ color: 'var(--danger)' }}
                          title={t('redis.favorites.delete')}
                        >
                          ×
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Batch mode controls */}
            {batchMode && (
              <div className="mb-2 p-2 rounded border" style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)' }}>
                <div className="flex gap-2 mb-2">
                  <button
                    onClick={toggleSelectAll}
                    className="flex-1 px-2 py-1 text-xs rounded border"
                    style={{
                      background: 'var(--bg-muted)',
                      color: 'var(--text-default)',
                      borderColor: 'var(--border-subtle)',
                    }}
                  >
                    {selectedKeys.size === keys.length ? t('redis.batch.deselect_all') : t('redis.batch.select_all')}
                  </button>
                  <button
                    onClick={handleBatchPreview}
                    disabled={selectedKeys.size === 0}
                    className="flex-1 px-2 py-1 text-xs rounded border"
                    style={{
                      background: 'var(--bg-muted)',
                      color: 'var(--text-default)',
                      borderColor: 'var(--border-subtle)',
                      opacity: selectedKeys.size === 0 ? 0.5 : 1,
                    }}
                  >
                    {t('redis.batch.preview')}
                  </button>
                </div>
                {selectedKeys.size > 0 && (
                  <button
                    onClick={handleBatchDelete}
                    className="w-full px-2 py-1.5 text-xs rounded"
                    style={{ background: 'var(--danger)', color: 'white' }}
                  >
                    {t('redis.batch.delete')} ({selectedKeys.size})
                  </button>
                )}
                <div className="mt-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                  {t('redis.batch.selected', { n: selectedKeys.size })}
                </div>
              </div>
            )}

            <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.browser.key_count', { n: keys.length })}
            </div>
          </div>

          {/* Key list */}
          <div className="flex-1 overflow-y-auto">
            {keys.length === 0 && !loading && (
              <div className="p-4 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
                {t('redis.browser.empty')}
              </div>
            )}
            {keys.map(key => (
              <div
                key={key.key}
                onClick={() => !batchMode && loadKeyDetail(key.key)}
                className="px-3 py-2 border-b cursor-pointer hover:bg-opacity-50 transition-colors"
                style={{
                  borderColor: 'var(--border-subtle)',
                  background: selectedKey?.key === key.key ? 'var(--bg-muted)' : 'transparent',
                }}
              >
                <div className="flex items-start gap-2">
                  {batchMode && (
                    <input
                      type="checkbox"
                      checked={selectedKeys.has(key.key)}
                      onChange={e => {
                        e.stopPropagation()
                        toggleKeySelection(key.key)
                      }}
                      onClick={e => e.stopPropagation()}
                      className="mt-1"
                    />
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-mono truncate" style={{ color: 'var(--text-strong)' }}>
                      {key.key}
                    </div>
                    <div className="flex gap-2 mt-1 text-xs" style={{ color: 'var(--text-muted)' }}>
                      <span className="px-1.5 py-0.5 rounded" style={{ background: 'var(--bg-surface)' }}>
                        {key.type}
                      </span>
                      <span>{formatTTL(key.ttl)}</span>
                      {key.memory && (
                        <span className={key.isBig ? 'font-bold' : ''} style={{ color: key.isBig ? 'var(--danger)' : 'var(--text-muted)' }}>
                          {formatMemory(key.memory)}
                        </span>
                      )}
                    </div>
                  </div>
                  {!isReadonly && !batchMode && (
                    <button
                      onClick={e => {
                        e.stopPropagation()
                        handleDelete(key.key)
                      }}
                      className="text-xs px-2 py-0.5 rounded hover:bg-red-100 dark:hover:bg-red-900"
                      style={{ color: 'var(--danger)' }}
                      title={t('redis.key.delete')}
                    >
                      ×
                    </button>
                  )}
                </div>
              </div>
            ))}
            {!done && !loading && (
              <button
                onClick={() => loadKeys(false)}
                className="w-full py-2 text-sm border-t hover:bg-opacity-50"
                style={{ color: 'var(--text-muted)', borderColor: 'var(--border-subtle)' }}
              >
                {t('redis.browser.load_more')}
              </button>
            )}
          </div>
        </div>

        {/* Right: Inspector / Command Console */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Tabs */}
          <div className="flex border-b" style={{ borderColor: 'var(--border-subtle)' }}>
            <button
              onClick={() => setViewMode('inspector')}
              className="px-4 py-2 text-sm font-medium border-b-2"
              style={{
                color: viewMode === 'inspector' ? 'var(--accent)' : 'var(--text-muted)',
                borderColor: viewMode === 'inspector' ? 'var(--accent)' : 'transparent',
              }}
            >
              {t('redis.key.type')}
            </button>
            <button
              onClick={() => setViewMode('command')}
              className="px-4 py-2 text-sm font-medium border-b-2"
              style={{
                color: viewMode === 'command' ? 'var(--accent)' : 'var(--text-muted)',
                borderColor: viewMode === 'command' ? 'var(--accent)' : 'transparent',
              }}
            >
              {t('redis.command.title')}
            </button>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-4">
            {viewMode === 'inspector' ? (
              selectedKey ? (
                <KeyInspector
                  detail={selectedKey}
                  isReadonly={isReadonly}
                  onSetTTL={handleSetTTL}
                  onDelete={() => handleDelete(selectedKey.key)}
                  onCopy={copyToClipboard}
                  t={t}
                />
              ) : (
                <div className="h-full flex items-center justify-center text-sm" style={{ color: 'var(--text-muted)' }}>
                  {t('redis.key.not_found')}
                </div>
              )
            ) : (
              <CommandConsole
                input={commandInput}
                onInputChange={setCommandInput}
                onRun={handleCommand}
                running={commandRunning}
                result={commandResult}
                isReadonly={isReadonly}
                t={t}
              />
            )}
          </div>
        </div>
      </div>

      {/* Batch Preview Modal */}
      {showBatchPreview && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[80vh] flex flex-col">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold" style={{ color: 'var(--text-strong)' }}>
                {t('redis.batch.preview')}
              </h3>
              <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
                {t('redis.batch.warning', { n: previewKeys.length })}
              </p>
            </div>
            <div className="flex-1 overflow-y-auto p-6">
              {previewKeys.length === 0 ? (
                <div className="text-center py-8" style={{ color: 'var(--text-muted)' }}>
                  {t('redis.batch.no_keys')}
                </div>
              ) : (
                <div className="space-y-2">
                  {previewKeys.map((key, idx) => (
                    <div
                      key={idx}
                      className="flex items-center gap-3 p-3 rounded border"
                      style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)' }}
                    >
                      <div className="flex-1 min-w-0">
                        <div className="font-mono text-sm truncate" style={{ color: 'var(--text-strong)' }}>
                          {key.key}
                        </div>
                        <div className="flex gap-2 mt-1 text-xs" style={{ color: 'var(--text-muted)' }}>
                          <span className="px-1.5 py-0.5 rounded" style={{ background: 'var(--bg-muted)' }}>
                            {key.type}
                          </span>
                          <span>{formatTTL(key.ttl)}</span>
                          {key.memory && (
                            <span className={key.isBig ? 'font-bold' : ''} style={{ color: key.isBig ? 'var(--danger)' : 'var(--text-muted)' }}>
                              {formatMemory(key.memory)}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
            <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex gap-3">
              <button
                onClick={() => {
                  setShowBatchPreview(false)
                  setPreviewKeys([])
                }}
                className="flex-1 px-4 py-2 text-sm rounded border"
                style={{
                  background: 'var(--bg-surface)',
                  color: 'var(--text-default)',
                  borderColor: 'var(--border-subtle)',
                }}
              >
                {t('redis.batch.cancel')}
              </button>
              <button
                onClick={handleBatchDelete}
                className="flex-1 px-4 py-2 text-sm rounded"
                style={{ background: 'var(--danger)', color: 'white' }}
              >
                {t('redis.batch.delete')} ({previewKeys.length})
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// Key Inspector Component
function KeyInspector({
  detail,
  isReadonly,
  onSetTTL,
  onDelete,
  onCopy,
  t,
}: {
  detail: RedisKeyDetail
  isReadonly: boolean
  onSetTTL: (ttl: number) => void
  onDelete: () => void
  onCopy: (text: string) => void
  t: (key: string, vars?: Record<string, any>) => string
}) {
  const [ttlInput, setTTLInput] = useState(String(detail.ttl))

  useEffect(() => {
    setTTLInput(String(detail.ttl))
  }, [detail.ttl])

  return (
    <div className="space-y-4">
      {/* Header */}
      <div>
        <div className="flex items-center gap-2 mb-2">
          <span className="text-xs font-mono px-2 py-0.5 rounded" style={{ background: 'var(--bg-muted)' }}>
            {detail.type}
          </span>
          {detail.encoding && (
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.key.encoding')}: {detail.encoding}
            </span>
          )}
        </div>
        <div className="text-sm font-mono break-all" style={{ color: 'var(--text-strong)' }}>
          {detail.key}
        </div>
        <button
          onClick={() => onCopy(detail.key)}
          className="mt-1 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.key_success')}
        >
          {t('redis.copy.key')}
        </button>
      </div>

      {/* TTL */}
      <div>
        <label className="block text-sm mb-1" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.ttl')}
        </label>
        <div className="flex gap-2 mb-2">
          <input
            type="number"
            value={ttlInput}
            onChange={e => setTTLInput(e.target.value)}
            disabled={isReadonly}
            className="px-3 py-1.5 text-sm rounded border"
            style={{
              background: 'var(--bg-surface)',
              color: 'var(--text-default)',
              borderColor: 'var(--border-subtle)',
            }}
            placeholder={t('redis.key.ttl_hint')}
          />
          {!isReadonly && (
            <button
              onClick={() => onSetTTL(parseInt(ttlInput, 10))}
              className="px-3 py-1.5 text-sm rounded"
              style={{ background: 'var(--accent)', color: 'white' }}
            >
              {t('redis.key.set_ttl')}
            </button>
          )}
        </div>
        <div className="text-xs mt-1 mb-2" style={{ color: 'var(--text-muted)' }}>
          {detail.ttl === -1 ? t('redis.browser.ttl_none') : `${detail.ttl}s`}
        </div>

        {/* TTL Quick Presets */}
        {!isReadonly && (
          <div className="flex flex-wrap gap-1">
            <span className="text-xs mr-2 self-center" style={{ color: 'var(--text-muted)' }}>
              {t('redis.ttl.presets')}:
            </span>
            <button
              onClick={() => onSetTTL(60)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.1min')}
            </button>
            <button
              onClick={() => onSetTTL(300)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.5min')}
            </button>
            <button
              onClick={() => onSetTTL(3600)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.1hour')}
            </button>
            <button
              onClick={() => onSetTTL(86400)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.1day')}
            </button>
            <button
              onClick={() => onSetTTL(604800)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.1week')}
            </button>
            <button
              onClick={() => onSetTTL(-1)}
              className="px-2 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
              }}
            >
              {t('redis.ttl.never')}
            </button>
          </div>
        )}
      </div>

      {/* Value */}
      <div>
        <label className="block text-sm mb-2" style={{ color: 'var(--text-muted)' }}>
          Value
        </label>
        <KeyValueViewer detail={detail} onCopy={onCopy} t={t} />
      </div>

      {/* Delete */}
      {!isReadonly && (
        <button
          onClick={onDelete}
          className="px-4 py-2 text-sm rounded"
          style={{ background: 'var(--danger)', color: 'white' }}
        >
          {t('redis.key.delete')}
        </button>
      )}
    </div>
  )
}

// Value viewer for different key types
function KeyValueViewer({
  detail,
  onCopy,
  t,
}: {
  detail: RedisKeyDetail
  onCopy: (text: string) => void
  t: (key: string, vars?: Record<string, any>) => string
}) {
  const { type } = detail
  const [hashFilter, setHashFilter] = useState('')
  const [listPage, setListPage] = useState(0)
  const [setPage, setSetPage] = useState(0)
  const [zsetPage, setZsetPage] = useState(0)
  const pageSize = 50

  if (type === 'string' && detail.string !== undefined) {
    return (
      <div>
        <pre className="p-3 text-sm font-mono rounded border overflow-auto max-h-96 whitespace-pre-wrap"
          style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)', color: 'var(--text-default)' }}>
          {detail.string}
        </pre>
        <button
          onClick={() => onCopy(detail.string!)}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  if (type === 'hash' && detail.hash) {
    const allEntries = Object.entries(detail.hash)
    const filteredEntries = hashFilter
      ? allEntries.filter(([field]) => field.toLowerCase().includes(hashFilter.toLowerCase()))
      : allEntries
    return (
      <div>
        <div className="flex items-center gap-2 mb-2">
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            {t('redis.key.hash_fields', { n: allEntries.length })}
          </div>
          <input
            type="text"
            value={hashFilter}
            onChange={e => setHashFilter(e.target.value)}
            placeholder={t('redis.key.hash_search')}
            className="flex-1 px-2 py-1 text-xs rounded border"
            style={{
              background: 'var(--bg-surface)',
              color: 'var(--text-default)',
              borderColor: 'var(--border-subtle)',
            }}
          />
        </div>
        <div className="border rounded overflow-hidden max-h-96 overflow-y-auto" style={{ borderColor: 'var(--border-subtle)' }}>
          {filteredEntries.length === 0 ? (
            <div className="px-3 py-4 text-center text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.key.hash_no_match')}
            </div>
          ) : (
            filteredEntries.map(([field, value]) => (
              <div key={field} className="flex border-b last:border-b-0" style={{ borderColor: 'var(--border-subtle)' }}>
                <div className="w-1/3 px-3 py-2 text-sm font-mono border-r"
                  style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)', color: 'var(--text-strong)' }}>
                  {field}
                </div>
                <div className="flex-1 px-3 py-2 text-sm font-mono break-all"
                  style={{ color: 'var(--text-default)' }}>
                  {value}
                </div>
              </div>
            ))
          )}
        </div>
        <button
          onClick={() => onCopy(JSON.stringify(detail.hash, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  if (type === 'list' && detail.list) {
    const start = listPage * pageSize
    const end = Math.min(start + pageSize, detail.list.length)
    const pageItems = detail.list.slice(start, end)
    const totalPages = Math.ceil(detail.list.length / pageSize)
    return (
      <div>
        <div className="flex items-center justify-between mb-2">
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            {t('redis.key.list_length', { n: detail.list.length })}
          </div>
          {totalPages > 1 && (
            <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.key.page', { current: listPage + 1, total: totalPages })}
            </div>
          )}
        </div>
        <div className="border rounded overflow-hidden max-h-96 overflow-y-auto" style={{ borderColor: 'var(--border-subtle)' }}>
          {pageItems.map((item, i) => (
            <div key={start + i} className="flex border-b last:border-b-0" style={{ borderColor: 'var(--border-subtle)' }}>
              <div className="w-12 px-3 py-2 text-xs font-mono border-r"
                style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)', color: 'var(--text-muted)' }}>
                {start + i}
              </div>
              <div className="flex-1 px-3 py-2 text-sm font-mono break-all"
                style={{ color: 'var(--text-default)' }}>
                {item}
              </div>
            </div>
          ))}
        </div>
        {totalPages > 1 && (
          <div className="flex gap-2 mt-2">
            <button
              onClick={() => setListPage(Math.max(0, listPage - 1))}
              disabled={listPage === 0}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: listPage === 0 ? 0.5 : 1,
              }}
            >
              {t('redis.key.prev')}
            </button>
            <button
              onClick={() => setListPage(Math.min(totalPages - 1, listPage + 1))}
              disabled={listPage === totalPages - 1}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: listPage === totalPages - 1 ? 0.5 : 1,
              }}
            >
              {t('redis.key.next')}
            </button>
          </div>
        )}
        <button
          onClick={() => onCopy(JSON.stringify(detail.list, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  if (type === 'set' && detail.set) {
    const start = setPage * pageSize
    const end = Math.min(start + pageSize, detail.set.length)
    const pageItems = detail.set.slice(start, end)
    const totalPages = Math.ceil(detail.set.length / pageSize)
    return (
      <div>
        <div className="flex items-center justify-between mb-2">
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            {t('redis.key.set_size', { n: detail.set.length })}
          </div>
          {totalPages > 1 && (
            <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.key.page', { current: setPage + 1, total: totalPages })}
            </div>
          )}
        </div>
        <div className="flex flex-wrap gap-2 max-h-96 overflow-y-auto">
          {pageItems.map((item, i) => (
            <span key={start + i} className="px-2 py-1 text-sm font-mono rounded"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)' }}>
              {item}
            </span>
          ))}
        </div>
        {totalPages > 1 && (
          <div className="flex gap-2 mt-2">
            <button
              onClick={() => setSetPage(Math.max(0, setPage - 1))}
              disabled={setPage === 0}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: setPage === 0 ? 0.5 : 1,
              }}
            >
              {t('redis.key.prev')}
            </button>
            <button
              onClick={() => setSetPage(Math.min(totalPages - 1, setPage + 1))}
              disabled={setPage === totalPages - 1}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: setPage === totalPages - 1 ? 0.5 : 1,
              }}
            >
              {t('redis.key.next')}
            </button>
          </div>
        )}
        <button
          onClick={() => onCopy(JSON.stringify(detail.set, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  if (type === 'zset' && detail.zset) {
    const start = zsetPage * pageSize
    const end = Math.min(start + pageSize, detail.zset.length)
    const pageItems = detail.zset.slice(start, end)
    const totalPages = Math.ceil(detail.zset.length / pageSize)
    return (
      <div>
        <div className="flex items-center justify-between mb-2">
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            {t('redis.key.set_size', { n: detail.zset.length })}
          </div>
          {totalPages > 1 && (
            <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {t('redis.key.page', { current: zsetPage + 1, total: totalPages })}
            </div>
          )}
        </div>
        <div className="border rounded overflow-hidden max-h-96 overflow-y-auto" style={{ borderColor: 'var(--border-subtle)' }}>
          {pageItems.map((item, i) => (
            <div key={start + i} className="flex border-b last:border-b-0" style={{ borderColor: 'var(--border-subtle)' }}>
              <div className="w-20 px-3 py-2 text-sm font-mono border-r"
                style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)', color: 'var(--text-muted)' }}>
                {item.score}
              </div>
              <div className="flex-1 px-3 py-2 text-sm font-mono break-all"
                style={{ color: 'var(--text-default)' }}>
                {item.member}
              </div>
            </div>
          ))}
        </div>
        {totalPages > 1 && (
          <div className="flex gap-2 mt-2">
            <button
              onClick={() => setZsetPage(Math.max(0, zsetPage - 1))}
              disabled={zsetPage === 0}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: zsetPage === 0 ? 0.5 : 1,
              }}
            >
              {t('redis.key.prev')}
            </button>
            <button
              onClick={() => setZsetPage(Math.min(totalPages - 1, zsetPage + 1))}
              disabled={zsetPage === totalPages - 1}
              className="px-3 py-1 text-xs rounded border"
              style={{
                background: 'var(--bg-surface)',
                color: 'var(--text-default)',
                borderColor: 'var(--border-subtle)',
                opacity: zsetPage === totalPages - 1 ? 0.5 : 1,
              }}
            >
              {t('redis.key.next')}
            </button>
          </div>
        )}
        <button
          onClick={() => onCopy(JSON.stringify(detail.zset, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  if (type === 'stream' && detail.stream) {
    const messages = detail.stream.messages || []
    const streamText = JSON.stringify({
      length: detail.stream.length,
      groups: detail.stream.groups,
      messages,
    }, null, 2)
    return (
      <div className="space-y-2">
        <div className="flex items-center gap-3 text-sm" style={{ color: 'var(--text-default)' }}>
          <span>{t('redis.key.stream_length', { n: detail.stream.length })}</span>
          <span>{t('redis.key.stream_groups', { n: detail.stream.groups })}</span>
        </div>
        {messages.length === 0 ? (
          <div className="text-xs p-3 rounded" style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
            {t('redis.stream.no_messages')}
          </div>
        ) : (
          <div className="border rounded overflow-hidden max-h-96 overflow-y-auto" style={{ borderColor: 'var(--border-subtle)' }}>
            <table className="w-full text-sm">
              <thead className="sticky top-0" style={{ background: 'var(--bg-muted)' }}>
                <tr>
                  <th className="text-left px-3 py-2 font-mono text-xs" style={{ color: 'var(--text-muted)' }}>
                    {t('redis.stream.id')}
                  </th>
                  <th className="text-left px-3 py-2 font-mono text-xs" style={{ color: 'var(--text-muted)' }}>
                    {t('redis.stream.fields')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {messages.map((msg) => (
                  <tr key={msg.id} className="border-t" style={{ borderColor: 'var(--border-subtle)' }}>
                    <td className="px-3 py-2 font-mono text-xs align-top" style={{ color: 'var(--text-strong)' }}>
                      {msg.id}
                    </td>
                    <td className="px-3 py-2 font-mono text-xs break-all" style={{ color: 'var(--text-default)' }}>
                      <pre className="whitespace-pre-wrap m-0">{JSON.stringify(msg.values, null, 2)}</pre>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {messages.length < detail.stream.length && (
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            Showing {messages.length} of {detail.stream.length} messages. Use XRANGE in the command console to view more.
          </div>
        )}
        <button
          onClick={() => onCopy(streamText)}
          className="text-xs"
          style={{ color: 'var(--accent)' }}
          title={t('redis.copy.value_success')}
        >
          {t('redis.copy.value')}
        </button>
      </div>
    )
  }

  return (
    <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
      Unsupported type: {type}
    </div>
  )
}

// Command Console Component
function CommandConsole({
  input,
  onInputChange,
  onRun,
  running,
  result,
  isReadonly,
  t,
}: {
  input: string
  onInputChange: (v: string) => void
  onRun: () => void
  running: boolean
  result: unknown
  isReadonly: boolean
  t: (key: string, vars?: Record<string, any>) => string
}) {
  return (
    <div className="space-y-4">
      {/* Input */}
      <div>
        <input
          type="text"
          value={input}
          onChange={e => onInputChange(e.target.value)}
          placeholder={t('redis.command.placeholder')}
          className="w-full px-3 py-2 text-sm font-mono rounded border"
          style={{
            background: 'var(--bg-surface)',
            color: 'var(--text-default)',
            borderColor: 'var(--border-subtle)',
          }}
          onKeyDown={e => e.key === 'Enter' && !running && onRun()}
          disabled={running}
        />
        <button
          onClick={onRun}
          disabled={running || !input.trim()}
          className="mt-2 px-4 py-1.5 text-sm rounded"
          style={{ background: 'var(--accent)', color: 'white', opacity: running || !input.trim() ? 0.5 : 1 }}
        >
          {running ? t('redis.command.running') : t('redis.command.run')}
        </button>
      </div>

      {/* Result */}
      {result !== null && (
        <div>
          <label className="block text-sm mb-2" style={{ color: 'var(--text-muted)' }}>
            Result
          </label>
          <pre className="p-3 text-sm font-mono rounded border overflow-auto max-h-96 whitespace-pre-wrap"
            style={{ background: 'var(--bg-surface)', borderColor: 'var(--border-subtle)', color: 'var(--text-default)' }}>
            {JSON.stringify(result, null, 2)}
          </pre>
        </div>
      )}

      {/* Readonly warning */}
      {isReadonly && (
        <div className="text-xs p-3 rounded"
          style={{ background: 'var(--warning)18', color: 'var(--warning)', border: '1px solid var(--warning)44' }}>
          {t('redis.command.readonly_blocked')}
        </div>
      )}
    </div>
  )
}
