import { useState, FormEvent } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'
import { Button, Field, Panel, StatusBanner, TextInput } from './ui'
import { Icon } from './icons'

interface LoginPageProps {
  onLoginSuccess: () => void
}

export default function LoginPage({ onLoginSuccess }: LoginPageProps) {
  const { login } = useAuth()
  const { t } = useI18n()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await login({ username, password })
      onLoginSuccess()
    } catch (err: any) {
      setError(err.message || t.auth.loginFailed)
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
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">Cloud Vault</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t.auth.signIn}</p>
        </div>

        <Panel>
          <form className="space-y-4 p-5" onSubmit={handleSubmit}>
            <Field label={t.auth.username}>
              <TextInput
                id="username"
                name="username"
                type="text"
                required
                placeholder={t.auth.username}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
            </Field>
            <Field label={t.auth.password}>
              <TextInput
                id="password"
                name="password"
                type="password"
                required
                placeholder={t.auth.password}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </Field>

          {error && (
            <StatusBanner tone="danger">{error}</StatusBanner>
          )}

            <Button type="submit" disabled={loading} variant="primary" className="w-full">
              {loading ? t.auth.signingIn : t.auth.signIn}
            </Button>
          </form>
        </Panel>
      </div>
    </div>
  )
}
