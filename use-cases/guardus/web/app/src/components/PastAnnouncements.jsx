import React, { useState, useMemo } from 'react'
import { XCircle, AlertTriangle, Info, CheckCircle, Circle, ChevronDown } from 'lucide-react'
import { formatAnnouncementMessage } from '@/utils/markdown'

const typeConfigs = {
  outage: {
    icon: XCircle,
    background: 'bg-red-50 dark:bg-red-900/20',
    borderColor: 'border-red-500 dark:border-red-400',
    iconColor: 'text-red-600 dark:text-red-400',
    text: 'text-red-700 dark:text-red-300',
  },
  warning: {
    icon: AlertTriangle,
    background: 'bg-yellow-50 dark:bg-yellow-900/20',
    borderColor: 'border-yellow-500 dark:border-yellow-400',
    iconColor: 'text-yellow-600 dark:text-yellow-400',
    text: 'text-yellow-700 dark:text-yellow-300',
  },
  information: {
    icon: Info,
    background: 'bg-blue-50 dark:bg-blue-900/20',
    borderColor: 'border-blue-500 dark:border-blue-400',
    iconColor: 'text-blue-600 dark:text-blue-400',
    text: 'text-blue-700 dark:text-blue-300',
  },
  operational: {
    icon: CheckCircle,
    background: 'bg-green-50 dark:bg-green-900/20',
    borderColor: 'border-green-500 dark:border-green-400',
    iconColor: 'text-green-600 dark:text-green-400',
    text: 'text-green-700 dark:text-green-300',
  },
  none: {
    icon: Circle,
    background: 'bg-gray-50 dark:bg-gray-800/20',
    borderColor: 'border-gray-500 dark:border-gray-400',
    iconColor: 'text-gray-600 dark:text-gray-400',
    text: 'text-gray-700 dark:text-gray-300',
  },
}

const normalizeDate = (date) => {
  const normalized = new Date(date)
  normalized.setHours(0, 0, 0, 0)
  return normalized
}

const getTypeIcon = (type) => typeConfigs[type]?.icon || Circle
const getTypeClasses = (type) => typeConfigs[type] || typeConfigs.none

const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

const formatTime = (timestamp) => {
  return new Date(timestamp).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

const formatFullTimestamp = (timestamp) => {
  return new Date(timestamp).toLocaleString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    timeZoneName: 'short',
  })
}

export default function PastAnnouncements({ announcements = [] }) {
  const [showAllAnnouncements, setShowAllAnnouncements] = useState(false)

  const displayedAnnouncements = useMemo(() => {
    if (!announcements.length) return {}

    const grouped = {}
    let oldest = new Date()

    announcements.forEach((announcement) => {
      const date = new Date(announcement.timestamp)
      const key = date.toDateString()
      grouped[key] = grouped[key] || []
      grouped[key].push(announcement)
      if (date < oldest) oldest = date
    })

    const today = normalizeDate(new Date())
    const endDate = showAllAnnouncements
      ? normalizeDate(oldest)
      : new Date(today.getTime() - 14 * 24 * 60 * 60 * 1000)

    const result = {}
    const todayKey = today.toDateString()
    if (grouped[todayKey]) result[todayKey] = grouped[todayKey]

    for (
      let date = new Date(today.getTime() - 24 * 60 * 60 * 1000);
      date >= endDate;
      date.setDate(date.getDate() - 1)
    ) {
      result[date.toDateString()] = grouped[date.toDateString()] || []
    }

    return result
  }, [announcements, showAllAnnouncements])

  const hasOlderAnnouncements = useMemo(() => {
    if (!announcements.length) return false
    const fourteenDaysAgo = new Date(
      normalizeDate(new Date()).getTime() - 14 * 24 * 60 * 60 * 1000
    )
    return announcements.some((a) => new Date(a.timestamp) < fourteenDaysAgo)
  }, [announcements])

  if (!announcements.length) return null

  return (
    <div>
      <h2 className="text-2xl font-semibold text-foreground mb-6">Past Announcements</h2>

      <div className="space-y-8">
        {Object.entries(displayedAnnouncements).map(([date, group]) => (
          <div key={date}>
            <div className="mb-3">
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                {formatDate(date)}
              </h3>
            </div>

            {group.length > 0 ? (
              <div className="space-y-3">
                {group.map((announcement, index) => {
                  const Icon = getTypeIcon(announcement.type)
                  const typeClasses = getTypeClasses(announcement.type)
                  return (
                    <div
                      key={`${date}-${index}-${announcement.timestamp}`}
                      className={`border-l-4 p-4 transition-all duration-200 ${typeClasses.background} ${typeClasses.borderColor}`}
                    >
                      <div className="flex items-start gap-3">
                        {React.createElement(Icon, {
                          className: `w-5 h-5 flex-shrink-0 mt-0.5 ${typeClasses.iconColor}`,
                        })}
                        <time
                          className={`text-sm font-mono whitespace-nowrap flex-shrink-0 mt-0.5 ${typeClasses.text}`}
                          title={formatFullTimestamp(announcement.timestamp)}
                        >
                          {formatTime(announcement.timestamp)}
                        </time>
                        <div className="flex-1 min-w-0">
                          <p
                            className="text-sm leading-relaxed text-gray-900 dark:text-gray-100"
                            dangerouslySetInnerHTML={{
                              __html: formatAnnouncementMessage(announcement.message),
                            }}
                          ></p>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="py-2">
                <p className="text-sm italic text-muted-foreground/60">
                  No incidents reported on this day
                </p>
              </div>
            )}
          </div>
        ))}

        {hasOlderAnnouncements && !showAllAnnouncements && (
          <div>
            <button
              onClick={() => setShowAllAnnouncements(true)}
              className="inline-flex items-center gap-2 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors duration-200 cursor-pointer group"
            >
              <ChevronDown className="w-4 h-4 group-hover:translate-y-0.5 transition-transform duration-200" />
              <span className="group-hover:underline">View older announcements</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
