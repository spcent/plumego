import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ConfirmDialog from './ConfirmDialog'

describe('ConfirmDialog', () => {
  it('renders nothing when closed', () => {
    const { container } = render(
      <ConfirmDialog
        open={false}
        title="Test"
        message="Test message"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders title and message when open', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Confirm Action"
        message="Are you sure?"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )
    expect(screen.getByText('Confirm Action')).toBeInTheDocument()
    expect(screen.getByText('Are you sure?')).toBeInTheDocument()
  })

  it('calls onConfirm when confirm button clicked', () => {
    const onConfirm = vi.fn()
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />
    )
    fireEvent.click(screen.getByText('Confirm'))
    expect(onConfirm).toHaveBeenCalledOnce()
  })

  it('calls onCancel when cancel button clicked', () => {
    const onCancel = vi.fn()
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />
    )
    fireEvent.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('calls onCancel when backdrop clicked', () => {
    const onCancel = vi.fn()
    const { container } = render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />
    )
    // Click the backdrop (outer div)
    fireEvent.click(container.firstChild!)
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('calls onCancel when Escape key pressed', () => {
    const onCancel = vi.fn()
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />
    )
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onCancel).toHaveBeenCalledOnce()
  })

  it('uses custom confirm label', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        confirmLabel="Delete"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )
    expect(screen.getByText('Delete')).toBeInTheDocument()
  })

  it('applies dangerous styling when dangerous=true', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        dangerous={true}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )
    const confirmButton = screen.getByText('Confirm')
    expect(confirmButton.className).toContain('btn-danger')
  })

  it('applies normal styling when dangerous=false', () => {
    render(
      <ConfirmDialog
        open={true}
        title="Test"
        message="Test"
        dangerous={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )
    const confirmButton = screen.getByText('Confirm')
    expect(confirmButton.className).toContain('btn-primary')
  })
})
