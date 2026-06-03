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
    <div className="min-h-[100dvh] bg-[hsl(var(--workspace-bg))] text-[hsl(var(--workspace-text))] md:flex">
      <aside className="hidden w-16 shrink-0 border-r border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-rail))] md:flex md:flex-col">
        <div className="flex h-14 shrink-0 items-center justify-center border-b border-[hsl(var(--workspace-border))]">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg text-[hsl(var(--workspace-accent))]">
            <Icon name="book" className="h-5 w-5" />
          </div>
        </div>
        <nav className="flex-1 overflow-y-auto px-2 py-4">
          {groups.map(group => (
            <div key={group.label} className="mb-4 last:mb-0">
              <div className="mx-auto mb-2 h-px w-7 bg-[hsl(var(--workspace-border))]" />
              <div className="space-y-1">
                {group.items.map(item => (
                  <NavButton
                    key={item.page}
                    item={item}
                    active={page === item.page}
                    onClick={() => setPage(item.page)}
                  />
                ))}
              </div>
            </div>
          ))}
        </nav>
        <div className="border-t border-[hsl(var(--workspace-border))] p-2">
          <button
            type="button"
            onClick={logout}
            className="flex h-10 w-full items-center justify-center rounded-lg text-[hsl(var(--workspace-muted))] transition-colors hover:bg-[hsl(var(--workspace-panel-soft))] hover:text-[hsl(var(--workspace-text))]"
            aria-label={t.nav.logout}
            title={t.nav.logout}
          >
            <Icon name="logout" className="h-4 w-4" />
          </button>
        </div>
      </aside>

      <div className="flex min-h-[100dvh] min-w-0 flex-1 flex-col">
        <header className="shrink-0 border-b border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel))]">
          <div className="grid min-h-14 grid-cols-1 items-center gap-4 px-4 md:grid-cols-[minmax(220px,1fr)_minmax(220px,1fr)_minmax(220px,1fr)] md:px-5">
            <div className="flex min-w-0 items-center gap-5">
              <div className="min-w-[120px] text-sm font-semibold tracking-tight">Cloud Vault</div>
              <div className="hidden min-w-0 items-center gap-2 text-sm text-[hsl(var(--workspace-muted))] lg:flex">
                <span>工作区</span>
                <span>/</span>
                <span className="truncate text-[hsl(var(--workspace-text))]">{activeItem?.label ?? '文档'}</span>
                <Icon name="chevronDown" className="h-3.5 w-3.5" />
              </div>
            </div>
            <div className="hidden min-w-0 justify-center md:flex">
              <div className="flex min-w-0 items-center gap-2">
                <span className="truncate text-sm font-semibold">{page === 'vault' ? '设计系统更新计划' : activeItem?.label}</span>
                <span className="rounded-md border border-[hsl(var(--workspace-border))] bg-[hsl(var(--workspace-panel-soft))] px-2 py-0.5 text-xs text-[hsl(var(--workspace-muted))]">已保存</span>
              </div>
            </div>
            <div className="flex min-w-0 items-center justify-end gap-3">
              <div className="relative hidden w-48 lg:block">
                <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-[hsl(var(--workspace-muted))]" />
                <input className="workspace-search-input h-8 pl-8 text-xs" placeholder="搜索 (⌘K)" />
              </div>
              <button type="button" className="workspace-icon-button hidden md:inline-flex" aria-label="导出" title="导出">
                <Icon name="file" className="h-4 w-4" />
              </button>
              <button type="button" className="workspace-action-button hidden md:inline-flex">
                <Icon name="share" className="h-4 w-4" />
                分享
              </button>
              <button type="button" className="workspace-icon-button hidden md:inline-flex" aria-label="更多" title="更多">
                <Icon name="more" className="h-4 w-4" />
              </button>
              <Button variant="ghost" size="sm" icon="logout" className="md:hidden" onClick={logout}>
                {t.nav.logout}
              </Button>
            </div>
          </div>
          <nav className="flex gap-1 overflow-x-auto border-t border-[hsl(var(--workspace-border))] px-3 py-2 md:hidden">
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

        <main className="min-h-0 flex-1 overflow-hidden bg-background">
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

function NavButton({ item, active, onClick }: { item: NavItem; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={item.label}
      title={item.label}
      className={cn(
        'group relative flex h-10 w-full items-center justify-center rounded-lg text-[hsl(var(--workspace-muted))] transition-[background-color,color,transform] duration-150 active:translate-y-px',
        active ? 'bg-[hsl(var(--workspace-accent)/0.16)] text-[hsl(var(--workspace-accent))]' : 'hover:bg-[hsl(var(--workspace-panel-soft))] hover:text-[hsl(var(--workspace-text))]',
      )}
    >
      {active && <span className="absolute left-[-8px] h-5 w-0.5 rounded-full bg-[hsl(var(--workspace-accent))]" />}
      <span
        className={cn(
          'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border transition-colors',
          active ? 'border-[hsl(var(--workspace-accent)/0.28)] bg-[hsl(var(--workspace-accent)/0.10)]' : 'border-transparent',
        )}
      >
        <Icon name={item.icon} className="h-4 w-4" />
      </span>
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
