import React, { useEffect, useRef, useState } from 'react'
import { Sun, Moon, RefreshCw } from 'lucide-react'

const REFRESH_INTERVALS = [
  { value: '10', label: '10s' },
  { value: '30', label: '30s' },
  { value: '60', label: '1m' },
  { value: '120', label: '2m' },
  { value: '300', label: '5m' },
  { value: '600', label: '10m' },
]
const DEFAULT_REFRESH_INTERVAL = '300'
const THEME_COOKIE_NAME = 'theme'
const THEME_COOKIE_MAX_AGE = 31536000
const STORAGE_KEYS = {
  REFRESH_INTERVAL: 'gatus:refresh-interval',
}

function wantsDarkMode() {
  const themeFromCookie = document.cookie.match(
    new RegExp(`${THEME_COOKIE_NAME}=(dark|light);?`)
  )?.[1]
  return (
    themeFromCookie === 'dark' ||
    (!themeFromCookie &&
      (window.matchMedia('(prefers-color-scheme: dark)').matches ||
        document.documentElement.classList.contains('dark')))
  )
}

function getStoredRefreshInterval() {
  const stored = localStorage.getItem(STORAGE_KEYS.REFRESH_INTERVAL)
  const parsedValue = stored && parseInt(stored)
  const isValid = parsedValue && parsedValue >= 10 && REFRESH_INTERVALS.some((i) => i.value === stored)
  return isValid ? stored : DEFAULT_REFRESH_INTERVAL
}

function formatRefreshInterval(value) {
  const interval = REFRESH_INTERVALS.find((i) => i.value === value)
  return interval ? interval.label : `${value}s`
}

export default function Settings({ onRefreshData }) {
  const [refreshIntervalValue, setRefreshIntervalValue] = useState(getStoredRefreshInterval())
  const [darkMode, setDarkMode] = useState(wantsDarkMode())
  const [showRefreshMenu, setShowRefreshMenu] = useState(false)
  const refreshIntervalHandler = useRef(null)
  const containerRef = useRef(null)
  const onRefreshDataRef = useRef(onRefreshData)
  onRefreshDataRef.current = onRefreshData

  const setRefreshInterval = (seconds) => {
    localStorage.setItem(STORAGE_KEYS.REFRESH_INTERVAL, seconds)
    if (refreshIntervalHandler.current) {
      clearInterval(refreshIntervalHandler.current)
    }
    refreshIntervalHandler.current = setInterval(() => {
      onRefreshDataRef.current?.()
    }, seconds * 1000)
  }

  const selectRefreshInterval = (value) => {
    setRefreshIntervalValue(value)
    setShowRefreshMenu(false)
    onRefreshDataRef.current?.()
    setRefreshInterval(value)
  }

  const setThemeCookie = (theme) => {
    document.cookie = `${THEME_COOKIE_NAME}=${theme}; path=/; max-age=${THEME_COOKIE_MAX_AGE}; samesite=strict`
  }

  const applyTheme = () => {
    const isDark = wantsDarkMode()
    setDarkMode(isDark)
    document.documentElement.classList.toggle('dark', isDark)
  }

  const toggleDarkMode = () => {
    const newTheme = wantsDarkMode() ? 'light' : 'dark'
    setThemeCookie(newTheme)
    applyTheme()
  }

  useEffect(() => {
    setRefreshInterval(refreshIntervalValue)
    applyTheme()
    const handleClickOutside = (event) => {
      if (containerRef.current && !containerRef.current.contains(event.target)) {
        setShowRefreshMenu(false)
      }
    }
    document.addEventListener('click', handleClickOutside)
    return () => {
      if (refreshIntervalHandler.current) {
        clearInterval(refreshIntervalHandler.current)
      }
      document.removeEventListener('click', handleClickOutside)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div
      ref={containerRef}
      className="fixed bottom-4 left-4 z-50 animate-slideIn"
    >
      <div className="flex items-center gap-1 bg-background/95 backdrop-blur-sm border rounded-full shadow-md p-1 hover:-translate-y-0.5 hover:shadow-xl transition-all">
        <button
          onClick={() => setShowRefreshMenu(!showRefreshMenu)}
          aria-label={`Refresh interval: ${formatRefreshInterval(refreshIntervalValue)}`}
          aria-expanded={showRefreshMenu}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-full hover:bg-accent transition-colors relative"
        >
          <RefreshCw className="w-3.5 h-3.5 text-muted-foreground" />
          <span className="text-xs font-medium">{formatRefreshInterval(refreshIntervalValue)}</span>

          {showRefreshMenu && (
            <div
              onClick={(e) => e.stopPropagation()}
              className="absolute bottom-full left-0 mb-2 bg-popover border rounded-lg shadow-lg overflow-hidden"
            >
              {REFRESH_INTERVALS.map((interval) => (
                <button
                  key={interval.value}
                  onClick={() => selectRefreshInterval(interval.value)}
                  className={`block w-full px-4 py-2 text-xs text-left hover:bg-accent transition-colors ${
                    refreshIntervalValue === interval.value && 'bg-accent'
                  }`}
                >
                  {interval.label}
                </button>
              ))}
            </div>
          )}
        </button>

        <div className="h-5 w-px bg-border/50" />

        <div className="relative group">
          <button
            onClick={toggleDarkMode}
            aria-label={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
            className="p-1.5 rounded-full hover:bg-accent transition-colors"
          >
            {darkMode ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
          </button>
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-popover text-popover-foreground text-xs rounded-md shadow-md opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap">
            {darkMode ? 'Light mode' : 'Dark mode'}
          </div>
        </div>
      </div>
    </div>
  )
}
