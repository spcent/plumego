import { useState, useEffect, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { useCurrentConn } from './MainLayout'
import { api, type RedisKeyEntry, type RedisKeyDetail } from '../api'
import { useToast } from '../components/Toast'

type ViewMode = 'inspector' | 'command'

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

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.showToast('Copied', 'success')
  }

  const formatTTL = (ttl: number) => {
    if (ttl === -1) return t('redis.browser.ttl_none')
    if (ttl === -2) return 'N/A'
    return t('redis.browser.ttl_sec', { n: ttl })
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
            <div className="flex gap-2">
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
            <div className="mt-2 text-xs" style={{ color: 'var(--text-muted)' }}>
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
                onClick={() => loadKeyDetail(key.key)}
                className="px-3 py-2 border-b cursor-pointer hover:bg-opacity-50 transition-colors"
                style={{
                  borderColor: 'var(--border-subtle)',
                  background: selectedKey?.key === key.key ? 'var(--bg-muted)' : 'transparent',
                }}
              >
                <div className="flex items-start gap-2">
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-mono truncate" style={{ color: 'var(--text-strong)' }}>
                      {key.key}
                    </div>
                    <div className="flex gap-2 mt-1 text-xs" style={{ color: 'var(--text-muted)' }}>
                      <span className="px-1.5 py-0.5 rounded" style={{ background: 'var(--bg-surface)' }}>
                        {key.type}
                      </span>
                      <span>{formatTTL(key.ttl)}</span>
                    </div>
                  </div>
                  {!isReadonly && (
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
        >
          {t('redis.key.copy')}
        </button>
      </div>

      {/* TTL */}
      <div>
        <label className="block text-sm mb-1" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.ttl')}
        </label>
        <div className="flex gap-2">
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
        <div className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
          {detail.ttl === -1 ? t('redis.browser.ttl_none') : `${detail.ttl}s`}
        </div>
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
        >
          {t('redis.key.copy')}
        </button>
      </div>
    )
  }

  if (type === 'hash' && detail.hash) {
    const entries = Object.entries(detail.hash)
    return (
      <div>
        <div className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.hash_fields', { n: entries.length })}
        </div>
        <div className="border rounded overflow-hidden" style={{ borderColor: 'var(--border-subtle)' }}>
          {entries.map(([field, value]) => (
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
          ))}
        </div>
        <button
          onClick={() => onCopy(JSON.stringify(detail.hash, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
        >
          {t('redis.key.copy')}
        </button>
      </div>
    )
  }

  if (type === 'list' && detail.list) {
    return (
      <div>
        <div className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.list_length', { n: detail.list.length })}
        </div>
        <div className="border rounded overflow-hidden" style={{ borderColor: 'var(--border-subtle)' }}>
          {detail.list.map((item, i) => (
            <div key={i} className="flex border-b last:border-b-0" style={{ borderColor: 'var(--border-subtle)' }}>
              <div className="w-12 px-3 py-2 text-xs font-mono border-r"
                style={{ background: 'var(--bg-muted)', borderColor: 'var(--border-subtle)', color: 'var(--text-muted)' }}>
                {i}
              </div>
              <div className="flex-1 px-3 py-2 text-sm font-mono break-all"
                style={{ color: 'var(--text-default)' }}>
                {item}
              </div>
            </div>
          ))}
        </div>
        <button
          onClick={() => onCopy(JSON.stringify(detail.list, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
        >
          {t('redis.key.copy')}
        </button>
      </div>
    )
  }

  if (type === 'set' && detail.set) {
    return (
      <div>
        <div className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.set_size', { n: detail.set.length })}
        </div>
        <div className="flex flex-wrap gap-2">
          {detail.set.map((item, i) => (
            <span key={i} className="px-2 py-1 text-sm font-mono rounded"
              style={{ background: 'var(--bg-muted)', color: 'var(--text-strong)' }}>
              {item}
            </span>
          ))}
        </div>
        <button
          onClick={() => onCopy(JSON.stringify(detail.set, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
        >
          {t('redis.key.copy')}
        </button>
      </div>
    )
  }

  if (type === 'zset' && detail.zset) {
    return (
      <div>
        <div className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
          {t('redis.key.set_size', { n: detail.zset.length })}
        </div>
        <div className="border rounded overflow-hidden" style={{ borderColor: 'var(--border-subtle)' }}>
          {detail.zset.map((item, i) => (
            <div key={i} className="flex border-b last:border-b-0" style={{ borderColor: 'var(--border-subtle)' }}>
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
        <button
          onClick={() => onCopy(JSON.stringify(detail.zset, null, 2))}
          className="mt-2 text-xs"
          style={{ color: 'var(--accent)' }}
        >
          {t('redis.key.copy')}
        </button>
      </div>
    )
  }

  if (type === 'stream' && detail.stream) {
    return (
      <div className="space-y-2">
        <div className="text-sm" style={{ color: 'var(--text-default)' }}>
          {t('redis.key.stream_length', { n: detail.stream.length })}
        </div>
        <div className="text-sm" style={{ color: 'var(--text-default)' }}>
          {t('redis.key.stream_groups', { n: detail.stream.groups })}
        </div>
        <div className="text-xs p-3 rounded" style={{ background: 'var(--bg-muted)', color: 'var(--text-muted)' }}>
          Stream viewer coming soon. Use XREAD or XRANGE in the command console.
        </div>
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
