import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ToastProvider } from './components/Toast'
import { ThemeProvider } from './ThemeContext'
import { I18nProvider } from './i18n'

const LoginPage = lazy(() => import('./pages/LoginPage'))
const MainLayout = lazy(() => import('./pages/MainLayout'))
const ConnectionsPage = lazy(() => import('./pages/ConnectionsPage'))
const TablesPage = lazy(() => import('./pages/TablesPage'))
const DataPage = lazy(() => import('./pages/DataPage'))
const StructurePage = lazy(() => import('./pages/StructurePage'))
const QueryPage = lazy(() => import('./pages/QueryPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))
const RedisKeyPanel = lazy(() => import('./pages/RedisKeyPanel'))
const MongoCollectionPanel = lazy(() => import('./pages/MongoCollectionPanel'))
const ElasticsearchIndexPanel = lazy(() => import('./pages/ElasticsearchIndexPanel'))

function RouteFallback() {
  return <div className="route-fallback" aria-label="Loading page" />
}

export default function App() {
  return (
    <ThemeProvider>
      <I18nProvider>
      <ToastProvider>
      <BrowserRouter>
        <Suspense fallback={<RouteFallback />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<MainLayout />}>
              <Route index element={<Navigate to="/connections" replace />} />
              <Route path="connections" element={<ConnectionsPage />} />
              <Route path="conn/:connId/databases" element={<TablesPage />} />
              <Route path="conn/:connId/db/:dbName/tables" element={<TablesPage />} />
              <Route path="conn/:connId/db/:dbName/tables/:tableName/data" element={<DataPage />} />
              <Route path="conn/:connId/db/:dbName/tables/:tableName/structure" element={<StructurePage />} />
              <Route path="conn/:connId/query" element={<QueryPage />} />
              <Route path="conn/:connId/redis/:redisDb" element={<RedisKeyPanel />} />
              <Route path="conn/:connId/mongo/:mongoDb/collections" element={<MongoCollectionPanel />} />
              <Route path="conn/:connId/mongo/:mongoDb/:mongoColl/documents" element={<MongoCollectionPanel />} />
              <Route path="conn/:connId/es/:esIndex" element={<ElasticsearchIndexPanel />} />
              <Route path="conn/:connId/es/alias/:esAlias" element={<ElasticsearchIndexPanel />} />
              <Route path="conn/:connId/es/data-stream/:esDataStream" element={<ElasticsearchIndexPanel />} />
              <Route path="settings" element={<SettingsPage />} />
            </Route>
          </Routes>
        </Suspense>
      </BrowserRouter>
      </ToastProvider>
      </I18nProvider>
    </ThemeProvider>
  )
}
