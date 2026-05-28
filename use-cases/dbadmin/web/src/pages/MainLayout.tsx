import { createContext, useContext, useEffect, useState } from 'react'
import { Outlet, Link, useNavigate, useParams } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useTheme } from '../ThemeContext'
import { useI18n } from '../i18n'

export const ConnectionsCtx = createContext<Connection[]>([])

export function useCurrentConn(connId: string | undefined): Connection | undefined {
  const conns = useContext(ConnectionsCtx)
  return conns.find(c => c.id === connId)
}

export default function MainLayout() {
  const [user, setUser] = useState<string | null>(null)
  const [connections, setConnections] = useState<Connection[]>([])
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})
  const [databases, setDatabases] = useState<Record<string, string[]>>({})
  const navigate = useNavigate()
  const params = useParams()
  const { theme, toggleTheme } = useTheme()
  const { lang, setLang, t } = useI18n()

  useEffect(() => {
    api.me().then(r => setUser(r.user)).catch(() => navigate('/login'))
    api.listConnections().then(setConnections).catch(console.error)
  }, [])

  async function handleLogout() {
    await api.logout().catch(() => {})
    navigate('/login')
  }

  async function toggleConn(connId: string) {
    const next = !expanded[connId]
    setExpanded(p => ({ ...p, [connId]: next }))
    if (next && !databases[connId]) {
      try {
        const dbs = await api.databases(connId)
        setDatabases(p => ({ ...p, [connId]: dbs }))
      } catch {}
    }
  }

  return (
    <div className="flex h-screen bg-gray-100 dark:bg-gray-950 text-sm">
      {/* Sidebar */}
      <aside className="w-56 bg-gray-800 dark:bg-gray-900 text-gray-200 flex flex-col shrink-0 overflow-x-hidden border-r border-gray-700">
        <div className="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
          <span className="font-bold text-white">DBAdmin</span>
          <div className="flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="text-xs text-gray-400 hover:text-white w-5 text-center"
              title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
            >
              {theme === 'dark' ? '☀' : '☾'}
            </button>
            <button
              onClick={() => setLang(lang === 'en' ? 'zh' : 'en')}
              className="text-xs text-gray-400 hover:text-white font-mono"
              title="Switch language"
            >
              {lang === 'en' ? '中' : 'EN'}
            </button>
            <button onClick={handleLogout} className="text-xs text-gray-400 hover:text-white">
              {t('nav.logout')}
            </button>
          </div>
        </div>
        <div className="px-4 py-2 border-b border-gray-700 flex items-center justify-between">
          <Link to="/connections" className="text-xs text-blue-400 hover:text-blue-300">
            {t('nav.manage_connections')}
          </Link>
          <Link to="/settings" className="text-xs text-gray-400 hover:text-gray-200">
            {t('nav.settings')}
          </Link>
        </div>
        <nav className="flex-1 overflow-y-auto py-1">
          {connections.map(conn => (
            <div key={conn.id}>
              <button
                onClick={() => toggleConn(conn.id)}
                className="w-full text-left px-4 py-2 hover:bg-gray-700 flex items-center gap-2"
              >
                <span className="text-xs font-mono bg-gray-600 px-1 rounded shrink-0">
                  {conn.driver === 'mysql' ? 'MY' : 'SQ'}
                </span>
                <span className="truncate flex-1 min-w-0">{conn.name}</span>
                {conn.readonly && (
                  <span className="text-xs bg-amber-600/70 text-amber-100 px-1 rounded shrink-0 font-mono">RO</span>
                )}
                <span className="shrink-0 text-gray-500">{expanded[conn.id] ? '▾' : '▸'}</span>
              </button>
              {expanded[conn.id] && (
                <div className="pl-2">
                  {(databases[conn.id] || []).map(db => {
                    const isActive = params.connId === conn.id && params.dbName === db
                    return (
                      <Link
                        key={db}
                        to={`/conn/${conn.id}/db/${db}/tables`}
                        className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded mx-1 my-0.5 truncate ${
                          isActive
                            ? 'bg-blue-600 text-white'
                            : 'text-gray-300 hover:bg-gray-700'
                        }`}
                      >
                        <span className="shrink-0 text-gray-400">🗄</span>
                        <span className="truncate">{db}</span>
                      </Link>
                    )
                  })}
                  <Link
                    to={`/conn/${conn.id}/query`}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-blue-400 hover:text-blue-300 mx-1"
                  >
                    <span className="shrink-0">▶</span>
                    <span>{t('nav.sql_console')}</span>
                  </Link>
                </div>
              )}
            </div>
          ))}
        </nav>
        {user && (
          <div className="px-4 py-2 border-t border-gray-700 text-xs text-gray-400 truncate">
            {user}
          </div>
        )}
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto bg-white dark:bg-gray-900">
        <ConnectionsCtx.Provider value={connections}>
          <Outlet />
        </ConnectionsCtx.Provider>
      </main>
    </div>
  )
}
