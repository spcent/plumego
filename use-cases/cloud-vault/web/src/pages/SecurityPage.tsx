import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'
import { Button, EmptyState, TextInput, cn } from '../components/ui'

interface Session {
  id: string
  user_agent: string
  ip_address: string
  created_at: string
  expires_at: string
}

export default function SecurityPage() {
  const { logout } = useAuth()
  const { t } = useI18n()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [sessions, setSessions] = useState<Session[]>([])
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    fetchSessions()
  }, [])

  async function fetchSessions() {
    try {
      const res = await fetch('/api/v1/auth/sessions')
      if (!res.ok) throw new Error('Failed to fetch sessions')
      const data = await res.json()
      setSessions(data.data?.sessions || [])
    } catch {
      setError(t.common.error || 'Failed to load sessions')
    }
  }

  async function handleChangePassword() {
    setMessage('')
    setError('')

    if (newPassword !== confirmPassword) {
      setError(t.auth.passwordMismatch || 'Passwords do not match')
      return
    }

    if (newPassword.length < 10) {
      setError(t.auth.passwordTooShort || 'Password must be at least 10 characters')
      return
    }

    try {
      const res = await fetch('/api/v1/auth/change-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      })
      if (!res.ok) throw new Error('Failed to change password')
      setMessage(t.common.success || 'Password changed successfully')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      await logout()
    } catch {
      setError(t.auth.invalidCurrentPassword || 'Invalid current password')
    }
  }

  async function handleRevokeSession(sessionId: string) {
    try {
      const res = await fetch(`/api/v1/auth/sessions/${sessionId}`, {
        method: 'DELETE',
      })
      if (!res.ok) throw new Error('Failed to revoke session')
      await fetchSessions()
    } catch {
      setError(t.common.error || 'Failed to revoke session')
    }
  }

  async function handleRevokeAll() {
    try {
      const res = await fetch('/api/v1/auth/sessions/revoke-all', {
        method: 'POST',
      })
      if (!res.ok) throw new Error('Failed to revoke all sessions')
      await logout()
    } catch {
      setError(t.common.error || 'Failed to revoke all sessions')
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto max-w-4xl px-5 py-8 md:px-8">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">{t.nav.security}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t.pages.security.sessions}</p>
        </div>

        <div className="space-y-6">
          <section className="rounded-lg border border-border bg-surface p-5 shadow-sm shadow-slate-950/5 dark:shadow-none">
            <div className="mb-5">
              <h2 className="text-lg font-semibold text-foreground">{t.auth.changePassword}</h2>
              <p className="mt-1 text-sm text-muted-foreground">{t.auth.passwordTooShort}</p>
            </div>

            <div className="space-y-5">
              <Field label={t.auth.currentPassword}>
                <TextInput
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                />
              </Field>

              <Field label={t.auth.newPassword}>
                <TextInput
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                />
              </Field>

              <Field label={t.auth.confirmPassword}>
                <TextInput
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                />
              </Field>

              <Button onClick={handleChangePassword} variant="primary" icon="lock">
                {t.auth.changePassword}
              </Button>
            </div>
          </section>

          <section className="rounded-lg border border-border bg-surface p-5 shadow-sm shadow-slate-950/5 dark:shadow-none">
            <div className="mb-5 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-foreground">{t.pages.security.sessions}</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  {sessions.length.toLocaleString()} {t.pages.security.activeSessions}
                </p>
              </div>
              <Button onClick={handleRevokeAll} variant="danger" icon="trash">
                {t.pages.security.revokeAll}
              </Button>
            </div>

            {sessions.length === 0 ? (
              <EmptyState compact icon="lock" title={t.pages.security.sessions} description={t.pages.security.noSessions} />
            ) : (
              <div className="divide-y divide-border rounded-lg border border-border">
                {sessions.map((session) => (
                  <div
                    key={session.id}
                    className="flex items-center justify-between gap-4 bg-background/40 px-4 py-3"
                  >
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium text-foreground">{session.user_agent}</div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {session.ip_address} · {new Date(session.created_at).toLocaleString()}
                      </div>
                    </div>
                    <Button
                      onClick={() => handleRevokeSession(session.id)}
                      variant="danger"
                      size="sm"
                    >
                      {t.pages.security.revoke}
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {message && (
            <StatusMessage tone="success">{message}</StatusMessage>
          )}
          {error && (
            <StatusMessage tone="error">{error}</StatusMessage>
          )}
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

function StatusMessage({ tone, children }: { tone: 'success' | 'error'; children: ReactNode }) {
  return (
    <div
      className={cn(
        'rounded-md border px-3 py-2 text-sm',
        tone === 'success'
          ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
          : 'border-destructive/25 bg-destructive/10 text-destructive',
      )}
    >
      {children}
    </div>
  )
}
