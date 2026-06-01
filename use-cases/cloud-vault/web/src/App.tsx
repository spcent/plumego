import { lazy, Suspense, useMemo, useState } from 'react'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import { ThemeProvider } from './contexts/ThemeContext'
import { I18nProvider, useI18n } from './i18n/I18nContext'
import LoginPage from './components/LoginPage'
import { Button, cn } from './components/ui'
import { Icon, type IconName } from './components/icons'

const VaultPage = lazy(() => import('./pages/VaultPage'))
const ImportPage = lazy(() => import('./pages/ImportPage'))
const SearchPage = lazy(() => import('./pages/SearchPage'))
const IndexPage = lazy(() => import('./pages/IndexPage'))
const DuplicatesPage = lazy(() => import('./pages/DuplicatesPage'))
const CollectionsPage = lazy(() => import('./pages/CollectionsPage'))
const ReviewPage = lazy(() => import('./pages/ReviewPage'))
const TopicsPage = lazy(() => import('./pages/TopicsPage'))
const AITasksPage = lazy(() => import('./pages/AITasksPage'))
const PromptsPage = lazy(() => import('./pages/PromptsPage'))
const SystemPage = lazy(() => import('./pages/SystemPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))
const AccountPage = lazy(() => import('./pages/AccountPage'))
const SecurityPage = lazy(() => import('./pages/SecurityPage'))
const SetupPage = lazy(() => import('./pages/SetupPage'))

type Page = 'vault' | 'search' | 'import' | 'index' | 'duplicates' | 'collections' | 'review' | 'topics' | 'ai' | 'prompts' | 'system' | 'settings' | 'account' | 'security'

interface OpenDoc {
  id: string
  query: string
}

interface NavItem {
  page: Page
  label: string
  icon: IconName
  description: string
}

interface NavGroup {
  label: string
  items: NavItem[]
}

function AppContent() {
  const { user, loading, setupRequired, logout } = useAuth()
  const { t } = useI18n()
  const [page, setPage] = useState<Page>('vault')
  const [openDoc, setOpenDoc] = useState<OpenDoc | null>(null)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true)

  const groups = useMemo<NavGroup[]>(() => [
    {
      label: 'Workspace',
      items: [
        { page: 'vault', label: t.nav.vault, icon: 'book', description: 'Write and review markdown notes' },
        { page: 'search', label: t.nav.search, icon: 'search', description: 'Find text, tags, and sources' },
        { page: 'import', label: t.nav.import, icon: 'import', description: 'Scan local markdown folders' },
        { page: 'index', label: t.nav.index, icon: 'database', description: 'Maintain full-text search' },
      ],
    },
    {
      label: 'Organize',
      items: [
        { page: 'duplicates', label: t.nav.duplicates, icon: 'copy', description: 'Compare similar documents' },
        { page: 'collections', label: t.nav.collections, icon: 'folder', description: 'Curate related material' },
        { page: 'topics', label: t.nav.topics, icon: 'grid', description: 'Inspect topic clusters' },
        { page: 'review', label: t.nav.review, icon: 'check', description: 'Process review queues' },
      ],
    },
    {
      label: 'Automation',
      items: [
        { page: 'ai', label: t.nav.aiTasks, icon: 'spark', description: 'Track AI background work' },
        { page: 'prompts', label: t.nav.prompts, icon: 'bolt', description: 'Manage prompt extracts' },
      ],
    },
    {
      label: 'Operations',
      items: [
        { page: 'system', label: t.nav.system, icon: 'box', description: 'Health, backup, diagnostics' },
        { page: 'settings', label: t.settings.title, icon: 'settings', description: 'Version and runtime settings' },
        { page: 'account', label: t.nav.account, icon: 'user', description: 'Profile and preferences' },
        { page: 'security', label: t.nav.security, icon: 'lock', description: 'Sessions and password' },
      ],
    },
  ], [t])

  const activeItem = groups.flatMap(group => group.items).find(item => item.page === page)

  if (loading) {
    return <FullPageLoading label={t.common.loading} />
  }

  if (setupRequired) {
    return (
      <Suspense fallback={<FullPageLoading label={t.common.loading} />}>
        <SetupPage />
      </Suspense>
    )
  }

  if (!user) {
    return <LoginPage onLoginSuccess={() => {}} />
  }

  function handleOpenFromSearch(id: string, query: string) {
    setOpenDoc({ id, query })
    setPage('vault')
  }

  return (
    <div className="min-h-[100dvh] bg-background text-foreground md:flex">
      <aside
        className={cn(
          'hidden shrink-0 border-r border-border bg-surface transition-[width] duration-200 md:flex md:flex-col',
          sidebarCollapsed ? 'w-16' : 'w-64',
        )}
      >
        <div className={cn('border-b border-border py-4', sidebarCollapsed ? 'px-3' : 'px-4')}>
          <div className={cn('flex items-center', sidebarCollapsed ? 'justify-center' : 'gap-3')}>
            <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-primary/20 bg-primary/10 text-primary">
              <Icon name="book" className="h-5 w-5" />
            </div>
            {!sidebarCollapsed && (
              <div className="min-w-0 flex-1">
                <div className="text-sm font-semibold tracking-tight">Cloud Vault</div>
                <div className="text-xs text-muted-foreground">Markdown workspace</div>
              </div>
            )}
            {!sidebarCollapsed && (
              <button
                type="button"
                onClick={() => setSidebarCollapsed(true)}
                className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                aria-label="Collapse sidebar"
                title="Collapse sidebar"
              >
                <Icon name="collapse" className="h-4 w-4" />
              </button>
            )}
          </div>
          {sidebarCollapsed && (
            <button
              type="button"
              onClick={() => setSidebarCollapsed(false)}
              className="mt-3 flex h-8 w-full items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              aria-label="Expand sidebar"
              title="Expand sidebar"
            >
              <Icon name="chevronRight" className="h-4 w-4" />
            </button>
          )}
        </div>
        <nav className={cn('flex-1 overflow-y-auto py-4', sidebarCollapsed ? 'px-2' : 'px-3')}>
          {groups.map(group => (
            <div key={group.label} className={cn('last:mb-0', sidebarCollapsed ? 'mb-3' : 'mb-5')}>
              {!sidebarCollapsed && (
                <div className="mb-2 px-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">
                  {group.label}
                </div>
              )}
              <div className="space-y-1">
                {group.items.map(item => (
                  <NavButton
                    key={item.page}
                    item={item}
                    active={page === item.page}
                    collapsed={sidebarCollapsed}
                    onClick={() => setPage(item.page)}
                  />
                ))}
              </div>
            </div>
          ))}
        </nav>
        <div className="border-t border-border p-3">
          {sidebarCollapsed ? (
            <button
              type="button"
              onClick={logout}
              className="flex h-9 w-full items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              aria-label={t.nav.logout}
              title={t.nav.logout}
            >
              <Icon name="logout" className="h-4 w-4" />
            </button>
          ) : (
            <Button variant="ghost" icon="logout" className="w-full justify-start" onClick={logout}>
              {t.nav.logout}
            </Button>
          )}
        </div>
      </aside>

      <div className="flex min-h-[100dvh] min-w-0 flex-1 flex-col">
        <header className="shrink-0 border-b border-border bg-surface/95">
          <div className="flex min-h-14 items-center gap-3 px-4 md:px-5">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
                {activeItem && <Icon name={activeItem.icon} className="h-4 w-4 text-primary" />}
                <span className="truncate">{activeItem?.label ?? 'Cloud Vault'}</span>
              </div>
              <div className="hidden truncate text-xs text-muted-foreground sm:block">
                {activeItem?.description ?? 'Markdown workspace'}
              </div>
            </div>
            <Button variant="ghost" size="sm" icon="logout" className="hidden md:inline-flex" onClick={logout}>
              {t.nav.logout}
            </Button>
          </div>
          <nav className="flex gap-1 overflow-x-auto border-t border-border px-3 py-2 md:hidden">
            {groups.flatMap(group => group.items).map(item => (
              <button
                key={item.page}
                type="button"
                onClick={() => setPage(item.page)}
                className={cn(
                  'inline-flex h-8 shrink-0 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium transition-colors',
                  page === item.page
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground',
                )}
              >
                <Icon name={item.icon} className="h-3.5 w-3.5" />
                {item.label}
              </button>
            ))}
          </nav>
        </header>

        <main className="min-h-0 flex-1 overflow-hidden">
          <Suspense fallback={<FullPageLoading label={t.common.loading} compact />}>
            {page === 'vault' && (
              <VaultPage
                initialDocId={openDoc?.id ?? null}
                highlightQuery={openDoc?.query ?? ''}
                onDocumentOpened={() => setOpenDoc(null)}
              />
            )}
            {page === 'search' && <SearchPage onOpenDocument={handleOpenFromSearch} />}
            {page === 'import' && <ImportPage />}
            {page === 'index' && <IndexPage />}
            {page === 'duplicates' && <DuplicatesPage />}
            {page === 'collections' && <CollectionsPage />}
            {page === 'topics' && <TopicsPage />}
            {page === 'review' && <ReviewPage />}
            {page === 'ai' && <AITasksPage />}
            {page === 'prompts' && <PromptsPage />}
            {page === 'system' && <SystemPage />}
            {page === 'settings' && <SettingsPage />}
            {page === 'account' && <AccountPage />}
            {page === 'security' && <SecurityPage />}
          </Suspense>
        </main>
      </div>
    </div>
  )
}

function NavButton({ item, active, collapsed, onClick }: { item: NavItem; active: boolean; collapsed: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={item.label}
      title={collapsed ? item.label : undefined}
      className={cn(
        'group flex w-full items-center rounded-md py-2 text-left transition-[background-color,color,transform] duration-150 active:translate-y-px',
        collapsed ? 'justify-center px-0' : 'gap-3 px-2',
        active ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-accent hover:text-foreground',
      )}
    >
      <span
        className={cn(
          'flex h-7 w-7 shrink-0 items-center justify-center rounded-md border transition-colors',
          active ? 'border-white/20 bg-white/15' : 'border-border bg-surface group-hover:bg-background',
        )}
      >
        <Icon name={item.icon} className="h-4 w-4" />
      </span>
      {!collapsed && (
        <span className="min-w-0">
          <span className="block truncate text-sm font-medium">{item.label}</span>
          <span className={cn('block truncate text-[11px]', active ? 'text-primary-foreground/75' : 'text-muted-foreground')}>
            {item.description}
          </span>
        </span>
      )}
    </button>
  )
}

function FullPageLoading({ label, compact }: { label: string; compact?: boolean }) {
  return (
    <div className={cn('flex items-center justify-center bg-background', compact ? 'h-full' : 'min-h-[100dvh]')}>
      <div className="flex items-center gap-3 rounded-lg border border-border bg-surface px-4 py-3 text-sm text-muted-foreground">
        <span className="h-2 w-2 animate-pulse rounded-full bg-primary" />
        {label}
      </div>
    </div>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <ThemeProvider>
        <I18nProvider>
          <AppContent />
        </I18nProvider>
      </ThemeProvider>
    </AuthProvider>
  )
}
