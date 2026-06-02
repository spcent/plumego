import { useState } from 'react'
import type { FormEvent } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'
import { Button, Field, Panel, StatusBanner, TextInput } from '../components/ui'
import { Icon } from '../components/icons'

export default function SetupPage() {
  const { setup } = useAuth()
  const { t } = useI18n()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (password !== confirmPassword) {
      setError(t.auth.passwordMismatch)
      return
    }

    if (password.length < 10) {
      setError(t.auth.passwordTooShort)
      return
    }

    setLoading(true)
    try {
      await setup({ username, email, password })
    } catch (err) {
      setError(err instanceof Error ? err.message : t.setup.setupFailed)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-[100dvh] items-center justify-center bg-background px-4 py-10">
      <div className="w-full max-w-md">
        <div className="mb-5 flex flex-col items-center text-center">
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg border border-primary/25 bg-primary/10 text-primary">
            <Icon name="book" className="h-6 w-6" />
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">{t.setup.title}</h1>
          <p className="mt-1 text-sm leading-5 text-muted-foreground">{t.setup.description}</p>
        </div>

        <Panel>
          <form className="space-y-4 p-5" onSubmit={handleSubmit}>
            <Field label={t.auth.username}>
              <TextInput
                id="username"
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
            </Field>

            <Field label={t.auth.email}>
              <TextInput
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={loading}
              />
            </Field>

            <Field label={t.auth.password} helper={t.auth.passwordTooShort}>
              <TextInput
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </Field>

            <Field label={t.auth.confirmPassword}>
              <TextInput
                id="confirmPassword"
                type="password"
                required
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={loading}
              />
            </Field>

          {error && (
            <StatusBanner tone="danger">{error}</StatusBanner>
          )}

            <Button type="submit" disabled={loading} variant="primary" className="w-full">
            {loading ? t.setup.creating : t.setup.createAccount}
            </Button>
          </form>
        </Panel>
      </div>
    </div>
  )
}
