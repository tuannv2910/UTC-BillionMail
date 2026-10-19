import { describe, it, expect } from 'vitest'
import {
	authTag,
	buildSegments,
	formatLatency,
	formatRate,
	isRunning,
	rateLevel,
} from './placement'
import type { PlacementSummary } from './interface'

const summary = (patch: Partial<PlacementSummary>): PlacementSummary => ({
	total: 0,
	pending: 0,
	inbox: 0,
	promotions: 0,
	social: 0,
	updates: 0,
	forums: 0,
	spam: 0,
	trash: 0,
	other: 0,
	missing: 0,
	error: 0,
	send_failed: 0,
	delivered: 0,
	inbox_rate: 0,
	inbox_any_rate: 0,
	spam_rate: 0,
	...patch,
})

describe('inbox placement helpers', () => {
	it('formats rates as percentages', () => {
		expect(formatRate(0.6)).toBe('60%')
		expect(formatRate(0.3333)).toBe('33.3%')
		expect(formatRate(undefined)).toBe('--')
	})

	it('formats latency', () => {
		expect(formatLatency(800)).toBe('800ms')
		expect(formatLatency(7012)).toBe('7s')
		expect(formatLatency(95000)).toBe('1m 35s')
		expect(formatLatency(undefined)).toBe('--')
	})

	it('builds only non-empty segments in order', () => {
		const segs = buildSegments(summary({ total: 4, inbox: 2, spam: 1, pending: 1 }))
		expect(segs.map(s => s.key)).toEqual(['inbox', 'spam', 'pending'])
		expect(segs[0].percent).toBe(50)
		expect(buildSegments(summary({ total: 0 }))).toEqual([])
		expect(buildSegments(undefined)).toEqual([])
	})

	it('detects running tests', () => {
		expect(isRunning('checking')).toBe(true)
		expect(isRunning('completed')).toBe(false)
		expect(isRunning('canceled')).toBe(false)
	})

	it('grades rates and auth verdicts', () => {
		expect(rateLevel(0.9)).toBe('success')
		expect(rateLevel(0.6)).toBe('warning')
		expect(rateLevel(0.1)).toBe('error')
		expect(authTag('pass')).toBe('success')
		expect(authTag('fail')).toBe('error')
		expect(authTag('none')).toBe('warning')
		expect(authTag(undefined)).toBe('default')
	})
})
