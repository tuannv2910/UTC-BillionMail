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
	createBlacklist,
	deleteBlacklist,
	getBlacklist,
	getBlacklistStats,
	importBlacklist,
	releaseBlacklist,
	updateBlacklist,
} from './blacklist'

const withSuccessMessage = expect.objectContaining({
	fetchOptions: expect.objectContaining({ successMessage: true }),
})

describe('contacts/blacklist API', () => {
	beforeEach(() => {
		vi.clearAllMocks()
	})

	it('getBlacklist calls GET /email_blacklist/list', () => {
		const params = { page: 1, page_size: 10, keyword: '', scope: 0, reason: 0, source: 0, status: -1 }
		getBlacklist(params)
		expect(mockGet).toHaveBeenCalledWith('/email_blacklist/list', { params })
	})

	it('getBlacklistStats calls GET /email_blacklist/stats', () => {
		getBlacklistStats()
		expect(mockGet).toHaveBeenCalledWith('/email_blacklist/stats')
	})

	it('createBlacklist calls POST /email_blacklist/create', () => {
		const params = { scope: 1, value: 'a@x.com', reason: 10, note: '', ttl_days: -1 }
		createBlacklist(params)
		expect(mockPost).toHaveBeenCalledWith('/email_blacklist/create', params, withSuccessMessage)
	})

	it('importBlacklist calls POST /email_blacklist/import', () => {
		const params = { scope: 2, values: ['x.com'], reason: 8, note: '', ttl_days: 0 }
		importBlacklist(params)
		expect(mockPost).toHaveBeenCalledWith('/email_blacklist/import', params, withSuccessMessage)
	})

	it('updateBlacklist calls POST /email_blacklist/update', () => {
		const params = { id: 3, reason: 1, note: 'n', expire_mode: 'keep' as const, expires_at: 0 }
		updateBlacklist(params)
		expect(mockPost).toHaveBeenCalledWith('/email_blacklist/update', params, withSuccessMessage)
	})

	it('releaseBlacklist calls POST /email_blacklist/release', () => {
		releaseBlacklist({ ids: [1, 2] })
		expect(mockPost).toHaveBeenCalledWith('/email_blacklist/release', { ids: [1, 2] }, withSuccessMessage)
	})

	it('deleteBlacklist calls POST /email_blacklist/delete', () => {
		deleteBlacklist({ ids: [4] })
		expect(mockPost).toHaveBeenCalledWith('/email_blacklist/delete', { ids: [4] }, withSuccessMessage)
	})
})
