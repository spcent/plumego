import { useState, useEffect } from 'react'
import { useI18n } from '../i18n/I18nContext'
import { getSettings, type SystemSettings } from '../api/system'

export default function SettingsPage() {
  const { t } = useI18n()
  const [settings, setSettings] = useState<SystemSettings | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    loadSettings()
  }, [])

  async function loadSettings() {
    setLoading(true)
    setError('')
    try {
      const s = await getSettings()
      setSettings(s)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load settings')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="max-w-2xl mx-auto px-6 py-6 space-y-6">
        <div>
          <h2 className="text-base font-semibold text-foreground">{t.settings.title}</h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {t.settings.version}: {settings?.version ?? '—'}
          </p>
        </div>

        {error && (
          <div className="text-sm text-destructive border border-destructive/30 rounded px-3 py-2">
            {error}
          </div>
        )}

        {loading && (
          <div className="text-sm text-muted-foreground">{t.common.loading}</div>
        )}

        {!loading && settings && (
          <div className="space-y-4">
            <section className="border border-border rounded-lg p-4 bg-background space-y-3">
              <SettingRow label={t.settings.version} value={settings.version} />
              <SettingRow
                label={t.settings.storageProvider}
                value={settings.storage_provider === 'local' ? t.settings.local : t.settings.qiniu}
              />
              <SettingRow
                label={t.settings.authEnabled}
                value={settings.auth_enabled ? t.settings.enabled : t.settings.disabled}
                highlight={settings.auth_enabled ? 'ok' : 'warn'}
              />
              <SettingRow
                label={t.settings.searchEnabled}
                value={settings.search_enabled ? t.settings.enabled : t.settings.disabled}
              />
              <SettingRow
                label={t.settings.aiEnabled}
                value={settings.ai_enabled ? t.settings.enabled : t.settings.disabled}
              />
              <SettingRow label={t.settings.databasePath} value={settings.database_path} mono />
              {settings.storage_root && (
                <SettingRow label={t.settings.storageRoot} value={settings.storage_root} mono />
              )}
            </section>

            <div className="flex justify-end">
              <button
                onClick={loadSettings}
                className="px-4 py-2 text-sm border border-border rounded hover:bg-accent"
              >
                {t.common.search}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function SettingRow({
  label,
  value,
  mono,
  highlight,
}: {
  label: string
  value: string
  mono?: boolean
  highlight?: 'ok' | 'warn'
}) {
  const valueClass = mono ? 'font-mono text-xs' : 'text-sm'
  const highlightClass =
    highlight === 'ok'
      ? 'text-green-600'
      : highlight === 'warn'
      ? 'text-yellow-600'
      : 'text-foreground'
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className={`${valueClass} ${highlightClass} text-right break-all`}>{value}</span>
    </div>
  )
}
