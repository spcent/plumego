import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'

export default function SetupPage() {
  const { setup } = useAuth()
  const { t } = useI18n()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
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
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-bold text-foreground">
            {t.setup.title}
          </h2>
          <p className="mt-2 text-center text-sm text-muted-foreground">
            {t.setup.description}
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-foreground">
                {t.auth.username}
              </label>
              <input
                id="username"
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-ring focus:border-ring bg-background text-foreground"
                disabled={loading}
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-foreground">
                {t.auth.email}
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-ring focus:border-ring bg-background text-foreground"
                disabled={loading}
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-foreground">
                {t.auth.password}
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-ring focus:border-ring bg-background text-foreground"
                disabled={loading}
              />
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-foreground">
                {t.auth.confirmPassword}
              </label>
              <input
                id="confirmPassword"
                type="password"
                required
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-ring focus:border-ring bg-background text-foreground"
                disabled={loading}
              />
            </div>
          </div>

          {error && (
            <div className="text-sm text-destructive text-center">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? t.setup.creating : t.setup.createAccount}
          </button>
        </form>
      </div>
    </div>
  )
}
