import React, { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import StatusBadge from '@/components/StatusBadge'
import { generatePrettyTimeAgo } from '@/utils/time'

export default function EndpointCard({
  endpoint,
  maxResults = 50,
  showAverageResponseTime = true,
  onShowTooltip,
}) {
  const navigate = useNavigate()
  const [selectedResultIndex, setSelectedResultIndex] = useState(null)

  const latestResult = useMemo(() => {
    if (!endpoint.results || endpoint.results.length === 0) return null
    return endpoint.results[endpoint.results.length - 1]
  }, [endpoint.results])

  const currentStatus = latestResult
    ? latestResult.success
      ? 'healthy'
      : 'unhealthy'
    : 'unknown'

  const hostname = latestResult?.hostname || null

  const displayResults = useMemo(() => {
    const results = [...(endpoint.results || [])]
    while (results.length < maxResults) {
      results.unshift(null)
    }
    return results.slice(-maxResults)
  }, [endpoint.results, maxResults])

  const formattedResponseTime = useMemo(() => {
    if (!endpoint.results || endpoint.results.length === 0) return 'N/A'
    let total = 0
    let count = 0
    let min = Infinity
    let max = 0

    for (const result of endpoint.results) {
      if (result.duration) {
        const durationMs = result.duration / 1000000
        total += durationMs
        count++
        min = Math.min(min, durationMs)
        max = Math.max(max, durationMs)
      }
    }

    if (count === 0) return 'N/A'

    if (showAverageResponseTime) {
      const avgMs = Math.round(total / count)
      return `~${avgMs}ms`
    } else {
      const minMs = Math.trunc(min)
      const maxMs = Math.trunc(max)
      if (minMs === maxMs) return `${minMs}ms`
      return `${minMs}-${maxMs}ms`
    }
  }, [endpoint.results, showAverageResponseTime])

  const oldestResultTime = useMemo(() => {
    if (!endpoint.results || endpoint.results.length === 0) return ''
    const oldestResultIndex = Math.max(0, endpoint.results.length - maxResults)
    return generatePrettyTimeAgo(endpoint.results[oldestResultIndex].timestamp)
  }, [endpoint.results, maxResults])

  const newestResultTime = useMemo(() => {
    if (!endpoint.results || endpoint.results.length === 0) return ''
    return generatePrettyTimeAgo(endpoint.results[endpoint.results.length - 1].timestamp)
  }, [endpoint.results])

  const navigateToDetails = () => navigate(`/endpoints/${endpoint.key}`)

  const handleMouseEnter = (result, event) => onShowTooltip?.(result, event, 'hover')
  const handleMouseLeave = (_result, event) => onShowTooltip?.(null, event, 'hover')

  const handleClick = (result, event, index) => {
    window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
    if (selectedResultIndex === index) {
      setSelectedResultIndex(null)
      onShowTooltip?.(null, event, 'click')
    } else {
      setSelectedResultIndex(index)
      onShowTooltip?.(result, event, 'click')
    }
  }

  useEffect(() => {
    const handleClearSelection = () => setSelectedResultIndex(null)
    window.addEventListener('clear-data-point-selection', handleClearSelection)
    return () => window.removeEventListener('clear-data-point-selection', handleClearSelection)
  }, [])

  return (
    <Card className="endpoint h-full flex flex-col transition hover:shadow-lg hover:scale-[1.01] dark:hover:border-gray-700">
      <CardHeader className="endpoint-header px-3 sm:px-6 pt-3 sm:pt-6 pb-2 space-y-0">
        <div className="flex items-start justify-between gap-2 sm:gap-3">
          <div className="flex-1 min-w-0 overflow-hidden">
            <CardTitle className="text-base sm:text-lg truncate">
              <span
                className="hover:text-primary cursor-pointer hover:underline text-sm sm:text-base block truncate"
                onClick={navigateToDetails}
                onKeyDown={(e) => e.key === 'Enter' && navigateToDetails()}
                title={endpoint.name}
                role="link"
                tabIndex={0}
                aria-label={`View details for ${endpoint.name}`}
              >
                {endpoint.name}
              </span>
            </CardTitle>
            <div className="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground min-h-[1.25rem]">
              {endpoint.group && (
                <span className="truncate" title={endpoint.group}>
                  {endpoint.group}
                </span>
              )}
              {endpoint.group && hostname && <span>•</span>}
              {hostname && (
                <span className="truncate" title={hostname}>
                  {hostname}
                </span>
              )}
            </div>
          </div>
          <div className="flex-shrink-0 ml-2">
            <StatusBadge status={currentStatus} />
          </div>
        </div>
      </CardHeader>
      <CardContent className="endpoint-content flex-1 pb-3 sm:pb-4 px-3 sm:px-6 pt-2">
        <div className="space-y-2">
          <div>
            <div className="flex items-center justify-between mb-1">
              <div className="flex-1"></div>
              <p
                className="text-xs text-muted-foreground"
                title={
                  showAverageResponseTime
                    ? 'Average response time'
                    : 'Minimum and maximum response time'
                }
              >
                {formattedResponseTime}
              </p>
            </div>
            <div className="flex gap-0.5">
              {displayResults.map((result, index) => {
                const isSelected = selectedResultIndex === index
                const baseClass = 'flex-1 h-6 sm:h-8 rounded-sm transition-all'
                let colorClass = 'bg-gray-200 dark:bg-gray-700'
                if (result) {
                  colorClass = result.success
                    ? isSelected
                      ? 'bg-green-700 cursor-pointer'
                      : 'bg-green-500 hover:bg-green-700 cursor-pointer'
                    : isSelected
                    ? 'bg-red-700 cursor-pointer'
                    : 'bg-red-500 hover:bg-red-700 cursor-pointer'
                }
                return (
                  <div
                    key={index}
                    className={`${baseClass} ${colorClass}`}
                    onMouseEnter={(e) => result && handleMouseEnter(result, e)}
                    onMouseLeave={(e) => result && handleMouseLeave(result, e)}
                    onClick={(e) => result && handleClick(result, e, index)}
                  />
                )
              })}
            </div>
            <div className="flex items-center justify-between text-xs text-muted-foreground mt-1">
              <span>{oldestResultTime}</span>
              <span>{newestResultTime}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
