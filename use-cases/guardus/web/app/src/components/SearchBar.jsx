import React, { useEffect, useState } from 'react'
import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const filterOptions = [
  { label: 'None', value: 'none' },
  { label: 'Failing', value: 'failing' },
  { label: 'Unstable', value: 'unstable' },
]

const sortOptions = [
  { label: 'Name', value: 'name' },
  { label: 'Group', value: 'group' },
  { label: 'Health', value: 'health' },
]

export default function SearchBar({
  onSearch,
  onShowOnlyFailingChange,
  onShowRecentFailuresChange,
  onGroupByGroupChange,
  onSortByChange,
  onInitializeCollapsedGroups,
}) {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterBy, setFilterBy] = useState(
    localStorage.getItem('gatus:filter-by') ||
      (typeof window !== 'undefined' && window.config?.defaultFilterBy) ||
      'none'
  )
  const [sortBy, setSortBy] = useState(
    localStorage.getItem('gatus:sort-by') ||
      (typeof window !== 'undefined' && window.config?.defaultSortBy) ||
      'name'
  )

  const handleFilterChange = (value, store = true) => {
    setFilterBy(value)
    if (store) localStorage.setItem('gatus:filter-by', value)

    onShowOnlyFailingChange?.(false)
    onShowRecentFailuresChange?.(false)

    if (value === 'failing') {
      onShowOnlyFailingChange?.(true)
    } else if (value === 'unstable') {
      onShowRecentFailuresChange?.(true)
    }
  }

  const handleSortChange = (value, store = true) => {
    setSortBy(value)
    if (store) localStorage.setItem('gatus:sort-by', value)

    onSortByChange?.(value)
    onGroupByGroupChange?.(value === 'group')

    if (value === 'group') {
      onInitializeCollapsedGroups?.()
    }
  }

  useEffect(() => {
    handleFilterChange(filterBy, false)
    handleSortChange(sortBy, false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSearchInput = (e) => {
    const value = e.target.value
    setSearchQuery(value)
    onSearch?.(value)
  }

  return (
    <div className="flex flex-col lg:flex-row gap-3 lg:gap-4 p-3 sm:p-4 bg-card rounded-lg border">
      <div className="flex-1">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <label htmlFor="search-input" className="sr-only">
            Search endpoints
          </label>
          <Input
            id="search-input"
            type="text"
            value={searchQuery}
            onChange={handleSearchInput}
            placeholder="Search endpoints..."
            className="pl-10 text-sm sm:text-base"
          />
        </div>
      </div>
      <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
        <div className="flex items-center gap-2 flex-1 sm:flex-initial">
          <label className="text-xs sm:text-sm font-medium text-muted-foreground whitespace-nowrap">
            Filter by:
          </label>
          <Select value={filterBy} onValueChange={(v) => handleFilterChange(v, true)}>
            <SelectTrigger className="flex-1 sm:w-[140px] md:w-[160px]">
              <SelectValue placeholder="None" />
            </SelectTrigger>
            <SelectContent>
              {filterOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-2 flex-1 sm:flex-initial">
          <label className="text-xs sm:text-sm font-medium text-muted-foreground whitespace-nowrap">
            Sort by:
          </label>
          <Select value={sortBy} onValueChange={(v) => handleSortChange(v, true)}>
            <SelectTrigger className="flex-1 sm:w-[90px] md:w-[100px]">
              <SelectValue placeholder="Name" />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  )
}
