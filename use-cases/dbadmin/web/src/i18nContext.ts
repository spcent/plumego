import { createContext, useContext } from 'react'

export type Locale = 'en' | 'zh'

export interface I18nContextValue {
  lang: Locale
  setLang: (l: Locale) => void
  t: (key: string, vars?: Record<string, string | number>) => string
}

export const I18nContext = createContext<I18nContextValue>({
  lang: 'en',
  setLang: () => {},
  t: k => k,
})

export const useI18n = () => useContext(I18nContext)
