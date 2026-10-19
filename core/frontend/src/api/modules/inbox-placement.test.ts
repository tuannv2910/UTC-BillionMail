import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()

vi.mock('@/api', () => ({
  instance: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
  },
}))

vi.mock('@/i18n', () => ({
  default: {
    global: { t: (key: string) => key },
  },
}))

import {
  getPlacementStatus,
  createPlacementTest,
  getPlacementTestList,
  getPlacementTestDetail,
  cancelPlacementTest,
  getSeedList,
  addSeed,
  getSeedOAuthUrl,
  updateSeed,
  checkSeed,
} from './inbox-placement'

describe('inbox-placement API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getPlacementStatus calls GET /inbox_placement/status', () => {
    getPlacementStatus()
    expect(mockGet).toHaveBeenCalledWith('/inbox_placement/status')
  })

  it('createPlacementTest calls POST /inbox_placement/create with loading + success message', () => {
    const params = {
      template_id: 3,
      addresser: 'news@abc.com',
      full_name: 'Shop',
      subject: 'Hi',
      seed_group: 'default',
      timeout_sec: 900,
    }
    createPlacementTest(params)
    expect(mockPost).toHaveBeenCalledWith(
      '/inbox_placement/create',
      params,
      expect.objectContaining({
        fetchOptions: expect.objectContaining({ successMessage: true, loading: expect.any(String) }),
      })
    )
  })

  it('getPlacementTestList calls GET /inbox_placement/list', () => {
    const params = { page: 1, page_size: 10, template_id: 3 }
    getPlacementTestList(params)
    expect(mockGet).toHaveBeenCalledWith('/inbox_placement/list', { params })
  })

  it('getPlacementTestDetail calls GET /inbox_placement/detail', () => {
    getPlacementTestDetail({ test_id: 't1' })
    expect(mockGet).toHaveBeenCalledWith('/inbox_placement/detail', { params: { test_id: 't1' } })
  })

  it('cancelPlacementTest calls POST /inbox_placement/cancel', () => {
    cancelPlacementTest({ test_id: 't1' })
    expect(mockPost).toHaveBeenCalledWith('/inbox_placement/cancel', { test_id: 't1' }, expect.any(Object))
  })

  it('seed endpoints map to /inbox_placement/seed_*', () => {
    getSeedList()
    expect(mockGet).toHaveBeenCalledWith('/inbox_placement/seed_list')

    addSeed({ email: 's@gmail.com', refresh_token: 'rt', seed_group: 'default' })
    expect(mockPost).toHaveBeenCalledWith(
      '/inbox_placement/seed_add',
      { email: 's@gmail.com', refresh_token: 'rt', seed_group: 'default' },
      expect.any(Object)
    )

    getSeedOAuthUrl({ seed_group: 'g1' })
    expect(mockPost).toHaveBeenCalledWith('/inbox_placement/seed_oauth', { seed_group: 'g1' })

    updateSeed({ id: 1, is_active: false })
    expect(mockPost).toHaveBeenCalledWith('/inbox_placement/seed_update', { id: 1, is_active: false }, expect.any(Object))

    checkSeed({ id: 1 })
    expect(mockPost).toHaveBeenCalledWith('/inbox_placement/seed_check', { id: 1 }, expect.any(Object))
  })
})
