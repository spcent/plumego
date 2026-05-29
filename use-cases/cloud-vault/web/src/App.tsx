import { useState, lazy, Suspense } from 'react'

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

type Page = 'vault' | 'search' | 'import' | 'index' | 'duplicates' | 'collections' | 'review' | 'topics' | 'ai' | 'prompts' | 'system'

interface OpenDoc {
  id: string
  query: string
}

export default function App() {
  const [page, setPage] = useState<Page>('vault')
  const [openDoc, setOpenDoc] = useState<OpenDoc | null>(null)

  function handleOpenFromSearch(id: string, query: string) {
    setOpenDoc({ id, query })
    setPage('vault')
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      {/* Tab navigation */}
      <nav className="h-9 flex items-center gap-1 px-3 border-b border-border bg-background shrink-0 overflow-x-auto">
        <TabButton active={page === 'vault'} onClick={() => setPage('vault')}>
          ☁ Vault
        </TabButton>
        <TabButton active={page === 'search'} onClick={() => setPage('search')}>
          ⌕ Search
        </TabButton>
        <TabButton active={page === 'import'} onClick={() => setPage('import')}>
          ⇑ Import
        </TabButton>
        <TabButton active={page === 'index'} onClick={() => setPage('index')}>
          ⊙ Index
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'duplicates'} onClick={() => setPage('duplicates')}>
          ⋈ Duplicates
        </TabButton>
        <TabButton active={page === 'collections'} onClick={() => setPage('collections')}>
          ▤ Collections
        </TabButton>
        <TabButton active={page === 'topics'} onClick={() => setPage('topics')}>
          ◉ Topics
        </TabButton>
        <TabButton active={page === 'review'} onClick={() => setPage('review')}>
          ✓ Review
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'ai'} onClick={() => setPage('ai')}>
          ✦ AI Tasks
        </TabButton>
        <TabButton active={page === 'prompts'} onClick={() => setPage('prompts')}>
          ⊞ Prompts
        </TabButton>
        <span className="text-muted-foreground mx-1 text-xs">|</span>
        <TabButton active={page === 'system'} onClick={() => setPage('system')}>
          ⊕ System
        </TabButton>
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
