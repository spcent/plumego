import { useState } from 'react'
import { useI18n } from '../i18n'
import { sqlHistorySettings } from '../hooks/useSqlHistory'

export default function SettingsPage() {
  const { t } = useI18n()
  const [historyEnabled, setHistoryEnabled] = useState(() => sqlHistorySettings.get().sqlHistoryEnabled)

  function handleToggleHistory(val: boolean) {
    setHistoryEnabled(val)
    sqlHistorySettings.setEnabled(val)
  }

  return (
    <div className="p-6 max-w-xl">
      <h1 className="text-xl font-bold text-gray-800 dark:text-gray-100 mb-6">{t('settings.title')}</h1>

      <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">{t('settings.sql_history')}</h2>
        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={historyEnabled}
            onChange={e => handleToggleHistory(e.target.checked)}
            className="mt-0.5"
          />
          <div>
            <div className="text-sm text-gray-800 dark:text-gray-200">{t('settings.sql_history_enabled')}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('settings.sql_history_hint')}</div>
          </div>
        </label>
      </section>
    </div>
  )
}
