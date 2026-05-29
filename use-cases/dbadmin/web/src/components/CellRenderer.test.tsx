import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import CellRenderer from './CellRenderer'

describe('CellRenderer', () => {
  it('renders NULL for null value', () => {
    render(<CellRenderer value={null} />)
    expect(screen.getByText('NULL')).toBeInTheDocument()
  })

  it('renders NULL for undefined value', () => {
    render(<CellRenderer value={undefined} />)
    expect(screen.getByText('NULL')).toBeInTheDocument()
  })

  it('renders (empty) for empty string', () => {
    render(<CellRenderer value="" />)
    expect(screen.getByText('(empty)')).toBeInTheDocument()
  })

  it('renders BLOB badge for BLOB values', () => {
    render(<CellRenderer value="<BLOB 1024 bytes>" />)
    expect(screen.getByText('BLOB')).toBeInTheDocument()
    expect(screen.getByText('1024 bytes')).toBeInTheDocument()
  })

  it('renders JSON badge for object-like strings', () => {
    render(<CellRenderer value='{"key":"value"}' />)
    expect(screen.getByText('JSON')).toBeInTheDocument()
  })

  it('renders JSON badge for array-like strings', () => {
    render(<CellRenderer value='[1,2,3]' />)
    expect(screen.getByText('JSON')).toBeInTheDocument()
  })

  it('does not render JSON badge for plain strings', () => {
    render(<CellRenderer value="hello world" />)
    expect(screen.queryByText('JSON')).not.toBeInTheDocument()
    expect(screen.getByText('hello world')).toBeInTheDocument()
  })

  it('truncates long JSON preview to 120 chars with ellipsis', () => {
    const longJson = '{"key":"' + 'x'.repeat(200) + '"}'
    const { container } = render(<CellRenderer value={longJson} />)
    const text = container.textContent ?? ''
    expect(text).toContain('…')
    // The JSON display should be capped
    const jsonSpan = container.querySelector('span.font-mono.text-\\[11px\\].truncate')
    expect(jsonSpan?.textContent?.length ?? 0).toBeLessThanOrEqual(125)
  })

  it('renders numeric values right-aligned for numeric column types', () => {
    const { container } = render(<CellRenderer value={42} colType="int" />)
    expect(container.querySelector('.tabular-nums')).toBeInTheDocument()
  })

  it('renders numeric values right-aligned based on value type', () => {
    const { container } = render(<CellRenderer value={3.14} />)
    expect(container.querySelector('.tabular-nums')).toBeInTheDocument()
  })

  it('truncates plain text longer than 120 chars', () => {
    const long = 'a'.repeat(150)
    const { container } = render(<CellRenderer value={long} />)
    expect(container.textContent).toContain('…')
    expect(container.textContent?.includes('a'.repeat(150))).toBe(false)
  })

  it('renders plain text under 120 chars without truncation', () => {
    const text = 'short text'
    render(<CellRenderer value={text} />)
    expect(screen.getByText(text)).toBeInTheDocument()
  })

  it('renders string numbers as plain text without colType hint', () => {
    render(<CellRenderer value="hello" colType="varchar(255)" />)
    expect(screen.getByText('hello')).toBeInTheDocument()
  })

  it('does not confuse objects that start with { but do not end with }', () => {
    render(<CellRenderer value="{not closed" />)
    expect(screen.queryByText('JSON')).not.toBeInTheDocument()
  })
})
