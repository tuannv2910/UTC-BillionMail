import { instance } from '@/api'
import i18n from '@/i18n'
import type {
	BlacklistCreateParams,
	BlacklistImportParams,
	BlacklistParams,
	BlacklistUpdateParams,
} from '@/views/contacts/blacklist/types/base'

const { t } = i18n.global

export const getBlacklist = (params: BlacklistParams) => {
	return instance.get('/email_blacklist/list', { params })
}

export const getBlacklistStats = () => {
	return instance.get('/email_blacklist/stats')
}

export const createBlacklist = (params: BlacklistCreateParams) => {
	return instance.post('/email_blacklist/create', params, {
		fetchOptions: {
			loading: t('contacts.blacklist.api.loading.saving'),
			successMessage: true,
		},
	})
}

export const importBlacklist = (params: BlacklistImportParams) => {
	return instance.post('/email_blacklist/import', params, {
		fetchOptions: {
			loading: t('contacts.blacklist.api.loading.importing'),
			successMessage: true,
		},
	})
}

export const updateBlacklist = (params: BlacklistUpdateParams) => {
	return instance.post('/email_blacklist/update', params, {
		fetchOptions: {
			loading: t('contacts.blacklist.api.loading.saving'),
			successMessage: true,
		},
	})
}

export const releaseBlacklist = (params: { ids: number[] }) => {
	return instance.post('/email_blacklist/release', params, {
		fetchOptions: {
			loading: t('contacts.blacklist.api.loading.releasing'),
			successMessage: true,
		},
	})
}

export const deleteBlacklist = (params: { ids: number[] }) => {
	return instance.post('/email_blacklist/delete', params, {
		fetchOptions: {
			loading: t('contacts.blacklist.api.loading.deleting'),
			successMessage: true,
		},
	})
}
