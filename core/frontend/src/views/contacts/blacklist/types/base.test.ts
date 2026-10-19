import { describe, it, expect } from 'vitest'
import { BlacklistStatus, isBlocking, parseValues, reasonKey, sourceKey } from './base'

describe('blacklist types', () => {
	it('maps stored ids to i18n keys', () => {
		expect(reasonKey(1)).toBe('invalid_mailbox')
		expect(reasonKey(6)).toBe('complaint')
		expect(reasonKey(11)).toBe('legacy')
		expect(reasonKey(99)).toBe('manual')
		expect(sourceKey(1)).toBe('bounce')
		expect(sourceKey(4)).toBe('import')
	})

	it('tells whether an entry still blocks sending', () => {
		const now = 1000
		expect(isBlocking({ status: BlacklistStatus.Active, expires_at: 0 }, now)).toBe(true)
		expect(isBlocking({ status: BlacklistStatus.Active, expires_at: 1001 }, now)).toBe(true)
		expect(isBlocking({ status: BlacklistStatus.Active, expires_at: 1000 }, now)).toBe(false)
		expect(isBlocking({ status: BlacklistStatus.Released, expires_at: 0 }, now)).toBe(false)
		expect(isBlocking({ status: BlacklistStatus.Observing, expires_at: 0 }, now)).toBe(false)
	})

	it('parses values separated by newlines, commas, semicolons and spaces', () => {
		const text = 'a@x.com\n"B@x.com", c@x.com;  a@x.com\r\n\n\'d@x.com\''
		expect(parseValues(text)).toEqual(['a@x.com', 'B@x.com', 'c@x.com', 'd@x.com'])
		expect(parseValues('  \n ')).toEqual([])
	})
})
