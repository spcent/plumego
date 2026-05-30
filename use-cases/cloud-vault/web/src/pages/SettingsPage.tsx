import { useState, useEffect } from 'react'
import { useI18n } from '../i18n/I18nContext'
import {
  getSettings,
  getVersion,
  getUpdateStatus,
  checkForUpdates,
  generateDiagnostics,
  listDiagnostics,
  getDiagnosticDownloadUrl,
  type SystemSettings,
  type VersionInfo,
  type UpdateStatus,
  type DiagnosticBundle,
} from '../api/system'

export default function SettingsPage() {
  const { t } = useI18n()
  const [settings, setSettings] = useState<SystemSettings | null>(null)
  const [version, setVersion] = useState<VersionInfo | null>(null)
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null)
  const [diagnostics, setDiagnostics] = useState<DiagnosticBundle[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [generatingDiagnostics, setGeneratingDiagnostics] = useState(false)

  useEffect(() => {
    loadSettings()
  }, [])

  async function loadSettings() {
    setLoading(true)
    setError('')
    try {
      const [s, v, u, d] = await Promise.all([
        getSettings(),
        getVersion(),
        getUpdateStatus(),
        listDiagnostics(),
      ])
      setSettings(s)
      setVersion(v)
      setUpdateStatus(u)
      setDiagnostics(d.bundles || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load settings')
    } finally {
      setLoading(false)
    }
  }

  async function handleCheckUpdate() {
    setCheckingUpdate(true)
    try {
      const status = await checkForUpdates()
      setUpdateStatus(status)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to check for updates')
    } finally {
      setCheckingUpdate(false)
    }
  }

  async function handleGenerateDiagnostics() {
    setGeneratingDiagnostics(true)
    try {
      const bundle = await generateDiagnostics()
      setDiagnostics([bundle, ...diagnostics])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate diagnostics')
    } finally {
      setGeneratingDiagnostics(false)
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
          <div className="space-y-6">
            {/* About Section */}
            <section className="border border-border rounded-lg p-4 bg-background space-y-3">
              <h3 className="text-sm font-medium text-foreground border-b border-border pb-2">
                About
              </h3>
              {version && (
                <>
                  <SettingRow label="Application" value="Markdown Cloud Vault" />
                  <SettingRow label="Version" value={version.version} />
                  <SettingRow label="Commit" value={version.commit} mono />
                  <SettingRow label="Build Time" value={version.build_time} mono />
                  <SettingRow label="Channel" value={version.channel} />
                  <SettingRow label="Runtime Mode" value="Desktop" />
                </>
              )}
            </section>

            {/* Configuration Section */}
            <section className="border border-border rounded-lg p-4 bg-background space-y-3">
              <h3 className="text-sm font-medium text-foreground border-b border-border pb-2">
                Configuration
              </h3>
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

            {/* Update Section */}
            <section className="border border-border rounded-lg p-4 bg-background space-y-3">
              <h3 className="text-sm font-medium text-foreground border-b border-border pb-2">
                Updates
              </h3>
              {updateStatus && (
                <>
                  <SettingRow label="Current Version" value={updateStatus.current_version} />
                  <SettingRow label="Latest Version" value={updateStatus.latest_version} />
                  <SettingRow
                    label="Update Available"
                    value={updateStatus.update_available ? 'Yes' : 'No'}
                    highlight={updateStatus.update_available ? 'warn' : 'ok'}
                  />
                  {updateStatus.checked_at && (
                    <SettingRow
                      label="Last Checked"
                      value={new Date(updateStatus.checked_at).toLocaleString()}
                    />
                  )}
                  {updateStatus.update_available && (
                    <div className="flex gap-2 pt-2">
                      {updateStatus.download_url && (
                        <a
                          href={updateStatus.download_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90"
                        >
                          Download Update
                        </a>
                      )}
                      {updateStatus.release_notes_url && (
                        <a
                          href={updateStatus.release_notes_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="px-3 py-1.5 text-xs border border-border rounded hover:bg-accent"
                        >
                          Release Notes
                        </a>
                      )}
                    </div>
                  )}
                </>
              )}
              <div className="flex justify-end pt-2">
                <button
                  onClick={handleCheckUpdate}
                  disabled={checkingUpdate}
                  className="px-3 py-1.5 text-xs border border-border rounded hover:bg-accent disabled:opacity-50"
                >
                  {checkingUpdate ? 'Checking...' : 'Check for Updates'}
                </button>
              </div>
            </section>

            {/* Diagnostics Section */}
            <section className="border border-border rounded-lg p-4 bg-background space-y-3">
              <h3 className="text-sm font-medium text-foreground border-b border-border pb-2">
                Diagnostics
              </h3>
              <p className="text-xs text-muted-foreground">
                Export diagnostic bundles to help troubleshoot issues. Bundles contain configuration, logs, and system information (with secrets redacted).
              </p>
              <div className="flex justify-end">
                <button
                  onClick={handleGenerateDiagnostics}
                  disabled={generatingDiagnostics}
                  className="px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50"
                >
                  {generatingDiagnostics ? 'Generating...' : 'Generate Diagnostic Bundle'}
                </button>
              </div>
              {diagnostics.length > 0 && (
                <div className="space-y-2 pt-2">
                  <h4 className="text-xs font-medium text-foreground">Available Bundles</h4>
                  {diagnostics.map((bundle) => (
                    <div
                      key={bundle.id}
                      className="flex items-center justify-between gap-2 text-xs border border-border rounded p-2"
                    >
                      <div className="flex-1 min-w-0">
                        <div className="font-mono truncate">{bundle.filename}</div>
                        <div className="text-muted-foreground">
                          {new Date(bundle.created_at).toLocaleString()} • {(bundle.size / 1024).toFixed(1)} KB
                        </div>
                      </div>
                      <a
                        href={getDiagnosticDownloadUrl(bundle.filename)}
                        className="px-2 py-1 border border-border rounded hover:bg-accent shrink-0"
                      >
                        Download
                      </a>
                    </div>
                  ))}
                </div>
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
