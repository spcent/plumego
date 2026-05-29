import React, { useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function Pagination({ numberOfResultsPerPage, currentPageProp = 1, onPageChange }) {
  const [currentPage, setCurrentPage] = useState(currentPageProp)

  const maxPages = (() => {
    let maxResults = 100
    if (typeof window !== 'undefined' && window.config && window.config.maximumNumberOfResults) {
      const parsed = parseInt(window.config.maximumNumberOfResults)
      if (!isNaN(parsed)) {
        maxResults = parsed
      }
    }
    return Math.ceil(maxResults / numberOfResultsPerPage)
  })()

  const nextPage = () => {
    const newPage = currentPage - 1
    setCurrentPage(newPage)
    onPageChange?.(newPage)
  }

  const previousPage = () => {
    const newPage = currentPage + 1
    setCurrentPage(newPage)
    onPageChange?.(newPage)
  }

  return (
    <div className="flex items-center justify-between">
      <Button
        variant="outline"
        size="sm"
        disabled={currentPage >= maxPages}
        onClick={previousPage}
        className="flex items-center gap-1"
      >
        <ChevronLeft className="h-4 w-4" />
        Previous
      </Button>

      <span className="text-sm text-muted-foreground">
        Page {currentPage} of {maxPages}
      </span>

      <Button
        variant="outline"
        size="sm"
        disabled={currentPage <= 1}
        onClick={nextPage}
        className="flex items-center gap-1"
      >
        Next
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  )
}
