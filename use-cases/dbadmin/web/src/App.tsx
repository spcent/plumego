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
            <Route path="conn/:connId/db/:dbName/table/:tableName/data" element={<DataPage />} />
            <Route path="conn/:connId/db/:dbName/table/:tableName/structure" element={<StructurePage />} />
            <Route path="conn/:connId/query" element={<QueryPage />} />
            <Route path="settings" element={<SettingsPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
      </ToastProvider>
      </I18nProvider>
    </ThemeProvider>
  )
}
