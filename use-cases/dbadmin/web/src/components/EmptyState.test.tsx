import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import EmptyState from './EmptyState'

describe('EmptyState', () => {
  it('renders message', () => {
    render(<EmptyState message="No data available" />)
    expect(screen.getByText('No data available')).toBeInTheDocument()
  })

  it('renders hint when provided', () => {
    render(
      <EmptyState
        message="No results"
        hint="Try adjusting your filters"
      />
    )
    expect(screen.getByText('Try adjusting your filters')).toBeInTheDocument()
  })

  it('does not render hint when not provided', () => {
    const { container } = render(<EmptyState message="Empty" />)
    const paragraphs = container.querySelectorAll('p')
    expect(paragraphs).toHaveLength(1) // Only message, no hint
  })

  it('renders action when provided', () => {
    render(
      <EmptyState
        message="No connections"
        action={<button>Add Connection</button>}
      />
    )
    expect(screen.getByText('Add Connection')).toBeInTheDocument()
  })

  it('does not render action when not provided', () => {
    const { container } = render(<EmptyState message="Empty" />)
    expect(container.querySelector('button')).toBeNull()
  })

  it('renders empty state icon', () => {
    const { container } = render(<EmptyState message="Empty" />)
    expect(container.querySelector('svg')).not.toBeNull()
  })
})
