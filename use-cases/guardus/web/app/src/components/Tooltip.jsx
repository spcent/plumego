import React, { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { prettifyTimestamp } from '@/utils/time'

export default function Tooltip({ event, result, isPersistent = false }) {
  const [hidden, setHidden] = useState(true)
  const [top, setTop] = useState(0)
  const [left, setLeft] = useState(0)
  const tooltipRef = useRef(null)
  const targetElementRef = useRef(null)
  const location = useLocation()

  const isSuiteResult = result && result.endpointResults !== undefined
  const endpointCount = isSuiteResult && result.endpointResults ? result.endpointResults.length : 0
  const successCount = isSuiteResult && result.endpointResults
    ? result.endpointResults.filter((e) => e.success).length
    : 0

  const updatePosition = async () => {
    if (!targetElementRef.current || !tooltipRef.current || hidden) return

    const targetRect = targetElementRef.current.getBoundingClientRect()
    const tooltipRect = tooltipRef.current.getBoundingClientRect()

    const scrollTop = window.pageYOffset || document.documentElement.scrollTop
    const scrollLeft = window.pageXOffset || document.documentElement.scrollLeft

    let newTop = targetRect.bottom + scrollTop + 8
    let newLeft = targetRect.left + scrollLeft

    const spaceBelow = window.innerHeight - targetRect.bottom
    const spaceAbove = targetRect.top

    if (spaceBelow < tooltipRect.height + 20) {
      if (spaceAbove > tooltipRect.height + 20) {
        newTop = targetRect.top + scrollTop - tooltipRect.height - 8
      } else {
        if (spaceAbove > spaceBelow) {
          newTop = scrollTop + 10
        } else {
          newTop = scrollTop + window.innerHeight - tooltipRect.height - 10
        }
      }
    }

    const spaceRight = window.innerWidth - targetRect.left
    if (spaceRight < tooltipRect.width + 20) {
      newLeft = targetRect.right + scrollLeft - tooltipRect.width
      if (newLeft < scrollLeft + 10) {
        newLeft = scrollLeft + 10
      }
    }

    setTop(Math.round(newTop))
    setLeft(Math.round(newLeft))
  }

  const reposition = async () => {
    if (!event || !event.type) return
    if (event.type === 'mouseenter' || event.type === 'click') {
      const target = event.target
      targetElementRef.current = target
      setHidden(false)
      requestAnimationFrame(() => updatePosition())
    } else if (event.type === 'mouseleave') {
      if (!isPersistent) {
        setHidden(true)
        targetElementRef.current = null
      }
    }
  }

  useEffect(() => {
    const handleResize = () => updatePosition()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [hidden])

  useEffect(() => {
    reposition()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [event?.type, result, isPersistent])

  useEffect(() => {
    if (!isPersistent && !result) {
      setHidden(true)
    } else if (result && (isPersistent || event?.type === 'mouseenter')) {
      setHidden(false)
      requestAnimationFrame(() => updatePosition())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPersistent, result])

  useEffect(() => {
    setHidden(true)
    targetElementRef.current = null
  }, [location.pathname])

  return (
    <div
      id="tooltip"
      ref={tooltipRef}
      className={`absolute z-50 px-3 py-2 text-sm rounded-md shadow-lg border transition-all duration-200 bg-popover text-popover-foreground border-border ${
        hidden ? 'invisible opacity-0' : 'visible opacity-100'
      }`}
      style={{ top: `${top}px`, left: `${left}px` }}
    >
      {result && (
        <div className="space-y-2">
          {isSuiteResult && (
            <div className="flex items-center gap-2">
              <span
                className={`inline-block w-2 h-2 rounded-full ${
                  result.success ? 'bg-green-500' : 'bg-red-500'
                }`}
              ></span>
              <span className="text-xs font-semibold">
                {result.success ? 'Suite Passed' : 'Suite Failed'}
              </span>
            </div>
          )}

          <div>
            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Timestamp
            </div>
            <div className="font-mono text-xs">{prettifyTimestamp(result.timestamp)}</div>
          </div>

          {isSuiteResult && result.endpointResults && (
            <div>
              <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Endpoints
              </div>
              <div className="font-mono text-xs">
                <span
                  className={successCount === endpointCount ? 'text-green-500' : 'text-yellow-500'}
                >
                  {successCount}/{endpointCount} passed
                </span>
              </div>
              {result.endpointResults.length > 0 && (
                <div className="mt-1 space-y-0.5">
                  {result.endpointResults.slice(0, 5).map((endpoint, index) => (
                    <div key={index} className="flex items-center gap-1 text-xs">
                      <span className={endpoint.success ? 'text-green-500' : 'text-red-500'}>
                        {endpoint.success ? '✓' : '✗'}
                      </span>
                      <span className="truncate">{endpoint.name}</span>
                      <span className="text-muted-foreground">
                        ({Math.trunc(endpoint.duration / 1000000)}ms)
                      </span>
                    </div>
                  ))}
                  {result.endpointResults.length > 5 && (
                    <div className="text-xs text-muted-foreground">
                      ... and {result.endpointResults.length - 5} more
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          <div>
            <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {isSuiteResult ? 'Total Duration' : 'Response Time'}
            </div>
            <div className="font-mono text-xs">
              {Math.trunc(result.duration / 1000000)}ms
            </div>
          </div>

          {!isSuiteResult && result.conditionResults && result.conditionResults.length > 0 && (
            <div>
              <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Conditions
              </div>
              <div className="font-mono text-xs space-y-0.5">
                {result.conditionResults.map((conditionResult, index) => (
                  <div key={index} className="flex items-start gap-1">
                    <span className={conditionResult.success ? 'text-green-500' : 'text-red-500'}>
                      {conditionResult.success ? '✓' : '✗'}
                    </span>
                    <span className="break-all">{conditionResult.condition}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {result.errors && result.errors.length > 0 && (
            <div>
              <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Errors
              </div>
              <div className="font-mono text-xs space-y-0.5">
                {result.errors.map((error, index) => (
                  <div key={index} className="text-red-500">
                    • {error}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
