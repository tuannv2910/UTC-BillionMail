import { BLACKLIST_REASONS, BlacklistScope, BlacklistStatus } from './types/base'

/**
 * @description Select options shared by the blacklist table, form and import modal
 */
export const useBlacklistOptions = () => {
	const { t } = useI18n()

	const reasonOptions = computed(() =>
		BLACKLIST_REASONS.map((key, index) => ({
			label: t(`contacts.blacklist.reasons.${key}`),
			value: index + 1,
		}))
	)

	// "legacy" only comes from migrating the old suspend list
	const editableReasonOptions = computed(() =>
		reasonOptions.value.filter(option => BLACKLIST_REASONS[option.value - 1] !== 'legacy')
	)

	const reasonFilterOptions = computed(() => [
		{ label: t('contacts.blacklist.filters.allReasons'), value: 0 },
		...reasonOptions.value,
	])

	const statusFilterOptions = computed(() => [
		{ label: t('contacts.blacklist.filters.allStatus'), value: BlacklistStatus.All },
		{ label: t('contacts.blacklist.status.active'), value: BlacklistStatus.Active },
		{ label: t('contacts.blacklist.status.observing'), value: BlacklistStatus.Observing },
		{ label: t('contacts.blacklist.status.released'), value: BlacklistStatus.Released },
	])

	const scopeOptions = computed(() => [
		{ label: t('contacts.blacklist.scope.email'), value: BlacklistScope.Email },
		{ label: t('contacts.blacklist.scope.domain'), value: BlacklistScope.Domain },
	])

	const scopeFilterOptions = computed(() => [
		{ label: t('contacts.blacklist.filters.allScopes'), value: 0 },
		...scopeOptions.value,
	])

	// ttl_days: -1 uses the reason's default, 0 blocks permanently
	const durationOptions = computed(() => [
		{ label: t('contacts.blacklist.duration.default'), value: -1 },
		{ label: t('contacts.blacklist.duration.permanent'), value: 0 },
		...[7, 30, 90, 180, 365].map(days => ({
			label: t('contacts.blacklist.duration.days', { days }),
			value: days,
		})),
	])

	return {
		reasonOptions,
		editableReasonOptions,
		reasonFilterOptions,
		statusFilterOptions,
		scopeOptions,
		scopeFilterOptions,
		durationOptions,
	}
}
