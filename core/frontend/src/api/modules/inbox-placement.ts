import i18n from '@/i18n'
import { instance } from '@/api'
import type {
	CreateTestParams,
	PlacementTestParams,
	SeedAddParams,
	SeedUpdateParams,
} from '@/views/template/inbox-placement/interface'

const { t } = i18n.global

export function getPlacementStatus() {
	return instance.get('/inbox_placement/status')
}

export function createPlacementTest(params: CreateTestParams) {
	return instance.post('/inbox_placement/create', params, {
		fetchOptions: {
			loading: t('template.inboxPlacement.loading.creating'),
			successMessage: true,
		},
	})
}

export function getPlacementTestList(params: PlacementTestParams) {
	return instance.get('/inbox_placement/list', { params })
}

export function getPlacementTestDetail(params: { test_id: string }) {
	return instance.get('/inbox_placement/detail', { params })
}

export function cancelPlacementTest(params: { test_id: string }) {
	return instance.post('/inbox_placement/cancel', params, {
		fetchOptions: {
			loading: t('template.inboxPlacement.loading.canceling'),
			successMessage: true,
		},
	})
}

export function getSeedList() {
	return instance.get('/inbox_placement/seed_list')
}

export function addSeed(params: SeedAddParams) {
	return instance.post('/inbox_placement/seed_add', params, {
		fetchOptions: {
			loading: t('template.inboxPlacement.loading.saving'),
			successMessage: true,
		},
	})
}

export function getSeedOAuthUrl(params: { seed_group: string }) {
	return instance.post('/inbox_placement/seed_oauth', params)
}

export function updateSeed(params: SeedUpdateParams) {
	return instance.post('/inbox_placement/seed_update', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export function checkSeed(params: { id: number }) {
	return instance.post('/inbox_placement/seed_check', params, {
		fetchOptions: {
			loading: t('template.inboxPlacement.loading.checking'),
		},
	})
}
