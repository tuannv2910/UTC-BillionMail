import type { Placement, PlacementSummary, TestStatus } from './interface'

type TagType = 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'

interface PlacementMeta {
	color: string
	tag: TagType
}

// Colors follow the product palette: green = good, amber = tabs/missing, red = bad.
export const placementMeta: Record<Placement, PlacementMeta> = {
	inbox: { color: '#20a53a', tag: 'success' },
	promotions: { color: '#2d8cf0', tag: 'info' },
	social: { color: '#8e6cef', tag: 'info' },
	updates: { color: '#19a0b4', tag: 'info' },
	forums: { color: '#5c7cfa', tag: 'info' },
	spam: { color: '#ef0808', tag: 'error' },
	trash: { color: '#a8071a', tag: 'error' },
	other: { color: '#909399', tag: 'default' },
	missing: { color: '#f0ad4e', tag: 'warning' },
	error: { color: '#d03050', tag: 'error' },
	send_failed: { color: '#d03050', tag: 'error' },
	pending: { color: '#c0c4cc', tag: 'default' },
	sent: { color: '#c0c4cc', tag: 'default' },
}

// Segments of the summary bar, in display order.
export const summarySegments = [
	'inbox',
	'promotions',
	'social',
	'updates',
	'forums',
	'spam',
	'trash',
	'other',
	'missing',
	'error',
	'send_failed',
	'pending',
] as const

export type SummarySegment = (typeof summarySegments)[number]

export const testStatusTag: Record<TestStatus, TagType> = {
	pending: 'default',
	sending: 'info',
	checking: 'info',
	completed: 'success',
	failed: 'error',
	canceled: 'warning',
}

export const isRunning = (status: TestStatus) =>
	status === 'pending' || status === 'sending' || status === 'checking'

export const formatRate = (rate?: number) => {
	if (rate === undefined || rate === null || Number.isNaN(rate)) return '--'
	return `${Math.round(rate * 1000) / 10}%`
}

export const formatLatency = (ms?: number) => {
	if (ms === undefined || ms === null) return '--'
	if (ms < 1000) return `${ms}ms`
	const s = ms / 1000
	if (s < 60) return `${Math.round(s * 10) / 10}s`
	return `${Math.floor(s / 60)}m ${Math.round(s % 60)}s`
}

export interface BarSegment {
	key: SummarySegment
	count: number
	percent: number
	color: string
}

// Non-empty segments of a summary with their share of the total.
export const buildSegments = (summary?: PlacementSummary): BarSegment[] => {
	if (!summary || !summary.total) return []
	return summarySegments
		.map(key => ({
			key,
			count: summary[key] || 0,
			percent: ((summary[key] || 0) / summary.total) * 100,
			color: placementMeta[key].color,
		}))
		.filter(seg => seg.count > 0)
}

// Health verdict of an inbox rate, used for the headline color.
export const rateLevel = (rate: number): TagType => {
	if (rate >= 0.8) return 'success'
	if (rate >= 0.5) return 'warning'
	return 'error'
}

export const authTag = (verdict?: string): TagType => {
	if (!verdict) return 'default'
	if (verdict === 'pass') return 'success'
	if (verdict === 'none' || verdict === 'neutral') return 'warning'
	return 'error'
}
