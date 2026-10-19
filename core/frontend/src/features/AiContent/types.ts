export interface AiModel {
	title: string
	supplierName: string
	modelId: string
	max_tokens: number
	capability: string[]
	status: boolean
}

export type ReviewSeverity = 'low' | 'medium' | 'high' | 'critical'

export type ReviewVerdict = 'pass' | 'review' | 'fail'

export type ReviewLocation = 'subject' | 'body' | 'link' | 'general'

export type ReviewCategory =
	| 'spam_trigger'
	| 'deceptive_subject'
	| 'misleading_claim'
	| 'false_urgency'
	| 'prohibited_content'
	| 'phishing_risk'
	| 'link_risk'
	| 'compliance'
	| 'formatting'
	| 'other'

export interface ReviewIssue {
	category: ReviewCategory
	severity: ReviewSeverity
	location: ReviewLocation
	excerpt: string
	reason: string
	suggestion: string
}

export interface ReviewResult {
	score: number
	max_score: number
	verdict: ReviewVerdict
	summary: string
	issues: ReviewIssue[]
	dropped_issues: number
	model: string
	prompt_version: string
	input_tokens: number
	output_tokens: number
}

export interface GenerateResult {
	subject: string
	preheader: string
	html: string
	model: string
	input_tokens: number
	output_tokens: number
}

export interface AiModelParams {
	supplier_name: string
	model_id: string
}

export interface AiReviewParams extends AiModelParams {
	subject: string
	content: string
	language: string
	unsubscribe_managed: boolean
}

export interface AiGenerateParams extends AiModelParams {
	brief: string
	subject: string
	current_html: string
	tone: string
	language: string
}
