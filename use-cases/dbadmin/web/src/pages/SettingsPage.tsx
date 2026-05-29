import { useState } from 'react'
import { useI18n } from '../i18n'
import { sqlHistorySettings } from '../hooks/useSqlHistory'

export default function SettingsPage() {
  const { t } = useI18n()
  const [historyEnabled, setHistoryEnabled] = useState(() => sqlHistorySettings.get().sqlHistoryEnabled)

  function toggleHistory(val: boolean) {
    setHistoryEnabled(val)
    sqlHistorySettings.setEnabled(val)
  }

  return (
    <div className="p-6 max-w-xl">
      <h1 className="text-xl font-bold mb-6" style={{ color: 'var(--text-strong)' }}>
        {t('settings.title')}
      </h1>

      <section
        className="rounded-lg p-4"
        style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-subtle)' }}
      >
        <h2 className="text-sm font-semibold mb-3" style={{ color: 'var(--text-default)' }}>
          {t('settings.sql_history')}
        </h2>
        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="checkbox"
            className="mt-0.5 shrink-0"
            checked={historyEnabled}
            onChange={e => toggleHistory(e.target.checked)}
          />
          <div>
            <div className="text-sm" style={{ color: 'var(--text-default)' }}>
              {t('settings.sql_history_enabled')}
            </div>
            <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>
              {t('settings.sql_history_hint')}
            </p>
          </div>
        </label>
      </section>
    </div>
  )
}
