import React, { useEffect, useRef, useState, lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import { Menu, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import Social from '@/components/Social'
import Tooltip from '@/components/Tooltip'
import Loading from '@/components/Loading'

const Home = lazy(() => import('@/views/Home'))
const EndpointDetails = lazy(() => import('@/views/EndpointDetails'))

export default function App() {
  const [retrievedConfig, setRetrievedConfig] = useState(false)
  const [config, setConfig] = useState({ oidc: false, authenticated: true })
  const [announcements, setAnnouncements] = useState([])
  const [tooltipState, setTooltipState] = useState({})
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [tooltipIsPersistent, setTooltipIsPersistent] = useState(false)
  const configIntervalRef = useRef(null)

  const logo =
    window.config?.logo && window.config.logo !== '{{ .UI.Logo }}' ? window.config.logo : ''
  const header =
    window.config?.header && window.config.header !== '{{ .UI.Header }}'
      ? window.config.header
      : 'Gatus'
  const link =
    window.config?.link && window.config.link !== '{{ .UI.Link }}' ? window.config.link : null
  const buttons = window.config?.buttons || []

  const fetchConfig = async () => {
    try {
      const response = await fetch(`/api/v1/config`, { credentials: 'include' })
      if (response.status === 200) {
        const data = await response.json()
        setConfig(data)
        setAnnouncements(data.announcements || [])
      }
      setRetrievedConfig(true)
    } catch (error) {
      console.error('Failed to fetch config:', error)
      setRetrievedConfig(true)
    }
  }

  const showTooltip = (result, event, action = 'hover') => {
    if (action === 'click') {
      if (!result) {
        setTooltipState({})
        setTooltipIsPersistent(false)
      } else {
        setTooltipState({ result, event })
        setTooltipIsPersistent(true)
      }
    } else if (action === 'hover') {
      if (!tooltipIsPersistent) {
        setTooltipState({ result, event })
      }
    }
  }

  const handleDocumentClick = (event) => {
    if (tooltipIsPersistent) {
      const tooltipElement = document.getElementById('tooltip')
      const clickedDataPoint = event.target.closest('.flex-1.h-6, .flex-1.h-8')
      if (
        tooltipElement &&
        !tooltipElement.contains(event.target) &&
        !clickedDataPoint
      ) {
        setTooltipState({})
        setTooltipIsPersistent(false)
        window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
      }
    }
  }

  useEffect(() => {
    fetchConfig()
    configIntervalRef.current = setInterval(fetchConfig, 600000)
    document.addEventListener('click', handleDocumentClick)
    return () => {
      if (configIntervalRef.current) clearInterval(configIntervalRef.current)
      document.removeEventListener('click', handleDocumentClick)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tooltipIsPersistent])

  if (!retrievedConfig) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background text-foreground">
        <Loading size="lg" />
      </div>
    )
  }

  return (
    <div className="bg-background text-foreground">
      <div className="relative">
        <header className="border-b bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/60">
          <div className="container mx-auto px-4 py-4 max-w-7xl">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                {link ? (
                  <a
                    href={link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-3 hover:opacity-80 transition-opacity"
                  >
                    <div className="w-12 h-12 flex items-center justify-center">
                      {logo ? (
                        <img src={logo} alt="Gatus" className="w-full h-full object-contain" />
                      ) : (
                        <span className="text-2xl font-bold">🦉</span>
                      )}
                    </div>
                    <div>
                      <h1 className="text-2xl font-bold tracking-tight">{header}</h1>
                      {buttons.length > 0 && (
                        <p className="text-sm text-muted-foreground">System Monitoring Dashboard</p>
                      )}
                    </div>
                  </a>
                ) : (
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 flex items-center justify-center">
                      {logo ? (
                        <img src={logo} alt="Gatus" className="w-full h-full object-contain" />
                      ) : (
                        <span className="text-2xl font-bold">🦉</span>
                      )}
                    </div>
                    <div>
                      <h1 className="text-2xl font-bold tracking-tight">{header}</h1>
                      {buttons.length > 0 && (
                        <p className="text-sm text-muted-foreground">System Monitoring Dashboard</p>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div className="flex items-center gap-2">
                {buttons.length > 0 && (
                  <nav className="hidden md:flex items-center gap-1">
                    {buttons.map((button) => (
                      <a
                        key={button.name}
                        href={button.link}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                      >
                        {button.name}
                      </a>
                    ))}
                  </nav>
                )}

                {buttons.length > 0 && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="md:hidden"
                    onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                  >
                    {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
                  </Button>
                )}
              </div>
            </div>

            {buttons.length > 0 && mobileMenuOpen && (
              <nav className="md:hidden mt-4 pt-4 border-t space-y-1">
                {buttons.map((button) => (
                  <a
                    key={button.name}
                    href={button.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    {button.name}
                  </a>
                ))}
              </nav>
            )}
          </div>
        </header>

        <main className="relative">
          <Suspense fallback={
            <div className="flex items-center justify-center min-h-[400px]">
              <Loading size="lg" />
            </div>
          }>
            <Routes>
              <Route path="/" element={<Home announcements={announcements} onShowTooltip={showTooltip} />} />
              <Route path="/endpoints/:key" element={<EndpointDetails onShowTooltip={showTooltip} />} />
            </Routes>
          </Suspense>
        </main>

        <footer className="border-t mt-auto">
          <div className="container mx-auto px-4 py-6 max-w-7xl">
            <div className="flex flex-col items-center gap-4">
              <div className="text-sm text-muted-foreground text-center">
                Powered by{' '}
                <a
                  href="https://gatus.io"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-medium text-emerald-800 hover:text-emerald-600"
                >
                  Gatus
                </a>
              </div>
              <Social />
            </div>
          </div>
        </footer>
      </div>

      <Tooltip
        result={tooltipState.result}
        event={tooltipState.event}
        isPersistent={tooltipIsPersistent}
      />
    </div>
  )
}
