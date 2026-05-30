import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ErrorState from './ErrorState'

describe('ErrorState', () => {
  it('renders default title when not provided', () => {
    render(<ErrorState message="Something went wrong" />)
    expect(screen.getByText('Error')).toBeInTheDocument()
  })

  it('renders custom title', () => {
    render(<ErrorState title="Connection Failed" message="Cannot connect" />)
    expect(screen.getByText('Connection Failed')).toBeInTheDocument()
  })

  it('renders error message', () => {
    render(<ErrorState message="Database not found" />)
    expect(screen.getByText('Database not found')).toBeInTheDocument()
  })

  it('renders detail in pre block when provided', () => {
    render(
      <ErrorState
        message="Error"
        detail="Stack trace: line 42"
      />
    )
    expect(screen.getByText('Stack trace: line 42')).toBeInTheDocument()
    expect(screen.getByText('Stack trace: line 42').tagName).toBe('PRE')
  })

  it('does not render detail when not provided', () => {
    const { container } = render(<ErrorState message="Error" />)
    expect(container.querySelector('pre')).toBeNull()
  })

  it('renders retry button when onRetry provided', () => {
    const onRetry = vi.fn()
    render(<ErrorState message="Error" onRetry={onRetry} />)
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  it('calls onRetry when retry button clicked', () => {
    const onRetry = vi.fn()
    render(<ErrorState message="Error" onRetry={onRetry} />)
    fireEvent.click(screen.getByText('Retry'))
    expect(onRetry).toHaveBeenCalledOnce()
  })

  it('does not render retry button when onRetry not provided', () => {
    render(<ErrorState message="Error" />)
    expect(screen.queryByText('Retry')).toBeNull()
  })
})
