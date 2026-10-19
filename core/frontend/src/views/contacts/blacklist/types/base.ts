// Enum values mirror internal/service/blacklist/types.go
export enum BlacklistScope {
	Email = 1,
	Domain = 2,
}

export enum BlacklistStatus {
	All = -1,
	Observing = 0,
	Active = 1,
	Released = 2,
}

export const BLACKLIST_REASONS = [
	'invalid_mailbox',
	'invalid_domain',
	'mailbox_disabled',
	'mailbox_full',
	'blocked',
	'complaint',
	'spam_trap',
	'disposable',
	'role_account',
	'manual',
	'legacy',
] as const

/** Reason id (1-based, as stored) -> i18n key */
export const reasonKey = (reason: number) => BLACKLIST_REASONS[reason - 1] ?? 'manual'

export const BLACKLIST_SOURCES = ['bounce', 'complaint', 'manual', 'import', 'api', 'legacy'] as const

export const sourceKey = (source: number) => BLACKLIST_SOURCES[source - 1] ?? 'manual'

export type BlacklistEntry = {
	id: number
	scope: BlacklistScope
	value: string
	domain: string
	reason: number
	source: number
	status: BlacklistStatus
	hit_count: number
	dsn: string
	last_response: string
	last_task_id: number
	expires_at: number
	note: string
	first_seen_at: number
	last_seen_at: number
	create_time: number
	update_time: number
}

export type BlacklistParams = {
	page: number
	page_size: number
	keyword: string
	scope: number
	reason: number
	source: number
	status: number
}

export type BlacklistStats = {
	active: number
	observing: number
	released: number
	by_reason: { reason: number; total: number }[]
}

export type BlacklistCreateParams = {
	scope: BlacklistScope
	value: string
	reason: number
	note: string
	ttl_days: number
}

export type BlacklistImportParams = Omit<BlacklistCreateParams, 'value'> & {
	values: string[]
}

export type BlacklistImportResult = {
	inserted: number
	updated: number
	invalid: { value: string; error: string }[]
}

export type BlacklistUpdateParams = {
	id: number
	reason: number
	note: string
	expire_mode: 'keep' | 'permanent' | 'until'
	expires_at: number
}

/** True when the entry is active and its block has not ended */
export const isBlocking = (entry: Pick<BlacklistEntry, 'status' | 'expires_at'>, nowSeconds: number) =>
	entry.status === BlacklistStatus.Active && (entry.expires_at === 0 || entry.expires_at > nowSeconds)

/** Split pasted or uploaded text into unique, trimmed values */
export const parseValues = (text: string) => {
	const seen = new Set<string>()
	return text
		.split(/[\s,;]+/)
		.map(value => value.trim().replace(/^["']|["']$/g, ''))
		.filter(value => {
			const key = value.toLowerCase()
			if (!value || seen.has(key)) return false
			seen.add(key)
			return true
		})
}
