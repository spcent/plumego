import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'
import { useTheme } from '../contexts/ThemeContext'

export default function AccountPage() {
  const { user, refreshUser } = useAuth()
  const { t, locale, setLocale } = useI18n()
  const { theme, setTheme } = useTheme()
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (user) {
      setDisplayName(user.display_name || '')
      setEmail(user.email || '')
    }
  }, [user])

  async function handleSave() {
    setSaving(true)
    setMessage('')
    try {
      const res = await fetch('/api/v1/auth/me', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          display_name: displayName,
          email,
          locale,
          theme,
        }),
      })
      if (!res.ok) throw new Error('Failed to update')
      await refreshUser()
      setMessage(t.common.success)
    } catch {
      setMessage(t.common.error)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">{t.nav.account}</h1>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">
            {t.auth.displayName}
          </label>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="w-full px-3 py-2 border rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">
            {t.auth.email}
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-3 py-2 border rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">
            {t.auth.locale}
          </label>
          <select
            value={locale}
            onChange={(e) => setLocale(e.target.value as 'en-US' | 'zh-CN' | 'zh-TW')}
            className="w-full px-3 py-2 border rounded"
          >
            <option value="en-US">{t.locales.enUS}</option>
            <option value="zh-CN">{t.locales.zhCN}</option>
            <option value="zh-TW">{t.locales.zhTW}</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">
            {t.auth.theme}
          </label>
          <select
            value={theme}
            onChange={(e) => setTheme(e.target.value as 'system' | 'light' | 'dark')}
            className="w-full px-3 py-2 border rounded"
          >
            <option value="system">{t.themes.system}</option>
            <option value="light">{t.themes.light}</option>
            <option value="dark">{t.themes.dark}</option>
          </select>
        </div>

        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {saving ? t.common.loading : t.common.save}
        </button>

        {message && (
          <div className="mt-4 p-3 rounded bg-green-100 text-green-800">
            {message}
          </div>
        )}
      </div>
    </div>
  )
}
