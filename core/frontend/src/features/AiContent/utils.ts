import type { AiModel, AiModelParams, ReviewIssue, ReviewSeverity, ReviewVerdict } from './types'

type TagType = 'default' | 'success' | 'info' | 'warning' | 'error'

const severityOrder: Record<ReviewSeverity, number> = { critical: 0, high: 1, medium: 2, low: 3 }

/**
 * @description Unique select value for a model (a model id may exist under several suppliers)
 */
export const modelKey = (model: Pick<AiModel, 'supplierName' | 'modelId'>) =>
	`${model.supplierName}::${model.modelId}`

/**
 * @description Convert a model select value to API params; empty lets the server pick the first enabled model
 */
export const modelParams = (key: string | null): AiModelParams => {
	const index = key ? key.indexOf('::') : -1
	if (!key || index < 0) return { supplier_name: '', model_id: '' }
	return { supplier_name: key.slice(0, index), model_id: key.slice(index + 2) }
}

/**
 * @description Color for a 0-10 score
 */
export const scoreColor = (score: number) => {
	if (score >= 8) return 'var(--color-primary-1)'
	if (score >= 5) return 'var(--color-warning-1)'
	return 'var(--color-error-1)'
}

export const verdictTagType = (verdict: ReviewVerdict): TagType => {
	if (verdict === 'pass') return 'success'
	if (verdict === 'review') return 'warning'
	return 'error'
}

export const severityTagType = (severity: ReviewSeverity): TagType => {
	if (severity === 'critical' || severity === 'high') return 'error'
	if (severity === 'medium') return 'warning'
	return 'info'
}

/**
 * @description Most severe issues first, keeping the model's order within a severity
 */
export const sortIssues = (issues: ReviewIssue[]) =>
	issues
		.map((issue, index) => ({ issue, index }))
		.sort((a, b) => severityOrder[a.issue.severity] - severityOrder[b.issue.severity] || a.index - b.index)
		.map(item => item.issue)

/**
 * @description Instruction for the generator to fix the issues a review found
 */
export const buildFixInstruction = (issues: ReviewIssue[]) => {
	const lines = sortIssues(issues).map((issue, i) => {
		const where = issue.excerpt ? `"${issue.excerpt}"` : `(${issue.location})`
		const fix = issue.suggestion ? ` Suggested fix: ${issue.suggestion}` : ''
		return `${i + 1}. [${issue.severity}] ${where}: ${issue.reason}.${fix}`
	})
	return [
		'Fix every deliverability and compliance issue listed below. Change only what is needed to fix them;',
		'keep the layout, branding, links, images and template variables.',
		'',
		...lines,
	].join('\n')
}

/**
 * @description Map the UI locale to the AI output language
 */
export const defaultAiLanguage = (locale: string) => {
	return ['en', 'zh', 'ja', 'vi'].includes(locale) ? locale : 'en'
}
