import React, { useEffect, useMemo, useState } from 'react'
import {
  Activity,
  Timer,
  RefreshCw,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  CheckCircle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import EndpointCard from '@/components/EndpointCard'
import SearchBar from '@/components/SearchBar'
import Settings from '@/components/Settings'
import Loading from '@/components/Loading'
import AnnouncementBanner from '@/components/AnnouncementBanner'
import PastAnnouncements from '@/components/PastAnnouncements'

export default function Home({ announcements = [], onShowTooltip }) {
  const [endpointStatuses, setEndpointStatuses] = useState([])
  const [loading, setLoading] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [showOnlyFailing, setShowOnlyFailing] = useState(false)
  const [showRecentFailures, setShowRecentFailures] = useState(false)
  const [showAverageResponseTime, setShowAverageResponseTime] = useState(
    localStorage.getItem('gatus:show-average-response-time') !== 'false'
  )
  const [groupByGroup, setGroupByGroup] = useState(false)
  const [sortBy, setSortBy] = useState(localStorage.getItem('gatus:sort-by') || 'name')
  const [uncollapsedGroups, setUncollapsedGroups] = useState(new Set())
  const resultPageSize = 50
  const itemsPerPage = 96

  const activeAnnouncements = useMemo(
    () => (announcements ? announcements.filter((a) => !a.archived) : []),
    [announcements]
  )
  const archivedAnnouncements = useMemo(
    () => (announcements ? announcements.filter((a) => a.archived) : []),
    [announcements]
  )

  const filteredEndpoints = useMemo(() => {
    let filtered = [...endpointStatuses]

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (endpoint) =>
          endpoint.name.toLowerCase().includes(query) ||
          (endpoint.group && endpoint.group.toLowerCase().includes(query))
      )
    }

    if (showOnlyFailing) {
      filtered = filtered.filter((endpoint) => {
        if (!endpoint.results || endpoint.results.length === 0) return false
        const latestResult = endpoint.results[endpoint.results.length - 1]
        return !latestResult.success
      })
    }

    if (showRecentFailures) {
      filtered = filtered.filter((endpoint) => {
        if (!endpoint.results || endpoint.results.length === 0) return false
        return endpoint.results.some((result) => !result.success)
      })
    }

    if (sortBy === 'health') {
      filtered.sort((a, b) => {
        const aHealthy =
          a.results && a.results.length > 0 && a.results[a.results.length - 1].success
        const bHealthy =
          b.results && b.results.length > 0 && b.results[b.results.length - 1].success
        if (!aHealthy && bHealthy) return -1
        if (aHealthy && !bHealthy) return 1
        return a.name.localeCompare(b.name)
      })
    }

    return filtered
  }, [endpointStatuses, searchQuery, showOnlyFailing, showRecentFailures, sortBy])

  const totalPages = Math.ceil(filteredEndpoints.length / itemsPerPage)

  const groupedEndpoints = useMemo(() => {
    if (!groupByGroup) return null
    const grouped = {}
    filteredEndpoints.forEach((endpoint) => {
      const group = endpoint.group || 'No Group'
      if (!grouped[group]) grouped[group] = []
      grouped[group].push(endpoint)
    })
    const sortedGroups = Object.keys(grouped).sort((a, b) => {
      if (a === 'No Group') return 1
      if (b === 'No Group') return -1
      return a.localeCompare(b)
    })
    const result = {}
    sortedGroups.forEach((group) => (result[group] = grouped[group]))
    return result
  }, [groupByGroup, filteredEndpoints])

  const paginatedEndpoints = useMemo(() => {
    if (groupByGroup) return groupedEndpoints
    const start = (currentPage - 1) * itemsPerPage
    const end = start + itemsPerPage
    return filteredEndpoints.slice(start, end)
  }, [groupByGroup, groupedEndpoints, filteredEndpoints, currentPage])

  const visiblePages = useMemo(() => {
    const pages = []
    const maxVisible = 5
    let start = Math.max(1, currentPage - Math.floor(maxVisible / 2))
    let end = Math.min(totalPages, start + maxVisible - 1)
    if (end - start < maxVisible - 1) start = Math.max(1, end - maxVisible + 1)
    for (let i = start; i <= end; i++) pages.push(i)
    return pages
  }, [currentPage, totalPages])

  const fetchData = async () => {
    const isInitialLoad = endpointStatuses.length === 0
    if (isInitialLoad) setLoading(true)
    try {
      const endpointResponse = await fetch(
        `/api/v1/endpoints/statuses?page=1&pageSize=${resultPageSize}`,
        { credentials: 'include' }
      )
      if (endpointResponse.status === 200) {
        const data = await endpointResponse.json()
        setEndpointStatuses(data)
      }
    } catch (error) {
      console.error('[Home][fetchData] Error:', error)
    } finally {
      if (isInitialLoad) setLoading(false)
    }
  }

  const refreshData = () => {
    setEndpointStatuses([])
    fetchData()
  }

  const handleSearch = (query) => {
    setSearchQuery(query)
    setCurrentPage(1)
  }

  const goToPage = (page) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const toggleShowAverageResponseTime = () => {
    const next = !showAverageResponseTime
    setShowAverageResponseTime(next)
    localStorage.setItem('gatus:show-average-response-time', next ? 'true' : 'false')
  }

  const calculateUnhealthyCount = (endpoints) => {
    return endpoints.filter((endpoint) => {
      if (!endpoint.results || endpoint.results.length === 0) return false
      const latestResult = endpoint.results[endpoint.results.length - 1]
      return !latestResult.success
    }).length
  }

  const toggleGroupCollapse = (groupName) => {
    const next = new Set(uncollapsedGroups)
    if (next.has(groupName)) next.delete(groupName)
    else next.add(groupName)
    setUncollapsedGroups(next)
    localStorage.setItem('gatus:uncollapsed-groups', JSON.stringify(Array.from(next)))
    localStorage.removeItem('gatus:collapsed-groups')
  }

  const initializeCollapsedGroups = () => {
    try {
      const saved = localStorage.getItem('gatus:uncollapsed-groups')
      if (saved) setUncollapsedGroups(new Set(JSON.parse(saved)))
    } catch (e) {
      console.warn('Failed to parse saved uncollapsed groups:', e)
      localStorage.removeItem('gatus:uncollapsed-groups')
    }
  }

  const dashboardHeading =
    window.config?.dashboardHeading &&
    window.config.dashboardHeading !== '{{ .UI.DashboardHeading }}'
      ? window.config.dashboardHeading
      : 'Health Dashboard'

  const dashboardSubheading =
    window.config?.dashboardSubheading &&
    window.config.dashboardSubheading !== '{{ .UI.DashboardSubheading }}'
      ? window.config.dashboardSubheading
      : 'Monitor the health of your endpoints in real-time'

  useEffect(() => {
    fetchData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="dashboard-container bg-background">
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="mb-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-4xl font-bold tracking-tight">{dashboardHeading}</h1>
              <p className="text-muted-foreground mt-2">{dashboardSubheading}</p>
            </div>
            <div className="flex items-center gap-4">
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
              <Button variant="ghost" size="icon" onClick={refreshData} title="Refresh data">
                <RefreshCw className="h-5 w-5" />
              </Button>
            </div>
          </div>

          <AnnouncementBanner announcements={activeAnnouncements} />

          <SearchBar
            onSearch={handleSearch}
            onShowOnlyFailingChange={setShowOnlyFailing}
            onShowRecentFailuresChange={setShowRecentFailures}
            onGroupByGroupChange={setGroupByGroup}
            onSortByChange={setSortBy}
            onInitializeCollapsedGroups={initializeCollapsedGroups}
          />
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loading size="lg" />
          </div>
        ) : filteredEndpoints.length === 0 ? (
          <div className="text-center py-20">
            <AlertCircle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">No endpoints found</h3>
            <p className="text-muted-foreground">
              {searchQuery || showOnlyFailing || showRecentFailures
                ? 'Try adjusting your filters'
                : 'No endpoints are configured'}
            </p>
          </div>
        ) : groupByGroup ? (
          <div className="space-y-6">
            {Object.entries(paginatedEndpoints).map(([group, items]) => (
              <div
                key={group}
                className="endpoint-group border rounded-lg overflow-hidden"
              >
                <div
                  onClick={() => toggleGroupCollapse(group)}
                  className="endpoint-group-header flex items-center justify-between p-4 bg-card border-b cursor-pointer hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    {uncollapsedGroups.has(group) ? (
                      <ChevronUp className="h-5 w-5 text-muted-foreground" />
                    ) : (
                      <ChevronDown className="h-5 w-5 text-muted-foreground" />
                    )}
                    <h2 className="text-xl font-semibold text-foreground">{group}</h2>
                  </div>
                  <div className="flex items-center gap-2">
                    {calculateUnhealthyCount(items) > 0 ? (
                      <span className="bg-red-600 text-white px-2 py-1 rounded-full text-sm font-medium">
                        {calculateUnhealthyCount(items)}
                      </span>
                    ) : (
                      <CheckCircle className="h-6 w-6 text-green-600" />
                    )}
                  </div>
                </div>

                {uncollapsedGroups.has(group) && (
                  <div className="endpoint-group-content p-4">
                    <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                      {items.map((endpoint) => (
                        <EndpointCard
                          key={endpoint.key}
                          endpoint={endpoint}
                          maxResults={resultPageSize}
                          showAverageResponseTime={showAverageResponseTime}
                          onShowTooltip={onShowTooltip}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <>
            <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
              {paginatedEndpoints.map((endpoint) => (
                <EndpointCard
                  key={endpoint.key}
                  endpoint={endpoint}
                  maxResults={resultPageSize}
                  showAverageResponseTime={showAverageResponseTime}
                  onShowTooltip={onShowTooltip}
                />
              ))}
            </div>

            {totalPages > 1 && (
              <div className="mt-8 flex items-center justify-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  disabled={currentPage === 1}
                  onClick={() => goToPage(currentPage - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>

                <div className="flex gap-1">
                  {visiblePages.map((page) => (
                    <Button
                      key={page}
                      variant={page === currentPage ? 'default' : 'outline'}
                      size="sm"
                      onClick={() => goToPage(page)}
                    >
                      {page}
                    </Button>
                  ))}
                </div>

                <Button
                  variant="outline"
                  size="icon"
                  disabled={currentPage === totalPages}
                  onClick={() => goToPage(currentPage + 1)}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            )}
          </>
        )}

        {archivedAnnouncements.length > 0 && (
          <div className="mt-12 pb-8">
            <PastAnnouncements announcements={archivedAnnouncements} />
          </div>
        )}
      </div>

      <Settings onRefreshData={fetchData} />
    </div>
  )
}
