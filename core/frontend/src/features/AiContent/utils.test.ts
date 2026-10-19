import { describe, it, expect } from 'vitest'
import type { ReviewIssue } from './types'
import {
	buildFixInstruction,
	defaultAiLanguage,
	modelKey,
	modelParams,
	scoreColor,
	severityTagType,
	sortIssues,
	verdictTagType,
} from './utils'

const issue = (over: Partial<ReviewIssue>): ReviewIssue => ({
	category: 'spam_trigger',
	severity: 'low',
	location: 'body',
	excerpt: '',
	reason: 'r',
	suggestion: '',
	...over,
})

describe('AiContent utils', () => {
	it('round-trips model keys, including ids that contain colons', () => {
		const key = modelKey({ supplierName: 'OpenAI', modelId: 'ft:gpt:custom' })
		expect(key).toBe('OpenAI::ft:gpt:custom')
		expect(modelParams(key)).toEqual({ supplier_name: 'OpenAI', model_id: 'ft:gpt:custom' })
	})

	it('falls back to server default model for empty keys', () => {
		expect(modelParams(null)).toEqual({ supplier_name: '', model_id: '' })
		expect(modelParams('broken')).toEqual({ supplier_name: '', model_id: '' })
	})

	it('maps score, verdict and severity to colors', () => {
		expect(scoreColor(9)).toBe('var(--color-primary-1)')
		expect(scoreColor(5)).toBe('var(--color-warning-1)')
		expect(scoreColor(4.5)).toBe('var(--color-error-1)')
		expect(verdictTagType('pass')).toBe('success')
		expect(verdictTagType('review')).toBe('warning')
		expect(verdictTagType('fail')).toBe('error')
		expect(severityTagType('critical')).toBe('error')
		expect(severityTagType('medium')).toBe('warning')
		expect(severityTagType('low')).toBe('info')
	})

	it('sorts issues by severity and keeps order within a severity', () => {
		const sorted = sortIssues([
			issue({ reason: 'low-1' }),
			issue({ severity: 'critical', reason: 'crit' }),
			issue({ reason: 'low-2' }),
			issue({ severity: 'medium', reason: 'med' }),
		])
		expect(sorted.map(i => i.reason)).toEqual(['crit', 'med', 'low-1', 'low-2'])
	})

	it('builds a fix instruction listing each issue', () => {
		const text = buildFixInstruction([
			issue({ severity: 'medium', excerpt: 'ACT NOW', reason: 'urgency', suggestion: 'Shop when ready' }),
			issue({ severity: 'high', location: 'general', reason: 'no unsubscribe' }),
		])
		expect(text).toContain('1. [high] (general): no unsubscribe.')
		expect(text).toContain('2. [medium] "ACT NOW": urgency. Suggested fix: Shop when ready')
		expect(text).toContain('keep the layout')
	})

	it('picks a supported AI language from the UI locale', () => {
		expect(defaultAiLanguage('zh')).toBe('zh')
		expect(defaultAiLanguage('fr')).toBe('en')
	})
})
