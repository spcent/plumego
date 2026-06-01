import { useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'
import { useTheme } from '../contexts/ThemeContext'
import { Button, SelectInput, TextInput, cn } from '../components/ui'

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
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto max-w-3xl px-5 py-8 md:px-8">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">{t.nav.account}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t.pages.account.preferences}</p>
        </div>

        <div className="rounded-lg border border-border bg-surface p-5 shadow-sm shadow-slate-950/5 dark:shadow-none">
          <div className="space-y-5">
            <Field label={t.auth.displayName}>
              <TextInput
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </Field>

            <Field label={t.auth.email}>
              <TextInput
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </Field>

            <Field label={t.auth.locale}>
              <SelectInput
                value={locale}
                onChange={(e) => setLocale(e.target.value as 'en-US' | 'zh-CN' | 'zh-TW')}
                className="w-full"
              >
                <option value="en-US">{t.locales.enUS}</option>
                <option value="zh-CN">{t.locales.zhCN}</option>
                <option value="zh-TW">{t.locales.zhTW}</option>
              </SelectInput>
            </Field>

            <Field label={t.auth.theme}>
              <SelectInput
                value={theme}
                onChange={(e) => setTheme(e.target.value as 'system' | 'light' | 'dark')}
                className="w-full"
              >
                <option value="system">{t.themes.system}</option>
                <option value="light">{t.themes.light}</option>
                <option value="dark">{t.themes.dark}</option>
              </SelectInput>
            </Field>

            <div className="flex items-center gap-3 pt-2">
              <Button
                onClick={handleSave}
                disabled={saving}
                variant="primary"
                icon="save"
              >
                {saving ? t.common.loading : t.common.save}
              </Button>

              {message && (
                <div
                  className={cn(
                    'rounded-md border px-3 py-2 text-sm',
                    message === t.common.success
                      ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                      : 'border-destructive/25 bg-destructive/10 text-destructive',
                  )}
                >
                  {message}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-medium text-foreground">{label}</span>
      {children}
    </label>
  )
}
