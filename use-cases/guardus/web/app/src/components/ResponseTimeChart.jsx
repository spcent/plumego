import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Line } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  TimeScale,
} from 'chart.js'
import annotationPlugin from 'chartjs-plugin-annotation'
import 'chartjs-adapter-date-fns'
import { generatePrettyTimeDifference } from '@/utils/time'
import Loading from './Loading'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  TimeScale,
  annotationPlugin
)

const getEventColor = () => 'rgba(239, 68, 68, 0.8)'

export default function ResponseTimeChart({
  endpointKey,
  duration,
  serverUrl = '',
  events = [],
}) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [timestamps, setTimestamps] = useState([])
  const [values, setValues] = useState([])
  const [isDark, setIsDark] = useState(document.documentElement.classList.contains('dark'))
  const [hoveredEventIndex, setHoveredEventIndex] = useState(null)
  const observerRef = useRef(null)

  const filteredEvents = useMemo(() => {
    if (!events || events.length === 0) return []
    const now = new Date()
    let fromTime
    switch (duration) {
      case '24h':
        fromTime = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        break
      case '7d':
        fromTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case '30d':
        fromTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      default:
        return []
    }

    const unhealthyEvents = []
    for (let i = 0; i < events.length; i++) {
      const event = events[i]
      if (event.type !== 'UNHEALTHY') continue

      const eventTime = new Date(event.timestamp)
      if (eventTime < fromTime || eventTime > now) continue

      let durationStr = null
      let isOngoing = false
      if (i + 1 < events.length) {
        const nextEvent = events[i + 1]
        durationStr = generatePrettyTimeDifference(nextEvent.timestamp, event.timestamp)
      } else {
        durationStr = generatePrettyTimeDifference(now, event.timestamp)
        isOngoing = true
      }

      unhealthyEvents.push({
        ...event,
        duration: durationStr,
        isOngoing,
      })
    }
    return unhealthyEvents
  }, [events, duration])

  const chartData = useMemo(() => {
    if (timestamps.length === 0) {
      return { labels: [], datasets: [] }
    }
    const labels = timestamps.map((ts) => new Date(ts))
    return {
      labels,
      datasets: [
        {
          label: 'Response Time (ms)',
          data: values,
          borderColor: isDark ? 'rgb(96, 165, 250)' : 'rgb(59, 130, 246)',
          backgroundColor: isDark ? 'rgba(96, 165, 250, 0.1)' : 'rgba(59, 130, 246, 0.1)',
          borderWidth: 2,
          pointRadius: 2,
          pointHoverRadius: 4,
          tension: 0.1,
          fill: true,
        },
      ],
    }
  }, [timestamps, values, isDark])

  const chartOptions = useMemo(() => {
    const maxY = values.length > 0 ? Math.max(...values) : 0
    const midY = maxY / 2

    return {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: isDark ? 'rgba(31, 41, 55, 0.95)' : 'rgba(255, 255, 255, 0.95)',
          titleColor: isDark ? '#f9fafb' : '#111827',
          bodyColor: isDark ? '#d1d5db' : '#374151',
          borderColor: isDark ? '#4b5563' : '#e5e7eb',
          borderWidth: 1,
          padding: 12,
          displayColors: false,
          callbacks: {
            title: (tooltipItems) => {
              if (tooltipItems.length > 0) {
                const date = new Date(tooltipItems[0].parsed.x)
                return date.toLocaleString()
              }
              return ''
            },
            label: (context) => `${context.parsed.y}ms`,
          },
        },
        annotation: {
          annotations: filteredEvents.reduce((acc, event, index) => {
            const eventTimestamp = new Date(event.timestamp).getTime()
            let closestValue = 0

            if (timestamps.length > 0 && values.length > 0) {
              const closestIndex = timestamps.reduce((closest, ts, idx) => {
                const tsTime = new Date(ts).getTime()
                const currentDistance = Math.abs(tsTime - eventTimestamp)
                const closestDistance = Math.abs(
                  new Date(timestamps[closest]).getTime() - eventTimestamp
                )
                return currentDistance < closestDistance ? idx : closest
              }, 0)
              closestValue = values[closestIndex]
            }

            const position = closestValue <= midY ? 'end' : 'start'

            acc[`event-${index}`] = {
              type: 'line',
              xMin: new Date(event.timestamp),
              xMax: new Date(event.timestamp),
              borderColor: getEventColor(),
              borderWidth: 1,
              borderDash: [5, 5],
              enter() {
                setHoveredEventIndex(index)
              },
              leave() {
                setHoveredEventIndex(null)
              },
              label: {
                display: () => hoveredEventIndex === index,
                content: [
                  event.isOngoing ? 'Status: ONGOING' : 'Status: RESOLVED',
                  `Unhealthy for ${event.duration}`,
                  `Started at ${new Date(event.timestamp).toLocaleString()}`,
                ],
                backgroundColor: getEventColor(),
                color: '#ffffff',
                font: { size: 11 },
                padding: 6,
                position,
              },
            }
            return acc
          }, {}),
        },
      },
      scales: {
        x: {
          type: 'time',
          time: {
            unit: duration === '24h' ? 'hour' : 'day',
            displayFormats: {
              hour: 'MMM d, ha',
              day: 'MMM d',
            },
          },
          grid: {
            color: isDark ? 'rgba(75, 85, 99, 0.3)' : 'rgba(229, 231, 235, 0.8)',
            drawBorder: false,
          },
          ticks: {
            color: isDark ? '#9ca3af' : '#6b7280',
            maxRotation: 0,
            autoSkipPadding: 20,
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: isDark ? 'rgba(75, 85, 99, 0.3)' : 'rgba(229, 231, 235, 0.8)',
            drawBorder: false,
          },
          ticks: {
            color: isDark ? '#9ca3af' : '#6b7280',
            callback: (value) => `${value}ms`,
          },
        },
      },
    }
  }, [values, timestamps, filteredEvents, duration, isDark, hoveredEventIndex])

  const fetchData = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch(
        `${serverUrl}/api/v1/endpoints/${endpointKey}/response-times/${duration}/history`,
        { credentials: 'include' }
      )
      if (response.status === 200) {
        const data = await response.json()
        setTimestamps(data.timestamps || [])
        setValues(data.values || [])
      } else {
        setError('Failed to load chart data')
      }
    } catch (err) {
      setError('Failed to load chart data')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [duration, endpointKey])

  useEffect(() => {
    observerRef.current = new MutationObserver(() => {
      setIsDark(document.documentElement.classList.contains('dark'))
    })
    observerRef.current.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })
    return () => observerRef.current?.disconnect()
  }, [])

  return (
    <div className="relative w-full" style={{ height: '300px' }}>
      {loading ? (
        <div className="absolute inset-0 flex items-center justify-center bg-background/50">
          <Loading />
        </div>
      ) : error ? (
        <div className="absolute inset-0 flex items-center justify-center text-muted-foreground">
          {error}
        </div>
      ) : (
        <Line data={chartData} options={chartOptions} />
      )}
    </div>
  )
}
