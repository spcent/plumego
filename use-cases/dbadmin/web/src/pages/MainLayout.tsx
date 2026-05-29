import { createContext, useContext, useEffect, useState } from 'react'
import { Outlet, Link, useNavigate } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useTheme } from '../ThemeContext'
import { useI18n } from '../i18n'
import ResourceExplorer from '../components/ResourceExplorer'

export const ConnectionsCtx = createContext<Connection[]>([])

export function useCurrentConn(connId: string | undefined): Connection | undefined {
  const conns = useContext(ConnectionsCtx)
  return conns.find(c => c.id === connId)
}

export default function MainLayout() {
  const [user, setUser] = useState<string | null>(null)
  const [connections, setConnections] = useState<Connection[]>([])
  const navigate = useNavigate()
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

  function refreshConnections() {
    api.listConnections().then(setConnections).catch(console.error)
  }

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--bg-app)', color: 'var(--text-default)' }}>
      {/* ── Sidebar ──────────────────────────────────────────── */}
      <aside
        className="flex flex-col shrink-0 overflow-hidden"
        style={{
          width: 260,
          background: 'var(--sb-bg)',
          borderRight: '1px solid var(--sb-border)',
        }}
      >
        {/* Sidebar header */}
        <div
          className="flex items-center justify-between px-3 h-11 shrink-0"
          style={{ borderBottom: '1px solid var(--sb-border)' }}
        >
          <span
            className="text-sm font-semibold tracking-tight select-none"
            style={{ color: 'var(--sb-text)' }}
          >
            DBAdmin
          </span>
          <div className="flex items-center gap-1.5">
            <button
              onClick={toggleTheme}
              className="w-6 h-6 flex items-center justify-center rounded text-[13px] hover:bg-white/10 transition-colors"
              style={{ color: 'var(--sb-muted)' }}
              title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
            >
              {theme === 'dark' ? '☀' : '☾'}
            </button>
            <button
              onClick={() => setLang(lang === 'en' ? 'zh' : 'en')}
              className="w-6 h-6 flex items-center justify-center rounded text-[11px] font-mono hover:bg-white/10 transition-colors"
              style={{ color: 'var(--sb-muted)' }}
              title="Switch language"
            >
              {lang === 'en' ? '中' : 'EN'}
            </button>
          </div>
        </div>

        {/* Quick links */}
        <div
          className="flex items-center gap-2 px-3 py-1.5 shrink-0"
          style={{ borderBottom: '1px solid var(--sb-border)' }}
        >
          <Link
            to="/connections"
            className="flex-1 text-[12px] text-center py-1 rounded transition-colors hover:bg-white/10"
            style={{ color: 'var(--sb-text)' }}
          >
            {t('nav.manage_connections')}
          </Link>
          <Link
            to="/settings"
            className="w-6 h-6 flex items-center justify-center rounded text-[13px] hover:bg-white/10 transition-colors"
            style={{ color: 'var(--sb-muted)' }}
            title={t('nav.settings')}
          >
            ⚙
          </Link>
        </div>

        {/* Resource Explorer — data-source-aware tree, replaces DatabaseTree */}
        <ResourceExplorer connections={connections} onRefresh={refreshConnections} />

        {/* Sidebar footer: logged-in user + logout */}
        <div
          className="flex items-center justify-between px-3 py-2 shrink-0"
          style={{
            borderTop: '1px solid var(--sb-border)',
            color: 'var(--sb-muted)',
            fontSize: 12,
          }}
        >
          <span className="truncate flex-1 min-w-0">{user ?? '…'}</span>
          <button
            onClick={handleLogout}
            className="ml-2 hover:text-white transition-colors"
            style={{ fontSize: 11 }}
          >
            {t('nav.logout')}
          </button>
        </div>
      </aside>

      {/* ── Workbench ─────────────────────────────────────────── */}
      <main
        className="flex-1 min-w-0 overflow-auto"
        style={{ background: 'var(--bg-app)' }}
      >
        <ConnectionsCtx.Provider value={connections}>
          <Outlet />
        </ConnectionsCtx.Provider>
      </main>
    </div>
  )
}
