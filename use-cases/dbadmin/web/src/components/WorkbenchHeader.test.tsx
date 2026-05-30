import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import WorkbenchHeader from './WorkbenchHeader'

// Mock useI18n hook
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

describe('WorkbenchHeader', () => {
  it('renders connection name when provided', () => {
    render(
      <WorkbenchHeader
        connectionName="production-db"
        resourcePath={['mydb']}
        datasourceType="mysql"
      />
    )
    expect(screen.getByText('production-db')).toBeInTheDocument()
  })

  it('renders resource path', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb', 'users']}
        datasourceType="mysql"
      />
    )
    expect(screen.getByText('mydb')).toBeInTheDocument()
    expect(screen.getByText('users')).toBeInTheDocument()
  })

  it('renders datasource type badge', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
      />
    )
    expect(screen.getByText('MY')).toBeInTheDocument()
  })

  it('renders readonly badge when readonly is true', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
        readonly={true}
      />
    )
    expect(screen.getByText('workbench.readonly')).toBeInTheDocument()
  })

  it('does not render readonly badge when readonly is false', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
        readonly={false}
      />
    )
    expect(screen.queryByText('workbench.readonly')).toBeNull()
  })

  it('renders refresh button when onRefresh provided', () => {
    const onRefresh = vi.fn()
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
        onRefresh={onRefresh}
      />
    )
    const refreshButton = screen.getByTitle('workbench.refresh')
    expect(refreshButton).toBeInTheDocument()
  })

  it('calls onRefresh when refresh button clicked', () => {
    const onRefresh = vi.fn()
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
        onRefresh={onRefresh}
      />
    )
    const refreshButton = screen.getByTitle('workbench.refresh')
    fireEvent.click(refreshButton)
    expect(onRefresh).toHaveBeenCalledOnce()
  })

  it('does not render refresh button when onRefresh not provided', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
      />
    )
    expect(screen.queryByText('workbench.refresh')).toBeNull()
  })

  it('renders meta information when provided', () => {
    render(
      <WorkbenchHeader
        resourcePath={['mydb']}
        datasourceType="mysql"
        meta={{ rowCount: 1500, docsCount: 2000, keyCount: 300 }}
      />
    )
    expect(screen.getByText('1500 rows')).toBeInTheDocument()
    expect(screen.getByText('2000 docs')).toBeInTheDocument()
    expect(screen.getByText('300 keys')).toBeInTheDocument()
  })
})
