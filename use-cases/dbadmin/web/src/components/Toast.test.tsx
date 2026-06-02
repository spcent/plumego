import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { ToastProvider } from './Toast'
import { useToast } from './toastContext'

// Test component that uses the toast hook
function TestComponent({ message, type = 'info' }: { message: string; type?: 'info' | 'success' | 'error' }) {
  const { showToast } = useToast()

  return (
    <button onClick={() => showToast(message, type)}>
      Show Toast
    </button>
  )
}

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders toast when shown', () => {
    render(
      <ToastProvider>
        <TestComponent message="Operation successful" type="success" />
      </ToastProvider>
    )

    const button = screen.getByText('Show Toast')
    act(() => {
      button.click()
    })

    expect(screen.getByText('Operation successful')).toBeInTheDocument()
  })

  it('auto-dismisses toast after 4.5 seconds', () => {
    render(
      <ToastProvider>
        <TestComponent message="Temporary message" />
      </ToastProvider>
    )

    const button = screen.getByText('Show Toast')
    act(() => {
      button.click()
    })

    expect(screen.getByText('Temporary message')).toBeInTheDocument()

    act(() => {
      vi.advanceTimersByTime(4600)
    })

    expect(screen.queryByText('Temporary message')).not.toBeInTheDocument()
  })

  it('renders info toast with correct styling', () => {
    render(
      <ToastProvider>
        <TestComponent message="Info message" type="info" />
      </ToastProvider>
    )

    const button = screen.getByText('Show Toast')
    act(() => {
      button.click()
    })

    const toast = screen.getByText('Info message').closest('[role="status"]')
    expect(toast).toBeInTheDocument()
  })

  it('renders success toast with correct styling', () => {
    render(
      <ToastProvider>
        <TestComponent message="Success message" type="success" />
      </ToastProvider>
    )

    const button = screen.getByText('Show Toast')
    act(() => {
      button.click()
    })

    const toast = screen.getByText('Success message').closest('[role="status"]')
    expect(toast).toBeInTheDocument()
  })

  it('renders error toast with correct styling', () => {
    render(
      <ToastProvider>
        <TestComponent message="Error message" type="error" />
      </ToastProvider>
    )

    const button = screen.getByText('Show Toast')
    act(() => {
      button.click()
    })

    const toast = screen.getByText('Error message').closest('[role="status"]')
    expect(toast).toBeInTheDocument()
  })
})
