import { useI18n } from '../i18n'

export default function SettingsPage() {
  const { t } = useI18n()

  return (
    <div className="p-6 max-w-xl">
      <h1 className="text-xl font-bold mb-6" style={{ color: 'var(--text-strong)' }}>
        {t('settings.title')}
      </h1>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <h2 className="text-sm font-semibold mb-2" style={{ color: 'var(--text-default)' }}>
          {t('settings.sql_history')}
        </h2>
        <p className="text-xs" style={{ color: 'var(--text-muted)' }}>
          {t('settings.sql_history_hint')}
        </p>
      </section>
    </div>
  )
}
