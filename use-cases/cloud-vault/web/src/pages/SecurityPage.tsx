import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n/I18nContext'

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
    } catch (err) {
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
    } catch (err) {
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
    } catch (err) {
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
    } catch (err) {
      setError(t.common.error || 'Failed to revoke all sessions')
    }
  }

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">{t.nav.security}</h1>

      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">{t.auth.changePassword}</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">
              {t.auth.currentPassword}
            </label>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              {t.auth.newPassword}
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              {t.auth.confirmPassword}
            </label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded"
            />
          </div>

          <button
            onClick={handleChangePassword}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            {t.auth.changePassword}
          </button>
        </div>
      </div>

      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">{t.pages.security.sessions}</h2>
          <button
            onClick={handleRevokeAll}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            {t.pages.security.revokeAll}
          </button>
        </div>

        <div className="space-y-2">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="p-4 border rounded flex items-center justify-between"
            >
              <div>
                <div className="font-medium">{session.user_agent}</div>
                <div className="text-sm text-muted-foreground">
                  {session.ip_address} · {new Date(session.created_at).toLocaleString()}
                </div>
              </div>
              <button
                onClick={() => handleRevokeSession(session.id)}
                className="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700"
              >
                {t.pages.security.revoke}
              </button>
            </div>
          ))}
        </div>
      </div>

      {message && (
        <div className="mt-4 p-3 rounded bg-green-100 text-green-800">{message}</div>
      )}
      {error && (
        <div className="mt-4 p-3 rounded bg-red-100 text-red-800">{error}</div>
      )}
    </div>
  )
}
