import React, { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  RefreshCw,
  ArrowUpCircle,
  ArrowDownCircle,
  PlayCircle,
  Activity,
  Timer,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import StatusBadge from '@/components/StatusBadge'
import EndpointCard from '@/components/EndpointCard'
import Settings from '@/components/Settings'
import Pagination from '@/components/Pagination'
import Loading from '@/components/Loading'
import ResponseTimeChart from '@/components/ResponseTimeChart'
import { generatePrettyTimeAgo, generatePrettyTimeDifference } from '@/utils/time'

export default function EndpointDetails({ onShowTooltip }) {
  const navigate = useNavigate()
  const { key } = useParams()

  const [endpointStatus, setEndpointStatus] = useState(null)
  const [currentStatus, setCurrentStatus] = useState(null)
  const [events, setEvents] = useState([])
  const [currentPage, setCurrentPage] = useState(1)
  const [showAverageResponseTime, setShowAverageResponseTime] = useState(
    localStorage.getItem('gatus:show-average-response-time') !== 'false'
  )
  const [selectedChartDuration, setSelectedChartDuration] = useState('24h')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [showResponseTimeChartAndBadges, setShowResponseTimeChartAndBadges] = useState(false)
  const resultPageSize = 50

  const latestResult = useMemo(() => {
    if (!currentStatus || !currentStatus.results || currentStatus.results.length === 0) return null
    return currentStatus.results[currentStatus.results.length - 1]
  }, [currentStatus])

  const currentHealthStatus = latestResult
    ? latestResult.success
      ? 'healthy'
      : 'unhealthy'
    : 'unknown'

  const hostname = latestResult?.hostname || null

  const toggleShowAverageResponseTime = () => {
    const next = !showAverageResponseTime
    setShowAverageResponseTime(next)
    localStorage.setItem('gatus:show-average-response-time', next ? 'true' : 'false')
  }

  const pageAverageResponseTime = useMemo(() => {
    if (!endpointStatus || !endpointStatus.results || endpointStatus.results.length === 0)
      return 'N/A'
    let total = 0
    let count = 0
    for (const result of endpointStatus.results) {
      if (result.duration) {
        total += result.duration
        count++
      }
    }
    if (count === 0) return 'N/A'
    return `${Math.round(total / count / 1000000)}ms`
  }, [endpointStatus])

  const pageResponseTimeRange = useMemo(() => {
    if (!endpointStatus || !endpointStatus.results || endpointStatus.results.length === 0)
      return 'N/A'
    let min = Infinity
    let max = 0
    let hasData = false

    for (const result of endpointStatus.results) {
      const duration = result.duration
      if (duration) {
        min = Math.min(min, duration)
        max = Math.max(max, duration)
        hasData = true
      }
    }

    if (!hasData) return 'N/A'
    const minMs = Math.trunc(min / 1000000)
    const maxMs = Math.trunc(max / 1000000)
    if (minMs === maxMs) return `${minMs}ms`
    return `${minMs}-${maxMs}ms`
  }, [endpointStatus])

  const lastCheckTime = useMemo(() => {
    if (!currentStatus || !currentStatus.results || currentStatus.results.length === 0)
      return 'Never'
    return generatePrettyTimeAgo(currentStatus.results[currentStatus.results.length - 1].timestamp)
  }, [currentStatus])

  const fetchData = async () => {
    setIsRefreshing(true)
    try {
      const response = await fetch(
        `/api/v1/endpoints/${key}/statuses?page=${currentPage}&pageSize=${resultPageSize}`,
        { credentials: 'include' }
      )
      if (response.status === 200) {
        const data = await response.json()
        setEndpointStatus(data)
        if (currentPage === 1) setCurrentStatus(data)

        const processedEvents = []
        if (data.events && data.events.length > 0) {
          for (let i = data.events.length - 1; i >= 0; i--) {
            const event = { ...data.events[i] }
            if (i === data.events.length - 1) {
              if (event.type === 'UNHEALTHY') event.fancyText = 'Endpoint is unhealthy'
              else if (event.type === 'HEALTHY') event.fancyText = 'Endpoint is healthy'
              else if (event.type === 'START') event.fancyText = 'Monitoring started'
            } else {
              const nextEvent = data.events[i + 1]
              if (event.type === 'HEALTHY') event.fancyText = 'Endpoint became healthy'
              else if (event.type === 'UNHEALTHY') {
                event.fancyText = nextEvent
                  ? 'Endpoint was unhealthy for ' +
                    generatePrettyTimeDifference(nextEvent.timestamp, event.timestamp)
                  : 'Endpoint became unhealthy'
              } else if (event.type === 'START') event.fancyText = 'Monitoring started'
            }
            event.fancyTimeAgo = generatePrettyTimeAgo(event.timestamp)
            processedEvents.push(event)
          }
        }
        setEvents(processedEvents)

        if (data.results && data.results.length > 0) {
          for (const r of data.results) {
            if (r.duration > 0) {
              setShowResponseTimeChartAndBadges(true)
              break
            }
          }
        }
      }
    } catch (error) {
      console.error('[Details][fetchData] Error:', error)
    } finally {
      setIsRefreshing(false)
    }
  }

  const goBack = () => navigate('/')
  const changePage = (page) => setCurrentPage(page)

  const prettifyTimestamp = (timestamp) => new Date(timestamp).toLocaleString()

  useEffect(() => {
    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage, key])

  if (!endpointStatus || !endpointStatus.name) {
    return (
      <div className="dashboard-container bg-background">
        <div className="container mx-auto px-4 py-8 max-w-7xl">
          <div className="flex items-center justify-center py-20">
            <Loading size="lg" />
          </div>
        </div>
        <Settings onRefreshData={fetchData} />
      </div>
    )
  }

  return (
    <div className="dashboard-container bg-background">
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="mb-6">
          <Button variant="ghost" className="mb-4" onClick={goBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Dashboard
          </Button>

          <div className="space-y-6">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-4xl font-bold tracking-tight">{endpointStatus.name}</h1>
                <div className="flex items-center gap-3 text-muted-foreground mt-2">
                  {endpointStatus.group && <span>Group: {endpointStatus.group}</span>}
                  {endpointStatus.group && hostname && <span>•</span>}
                  {hostname && <span>{hostname}</span>}
                </div>
              </div>
              <StatusBadge status={currentHealthStatus} />
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Current Status
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {currentHealthStatus === 'healthy' ? 'Operational' : 'Issues Detected'}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Avg Response Time
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{pageAverageResponseTime}</div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Response Time Range
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{pageResponseTimeRange}</div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Last Check
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{lastCheckTime}</div>
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Recent Checks</CardTitle>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={toggleShowAverageResponseTime}
                      title={
                        showAverageResponseTime
                          ? 'Show min-max response time'
                          : 'Show average response time'
                      }
                    >
                      {showAverageResponseTime ? (
                        <Activity className="h-5 w-5" />
                      ) : (
                        <Timer className="h-5 w-5" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={fetchData}
                      title="Refresh data"
                      disabled={isRefreshing}
                    >
                      <RefreshCw
                        className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`}
                      />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <EndpointCard
                    endpoint={endpointStatus}
                    maxResults={resultPageSize}
                    showAverageResponseTime={showAverageResponseTime}
                    onShowTooltip={onShowTooltip}
                    className="border-0 shadow-none bg-transparent p-0"
                  />
                  {endpointStatus.key && (
                    <div className="pt-4 border-t">
                      <Pagination
                        numberOfResultsPerPage={resultPageSize}
                        currentPageProp={currentPage}
                        onPageChange={changePage}
                      />
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {showResponseTimeChartAndBadges && (
              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle>Response Time Trend</CardTitle>
                      <select
                        value={selectedChartDuration}
                        onChange={(e) => setSelectedChartDuration(e.target.value)}
                        className="text-sm bg-background border rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-ring"
                      >
                        <option value="24h">24 hours</option>
                        <option value="7d">7 days</option>
                        <option value="30d">30 days</option>
                      </select>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <ResponseTimeChart
                      endpointKey={endpointStatus.key}
                      duration={selectedChartDuration}
                      serverUrl=""
                      events={endpointStatus.events || []}
                    />
                  </CardContent>
                </Card>

                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                  {['30d', '7d', '24h', '1h'].map((period) => (
                    <Card key={period}>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium text-muted-foreground text-center">
                          {period === '30d'
                            ? 'Last 30 days'
                            : period === '7d'
                            ? 'Last 7 days'
                            : period === '24h'
                            ? 'Last 24 hours'
                            : 'Last hour'}
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <img
                          src={`/api/v1/endpoints/${endpointStatus.key}/response-times/${period}/badge.svg`}
                          alt={`${period} response time`}
                          className="mx-auto mt-2"
                        />
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            <Card>
              <CardHeader>
                <CardTitle>Uptime Statistics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                  {['30d', '7d', '24h', '1h'].map((period) => (
                    <div key={period} className="text-center">
                      <p className="text-sm text-muted-foreground mb-2">
                        {period === '30d'
                          ? 'Last 30 days'
                          : period === '7d'
                          ? 'Last 7 days'
                          : period === '24h'
                          ? 'Last 24 hours'
                          : 'Last hour'}
                      </p>
                      <img
                        src={`/api/v1/endpoints/${endpointStatus.key}/uptimes/${period}/badge.svg`}
                        alt={`${period} uptime`}
                        className="mx-auto"
                      />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Current Health</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center">
                  <img
                    src={`/api/v1/endpoints/${endpointStatus.key}/health/badge.svg`}
                    alt="health badge"
                    className="mx-auto"
                  />
                </div>
              </CardContent>
            </Card>

            {events.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle>Events</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {events.map((event, idx) => (
                      <div
                        key={`${event.timestamp}-${idx}`}
                        className="flex items-start gap-4 pb-4 border-b last:border-0"
                      >
                        <div className="mt-1">
                          {event.type === 'HEALTHY' ? (
                            <ArrowUpCircle className="h-5 w-5 text-green-500" />
                          ) : event.type === 'UNHEALTHY' ? (
                            <ArrowDownCircle className="h-5 w-5 text-red-500" />
                          ) : (
                            <PlayCircle className="h-5 w-5 text-muted-foreground" />
                          )}
                        </div>
                        <div className="flex-1">
                          <p className="font-medium">{event.fancyText}</p>
                          <p className="text-sm text-muted-foreground">
                            {prettifyTimestamp(event.timestamp)} • {event.fancyTimeAgo}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </div>

      <Settings onRefreshData={fetchData} />
    </div>
  )
}
