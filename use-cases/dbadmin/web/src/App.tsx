import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ToastProvider } from './components/Toast'
import { ThemeProvider } from './ThemeContext'
import { I18nProvider } from './i18n'
import LoginPage from './pages/LoginPage'
import MainLayout from './pages/MainLayout'
import ConnectionsPage from './pages/ConnectionsPage'
import TablesPage from './pages/TablesPage'
import DataPage from './pages/DataPage'
import StructurePage from './pages/StructurePage'
import QueryPage from './pages/QueryPage'
import SettingsPage from './pages/SettingsPage'
import RedisKeyPanel from './pages/RedisKeyPanel'
import MongoCollectionPanel from './pages/MongoCollectionPanel'
import ElasticsearchIndexPanel from './pages/ElasticsearchIndexPanel'

export default function App() {
  return (
    <ThemeProvider>
      <I18nProvider>
      <ToastProvider>
      <BrowserRouter>
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
      </BrowserRouter>
      </ToastProvider>
      </I18nProvider>
    </ThemeProvider>
  )
}
