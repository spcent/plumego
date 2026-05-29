import { useState } from 'react'
import VaultPage from './pages/VaultPage'
import ImportPage from './pages/ImportPage'
import SearchPage from './pages/SearchPage'
import IndexPage from './pages/IndexPage'

type Page = 'vault' | 'search' | 'import' | 'index'

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
      <nav className="h-9 flex items-center gap-1 px-3 border-b border-border bg-background shrink-0">
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
      </nav>

      {/* Page content */}
      <div className="flex-1 overflow-hidden">
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
