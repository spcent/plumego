import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import { useI18n } from '../i18nContext'
import { DatabaseIcon } from '../components/Icons'

export default function LoginPage() {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { t } = useI18n()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.login(username, password)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app-shell flex min-h-[100dvh] items-center justify-center px-4">
      <div className="panel w-full max-w-[380px] p-7">
        <div className="mb-6 flex flex-col items-center text-center">
          <div
            className="mb-3 grid h-12 w-12 place-items-center rounded-xl border"
            style={{ borderColor: 'var(--border-subtle)', background: 'var(--bg-muted)', color: 'var(--accent)' }}
          >
            <DatabaseIcon className="h-6 w-6" />
          </div>
          <h1 className="text-xl font-semibold" style={{ color: 'var(--text-default)' }}>DBAdmin</h1>
          <p className="mt-1 text-sm" style={{ color: 'var(--text-muted)' }}>
            Secure database operations console
          </p>
        </div>
        {error && (
          <div
            className="mb-4 rounded-md border px-3 py-2 text-sm"
            style={{ borderColor: 'var(--danger-border)', background: 'var(--danger-soft)', color: 'var(--danger)' }}
          >
            {error}
          </div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--text-muted)' }}>{t('login.username')}</label>
            <input
              type="text"
              value={username}
              onChange={e => setUsername(e.target.value)}
              className="input"
              autoComplete="username"
              required
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium" style={{ color: 'var(--text-muted)' }}>{t('login.password')}</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="input"
              autoComplete="current-password"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary w-full justify-center disabled:opacity-50"
          >
            {loading ? t('login.submitting') : t('login.submit')}
          </button>
        </form>
      </div>
    </div>
  )
}
