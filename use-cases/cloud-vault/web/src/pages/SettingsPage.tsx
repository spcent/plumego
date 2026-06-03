import { useEffect, useState } from 'react'
import { useI18n } from '../i18n/I18nContext'
import {
  checkForUpdates,
  generateDiagnostics,
  getDiagnosticDownloadUrl,
  getSettings,
  getUpdateStatus,
  getVersion,
  listDiagnostics,
  type DiagnosticBundle,
  type SystemSettings,
  type UpdateStatus,
  type VersionInfo,
} from '../api/system'
import { Button, EmptyState, PageFrame, Panel, SkeletonRows, StatusBanner, cn } from '../components/ui'

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
    setError('')
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
    setError('')
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
    <PageFrame
      title={t.settings.title}
      description="Runtime configuration, version details, update checks, and diagnostic exports."
      action={<Button variant="secondary" size="sm" icon="refresh" onClick={loadSettings}>Refresh</Button>}
      width="wide"
    >
      {error && <StatusBanner tone="danger">{error}</StatusBanner>}
      {loading && <Panel><SkeletonRows count={6} /></Panel>}

      {!loading && settings && (
        <div className="grid gap-5 lg:grid-cols-2">
          <Panel title="About" description="Build identity and runtime channel.">
            <div className="divide-y divide-border">
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
            </div>
          </Panel>

          <Panel title="Configuration" description="Effective settings loaded by the app.">
            <div className="divide-y divide-border">
              <SettingRow label={t.settings.version} value={settings.version} />
              <SettingRow label={t.settings.storageProvider} value={settings.storage_provider === 'local' ? t.settings.local : t.settings.qiniu} />
              <SettingRow label={t.settings.authEnabled} value={settings.auth_enabled ? t.settings.enabled : t.settings.disabled} highlight={settings.auth_enabled ? 'ok' : 'warn'} />
              <SettingRow label={t.settings.searchEnabled} value={settings.search_enabled ? t.settings.enabled : t.settings.disabled} highlight={settings.search_enabled ? 'ok' : 'warn'} />
              <SettingRow label={t.settings.aiEnabled} value={settings.ai_enabled ? t.settings.enabled : t.settings.disabled} highlight={settings.ai_enabled ? 'ok' : 'warn'} />
              <SettingRow label={t.settings.databasePath} value={settings.database_path} mono />
              {settings.storage_root && <SettingRow label={t.settings.storageRoot} value={settings.storage_root} mono />}
            </div>
          </Panel>

          <Panel
            title="Updates"
            description="Check whether a newer application build is available."
            action={<Button size="sm" variant="secondary" onClick={handleCheckUpdate} disabled={checkingUpdate}>{checkingUpdate ? 'Checking...' : 'Check'}</Button>}
          >
            <div className="divide-y divide-border">
              {updateStatus && (
                <>
                  <SettingRow label="Status" value={updateStatus.check_enabled ? 'Enabled' : 'Disabled'} highlight={updateStatus.check_enabled ? 'ok' : 'warn'} />
                  <SettingRow label="Current Version" value={updateStatus.current_version.version} />
                  <SettingRow label="Latest Version" value={updateStatus.latest_release?.version || '-'} />
                  <SettingRow label="Update Available" value={updateStatus.update_available ? 'Yes' : 'No'} highlight={updateStatus.update_available ? 'warn' : 'ok'} />
                  {updateStatus.last_check && <SettingRow label="Last Checked" value={new Date(updateStatus.last_check).toLocaleString()} />}
                </>
              )}
            </div>
            {updateStatus?.update_available && (
              <div className="flex gap-2 border-t border-border p-3">
                {updateStatus.latest_release?.download_url && <AppLink href={updateStatus.latest_release.download_url}>Download Update</AppLink>}
                {updateStatus.latest_release?.release_url && <AppLink href={updateStatus.latest_release.release_url}>Release Notes</AppLink>}
              </div>
            )}
          </Panel>

          <Panel
            title="Diagnostics"
            description="Export redacted diagnostic bundles for troubleshooting."
            action={<Button size="sm" variant="primary" icon="archive" onClick={handleGenerateDiagnostics} disabled={generatingDiagnostics}>{generatingDiagnostics ? 'Generating...' : 'Generate'}</Button>}
          >
            <div className="p-4">
              {diagnostics.length === 0 ? (
                <EmptyState compact icon="archive" title="No diagnostic bundles" description="Generate a bundle when you need to inspect runtime state." />
              ) : (
                <div className="divide-y divide-border overflow-hidden rounded-lg border border-border">
                  {diagnostics.map(bundle => (
                    <div key={bundle.id} className="flex items-center justify-between gap-3 bg-background/45 px-3 py-3 text-xs">
                      <div className="min-w-0 flex-1">
                        <div className="truncate font-mono text-foreground">{bundle.filename}</div>
                        <div className="mt-0.5 text-muted-foreground">
                          {new Date(bundle.created_at).toLocaleString()} · {(bundle.size / 1024).toFixed(1)} KB
                        </div>
                      </div>
                      <a
                        href={getDiagnosticDownloadUrl(bundle.filename)}
                        className="inline-flex h-8 shrink-0 items-center rounded-md border border-border bg-surface px-2.5 font-medium text-foreground transition-colors hover:bg-accent"
                      >
                        Download
                      </a>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </Panel>
        </div>
      )}
    </PageFrame>
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
  return (
    <div className="flex items-center justify-between gap-4 px-4 py-2.5">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span
        className={cn(
          'max-w-[62%] break-all text-right text-sm text-foreground',
          mono && 'font-mono text-xs',
          highlight === 'ok' && 'text-emerald-600 dark:text-emerald-300',
          highlight === 'warn' && 'text-amber-600 dark:text-amber-300',
        )}
      >
        {value}
      </span>
    </div>
  )
}

function AppLink({ href, children }: { href: string; children: string }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex h-8 items-center rounded-md border border-border bg-surface px-2.5 text-xs font-medium text-foreground transition-colors hover:bg-accent"
    >
      {children}
    </a>
  )
}
