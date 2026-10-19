export type TestStatus = 'pending' | 'sending' | 'checking' | 'completed' | 'failed' | 'canceled'

export type Placement =
	| 'pending'
	| 'sent'
	| 'send_failed'
	| 'inbox'
	| 'promotions'
	| 'social'
	| 'updates'
	| 'forums'
	| 'spam'
	| 'trash'
	| 'other'
	| 'missing'
	| 'error'

export interface PlacementSummary {
	total: number
	pending: number
	inbox: number
	promotions: number
	social: number
	updates: number
	forums: number
	spam: number
	trash: number
	other: number
	missing: number
	error: number
	send_failed: number
	delivered: number
	inbox_rate: number
	inbox_any_rate: number
	spam_rate: number
}

export interface SeedResult {
	seed: string
	provider: string
	status: Placement
	sent_at: number
	received_at: number
	latency_ms?: number
	spf?: string
	dkim?: string
	dmarc?: string
	labels?: string[]
	error?: string
}

export interface PlacementTest {
	test_id: string
	domain: string
	from: string
	subject: string
	status: TestStatus
	seed_group: string
	reference?: string
	template_id: number
	timeout_sec: number
	error?: string
	created_at: number
	sent_at: number
	completed_at: number
	summary: PlacementSummary
	results: SeedResult[]
}

export interface PlacementTestParams {
	page: number
	page_size: number
	template_id?: number
	status?: string
}

export interface CreateTestParams {
	template_id: number
	addresser: string
	full_name: string
	subject: string
	seed_group: string
	timeout_sec: number
}

export interface Seed {
	id: number
	provider: string
	email: string
	seed_group: string
	addressing: string
	oauth_client?: string
	is_active: boolean
	health_status: 'ok' | 'auth_revoked' | 'error' | 'unknown'
	health_fail_count: number
	last_health_at: number
	last_error?: string
	created_at: number
}

export interface SeedAddParams {
	email: string
	refresh_token: string
	seed_group: string
}

export interface SeedUpdateParams {
	id: number
	is_active?: boolean
	seed_group?: string
}

export interface PlacementServiceStatus {
	enabled: boolean
	reachable: boolean
	admin_seed: boolean
}
