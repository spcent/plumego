import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import RecentResources from './RecentResources'
import {
  RECENT_RESOURCES_KEY,
  buildRecentResourceDescriptor,
  clearRecentResources,
  reloadRecentResourcesFromStorage,
} from './recentResourcesModel'
import type { Connection } from '../api'

vi.mock('../i18nContext', () => ({
  useI18n: () => ({
    t: (key: string, vars?: Record<string, string | number>) => {
      if (key === 'redis.browser.db') return `DB ${vars?.n}`
      const labels: Record<string, string> = {
        'query.title': 'SQL Console',
        'resource.view_data': 'View Data',
        'resource.view_structure': 'View Structure',
        'resource.tables': 'Tables',
        'resource.collections': 'Collections',
        'mongodb.browser.title': 'Document Browser',
        'elasticsearch.index_name': 'Index',
        'recent.title': 'Recent',
        'recent.clear': 'Clear recent resources',
        'recent.redis_key': 'Redis Key',
        'recent.es_alias': 'Alias',
        'recent.es_data_stream': 'Data Stream',
      }
      return labels[key] ?? key
    },
  }),
}))

const connections: Connection[] = [
  {
    id: 'c1',
    name: 'Billing MySQL',
    driver: 'mysql',
  },
  {
    id: 'redis1',
    name: 'Cache Redis',
    driver: 'redis',
  },
]

describe('RecentResources', () => {
  beforeEach(() => {
    const values = new Map<string, string>()
    Object.defineProperty(globalThis, 'localStorage', {
      configurable: true,
      value: {
        getItem: vi.fn((key: string) => values.get(key) ?? null),
        setItem: vi.fn((key: string, value: string) => {
          values.set(key, value)
        }),
        removeItem: vi.fn((key: string) => {
          values.delete(key)
        }),
        clear: vi.fn(() => {
          values.clear()
        }),
      },
    })
    clearRecentResources()
  })

  it('builds a descriptor for a SQL table data route', () => {
    const item = buildRecentResourceDescriptor(
      '/conn/c1/db/billing/tables/invoices/data',
      connections,
      key => key,
    )

    expect(item).toMatchObject({
      label: 'invoices',
      detail: 'billing · resource.view_data · Billing MySQL',
      driver: 'mysql',
    })
  })

  it('tracks the current resource route in localStorage', async () => {
    render(
      <MemoryRouter initialEntries={['/conn/c1/db/billing/tables/invoices/data']}>
        <RecentResources connections={connections} />
      </MemoryRouter>,
    )

    expect(await screen.findByText('invoices')).toBeInTheDocument()
    expect(screen.getByText('billing · View Data · Billing MySQL')).toBeInTheDocument()

    await waitFor(() => {
      const raw = localStorage.getItem(RECENT_RESOURCES_KEY)
      expect(raw).toContain('/conn/c1/db/billing/tables/invoices/data')
    })
  })

  it('clears saved recent resources', async () => {
    localStorage.setItem(
      RECENT_RESOURCES_KEY,
      JSON.stringify([{ path: '/conn/redis1/redis/0?key=session%3A1', lastVisited: 1 }]),
    )
    reloadRecentResourcesFromStorage()

    render(
      <MemoryRouter initialEntries={['/connections']}>
        <RecentResources connections={connections} />
      </MemoryRouter>,
    )

    expect(await screen.findByText('session:1')).toBeInTheDocument()

    fireEvent.click(screen.getByTitle('Clear recent resources'))

    expect(localStorage.getItem(RECENT_RESOURCES_KEY)).toBeNull()
    expect(screen.queryByText('session:1')).toBeNull()
  })
})
