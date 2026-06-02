import { useEffect, useState } from 'react'
import { Outlet, Link, useNavigate } from 'react-router-dom'
import { api, type Connection } from '../api'
import { useTheme } from '../theme'
import { useI18n } from '../i18nContext'
import ResourceExplorer from '../components/ResourceExplorer'
import { DatabaseIcon, LogoutIcon, MoonIcon, SettingsIcon, SunIcon } from '../components/Icons'
import { ConnectionsCtx } from '../context/connections'

export default function MainLayout() {
  const [user, setUser] = useState<string | null>(null)
  const [role, setRole] = useState<string | null>(null)
  const [connections, setConnections] = useState<Connection[]>([])
  const navigate = useNavigate()
  const { theme, toggleTheme } = useTheme()
  const { lang, setLang, t } = useI18n()

  useEffect(() => {
    let active = true

    api.me()
      .then(async r => {
        if (!active) return
        setUser(r.user)
        setRole(r.role)
        const conns = await api.listConnections()
        if (active) setConnections(conns)
      })
      .catch(() => {
        if (active) navigate('/login')
      })

    return () => { active = false }
  }, [navigate])

  async function handleLogout() {
    await api.logout().catch(() => {})
    navigate('/login')
  }

  function refreshConnections() {
    api.listConnections().then(setConnections).catch(console.error)
  }

  return (
    <div className="app-shell flex h-[100dvh] overflow-hidden">
      {/* ── Sidebar ──────────────────────────────────────────── */}
      <aside
        className="flex flex-col shrink-0 overflow-hidden"
        style={{
          width: 272,
          background: 'var(--sb-bg)',
          borderRight: '1px solid var(--sb-border)',
          boxShadow: '10px 0 32px -28px rgba(0,0,0,.9)',
        }}
      >
        {/* Sidebar header */}
        <div
          className="flex items-center justify-between px-3 h-14 shrink-0"
          style={{ borderBottom: '1px solid var(--sb-border)' }}
        >
          <div className="flex items-center gap-2 min-w-0">
            <span className="grid h-8 w-8 place-items-center rounded-lg" style={{ background: 'var(--sb-surface)', color: '#8bb8ff' }}>
              <DatabaseIcon className="h-4 w-4" />
            </span>
            <div className="min-w-0">
              <div className="text-sm font-semibold tracking-tight select-none" style={{ color: 'var(--sb-text)' }}>
                DBAdmin
              </div>
              <div className="text-[11px] leading-none" style={{ color: 'var(--sb-muted)' }}>
                Data workbench
              </div>
            </div>
          </div>
          <div className="flex items-center gap-1.5">
            <button
              onClick={toggleTheme}
              className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-white/10 transition-colors"
              style={{ color: 'var(--sb-muted)' }}
              title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
            >
              {theme === 'dark' ? <SunIcon className="h-3.5 w-3.5" /> : <MoonIcon className="h-3.5 w-3.5" />}
            </button>
            <button
              onClick={() => setLang(lang === 'en' ? 'zh' : 'en')}
              className="w-7 h-7 flex items-center justify-center rounded-md text-[11px] font-mono hover:bg-white/10 transition-colors"
              style={{ color: 'var(--sb-muted)' }}
              title="Switch language"
            >
              {lang === 'en' ? '中' : 'EN'}
            </button>
          </div>
        </div>

        {/* Quick links */}
        <div
          className="flex items-center gap-2 px-3 py-2.5 shrink-0"
          style={{ borderBottom: '1px solid var(--sb-border)' }}
        >
          <Link
            to="/connections"
            className="flex-1 text-[12px] text-center py-1.5 rounded-md transition-colors hover:bg-white/10"
            style={{ color: 'var(--sb-text)' }}
          >
            {t('nav.manage_connections')}
          </Link>
          <Link
            to="/settings"
            className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-white/10 transition-colors"
            style={{ color: 'var(--sb-muted)' }}
            title={t('nav.settings')}
          >
            <SettingsIcon className="h-4 w-4" />
          </Link>
        </div>

        {/* Resource Explorer — data-source-aware tree, replaces DatabaseTree */}
        <ResourceExplorer connections={connections} onRefresh={refreshConnections} />

        {/* Sidebar footer: logged-in user + logout */}
        <div
          className="flex items-center justify-between gap-2 px-3 py-3 shrink-0"
          style={{
            borderTop: '1px solid var(--sb-border)',
            color: 'var(--sb-muted)',
            fontSize: 12,
          }}
        >
          <div className="truncate flex-1 min-w-0">
            <div className="truncate" style={{ color: 'var(--sb-text)' }}>{user ?? '...'}</div>
            <div className="uppercase tracking-wide text-[10px]">{role ?? 'session'}</div>
          </div>
          <button
            onClick={handleLogout}
            className="h-8 w-8 grid place-items-center rounded-md hover:bg-white/10 hover:text-white transition-colors"
            title={t('nav.logout')}
          >
            <LogoutIcon className="h-4 w-4" />
          </button>
        </div>
      </aside>

      {/* ── Workbench ─────────────────────────────────────────── */}
      <main
        className="flex-1 min-w-0 overflow-auto"
      >
        <ConnectionsCtx.Provider value={connections}>
          <Outlet />
        </ConnectionsCtx.Provider>
      </main>
    </div>
  )
}
