import { useState, lazy, Suspense } from 'react'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import { ThemeProvider } from './contexts/ThemeContext'
import { I18nProvider, useI18n } from './i18n/I18nContext'
import LoginPage from './components/LoginPage'

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
const AccountPage = lazy(() => import('./pages/AccountPage'))
const SecurityPage = lazy(() => import('./pages/SecurityPage'))

type Page = 'vault' | 'search' | 'import' | 'index' | 'duplicates' | 'collections' | 'review' | 'topics' | 'ai' | 'prompts' | 'system' | 'account' | 'security'

interface OpenDoc {
  id: string
  query: string
}

function AppContent() {
  const { user, loading, logout } = useAuth()
  const { t } = useI18n()
  const [page, setPage] = useState<Page>('vault')
  const [openDoc, setOpenDoc] = useState<OpenDoc | null>(null)

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">{t.common.loading}</div>
      </div>
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
    <div className="h-screen flex flex-col overflow-hidden">
      {/* Tab navigation */}
      <nav className="h-9 flex items-center gap-1 px-3 border-b border-border bg-background shrink-0 overflow-x-auto">
        <TabButton active={page === 'vault'} onClick={() => setPage('vault')}>
          ☁ {t.nav.vault}
        </TabButton>
        <TabButton active={page === 'search'} onClick={() => setPage('search')}>
          ⌕ {t.nav.search}
        </TabButton>
        <TabButton active={page === 'import'} onClick={() => setPage('import')}>
          ⇑ {t.nav.import}
        </TabButton>
        <TabButton active={page === 'index'} onClick={() => setPage('index')}>
          ⊙ {t.nav.index}
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'duplicates'} onClick={() => setPage('duplicates')}>
          ⋈ {t.nav.duplicates}
        </TabButton>
        <TabButton active={page === 'collections'} onClick={() => setPage('collections')}>
          ▤ {t.nav.collections}
        </TabButton>
        <TabButton active={page === 'topics'} onClick={() => setPage('topics')}>
          ◉ {t.nav.topics}
        </TabButton>
        <TabButton active={page === 'review'} onClick={() => setPage('review')}>
          ✓ {t.nav.review}
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'ai'} onClick={() => setPage('ai')}>
          ✦ {t.nav.aiTasks}
        </TabButton>
        <TabButton active={page === 'prompts'} onClick={() => setPage('prompts')}>
          ⊞ {t.nav.prompts}
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'system'} onClick={() => setPage('system')}>
          ⊕ {t.nav.system}
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'account'} onClick={() => setPage('account')}>
          {t.nav.account}
        </TabButton>
        <TabButton active={page === 'security'} onClick={() => setPage('security')}>
          {t.nav.security}
        </TabButton>
        <button
          onClick={logout}
          className="px-3 py-1 text-sm rounded text-muted-foreground hover:text-foreground hover:bg-accent transition-colors ml-auto"
        >
          {t.nav.logout}
        </button>
      </nav>

      {/* Page content */}
      <div className="flex-1 overflow-hidden">
        <Suspense fallback={<div className="flex items-center justify-center h-full text-muted-foreground text-sm">Loading…</div>}>
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
        </Suspense>
      </div>
    </div>
  )
}

function TabButton({
  active, onClick, children,
}: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1 text-sm rounded transition-colors ${
        active
          ? 'bg-primary text-primary-foreground font-medium'
          : 'text-muted-foreground hover:text-foreground hover:bg-accent'
      }`}
    >
      {children}
    </button>
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
