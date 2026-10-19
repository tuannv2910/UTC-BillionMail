<template>
	<div>
		<div class="stats-row">
			<div v-for="card in statCards" :key="card.key" class="stat-card">
				<div class="stat-label">{{ card.label }}</div>
				<div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
			</div>
			<div class="stat-card stat-reasons">
				<div class="stat-label">{{ $t('contacts.blacklist.stats.topReasons') }}</div>
				<div v-if="topReasons.length" class="flex flex-wrap gap-6px mt-6px">
					<n-tag v-for="item in topReasons" :key="item.reason" size="small" :bordered="false">
						{{ $t(`contacts.blacklist.reasons.${reasonKey(item.reason)}`) }}: {{ item.total }}
					</n-tag>
				</div>
				<div v-else class="stat-value">--</div>
			</div>
		</div>

		<bt-table-layout>
			<template #toolsLeft>
				<n-button type="primary" @click="handleAdd">{{ $t('contacts.blacklist.buttons.add') }}</n-button>
				<n-button @click="handleImport">{{ $t('contacts.blacklist.buttons.import') }}</n-button>
				<n-button :disabled="tableKeys.length === 0" @click="handleBatchRelease">
					{{ $t('contacts.blacklist.buttons.release') }}
				</n-button>
				<n-button :disabled="tableKeys.length === 0" @click="handleBatchDelete">
					{{ $t('common.actions.delete') }}
				</n-button>
			</template>
			<template #toolsRight>
				<n-select
					v-model:value="tableParams.status"
					class="w-130px"
					:options="statusFilterOptions"
					@update:value="resetTable">
				</n-select>
				<n-select
					v-model:value="tableParams.reason"
					class="w-170px"
					:options="reasonFilterOptions"
					@update:value="resetTable">
				</n-select>
				<n-select
					v-model:value="tableParams.scope"
					class="w-120px"
					:options="scopeFilterOptions"
					@update:value="resetTable">
				</n-select>
				<bt-search
					v-model:value="tableParams.keyword"
					:width="220"
					:placeholder="$t('contacts.blacklist.search.placeholder')"
					@search="resetTable">
				</bt-search>
			</template>
			<template #table>
				<n-data-table v-bind="tableProps" :columns="columns" :scroll-x="1300"></n-data-table>
			</template>
			<template #pageRight>
				<bt-table-page v-bind="pageProps" @refresh="refresh"></bt-table-page>
			</template>
			<template #modal>
				<form-modal></form-modal>
				<import-modal></import-modal>
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex, NTag, NTooltip } from 'naive-ui'
import { confirm, formatTime, isObject } from '@/utils'
import { useDataTable } from '@/hooks/useDataTable'
import { useModal } from '@/hooks/modal/useModal'
import {
	createBlacklist,
	deleteBlacklist,
	getBlacklist,
	getBlacklistStats,
	releaseBlacklist,
} from '@/api/modules/contacts/blacklist'
import {
	BlacklistScope,
	BlacklistStatus,
	isBlocking,
	reasonKey,
	sourceKey,
	type BlacklistEntry,
	type BlacklistParams,
	type BlacklistStats,
} from './types/base'
import { useBlacklistOptions } from './useBlacklistOptions'

import BlacklistForm from './components/BlacklistForm.vue'
import BlacklistImport from './components/BlacklistImport.vue'

const { t } = useI18n()

const { reasonFilterOptions, statusFilterOptions, scopeFilterOptions } = useBlacklistOptions()

// Temporary reasons often recover on their own; the rest are permanent problems
const SOFT_REASONS = ['mailbox_full', 'blocked']

const stats = ref<BlacklistStats>({ active: 0, observing: 0, released: 0, by_reason: [] })

const statCards = computed(() => [
	{ key: 'active', label: t('contacts.blacklist.stats.active'), value: stats.value.active, color: 'var(--color-error-1)' },
	{
		key: 'observing',
		label: t('contacts.blacklist.stats.observing'),
		value: stats.value.observing,
		color: 'var(--color-warning-1)',
	},
	{
		key: 'released',
		label: t('contacts.blacklist.stats.released'),
		value: stats.value.released,
		color: 'var(--color-text-2)',
	},
])

const topReasons = computed(() => [...stats.value.by_reason].sort((a, b) => b.total - a.total).slice(0, 4))

const fetchStats = async () => {
	const res = await getBlacklistStats()
	if (isObject<BlacklistStats>(res)) {
		stats.value = { ...res, by_reason: res.by_reason || [] }
	}
}

const { tableParams, tableProps, pageProps, tableKeys, fetchTable, resetTable } = useDataTable<
	BlacklistEntry,
	BlacklistParams
>({
	params: {
		page: 1,
		page_size: 20,
		keyword: '',
		scope: 0,
		reason: 0,
		source: 0,
		status: BlacklistStatus.All,
	},
	fetchFn: getBlacklist,
})

const refresh = () => {
	fetchTable()
	fetchStats()
}

fetchStats()

const renderStatus = (row: BlacklistEntry) => {
	const now = Math.floor(Date.now() / 1000)
	if (row.status === BlacklistStatus.Released) {
		return <NTag size="small" bordered={false} type="success">{t('contacts.blacklist.status.released')}</NTag>
	}
	if (row.status === BlacklistStatus.Observing) {
		return <NTag size="small" bordered={false} type="warning">{t('contacts.blacklist.status.observing')}</NTag>
	}
	if (!isBlocking(row, now)) {
		return <NTag size="small" bordered={false}>{t('contacts.blacklist.status.expired')}</NTag>
	}
	return <NTag size="small" bordered={false} type="error">{t('contacts.blacklist.status.active')}</NTag>
}

const columns = computed<DataTableColumns<BlacklistEntry>>(() => [
	{ type: 'selection', width: 40 },
	{
		key: 'value',
		title: t('contacts.blacklist.columns.value'),
		minWidth: 240,
		render: row => (
			<div class="flex items-center gap-6px">
				<span class="break-all">{row.value}</span>
				{row.scope === BlacklistScope.Domain && (
					<NTag size="tiny" bordered={false} type="info">
						{t('contacts.blacklist.scope.domain')}
					</NTag>
				)}
			</div>
		),
	},
	{
		key: 'reason',
		title: t('contacts.blacklist.columns.reason'),
		width: 170,
		render: row => {
			const key = reasonKey(row.reason)
			const tag = (
				<NTag size="small" bordered={false} type={SOFT_REASONS.includes(key) ? 'warning' : 'error'}>
					{t(`contacts.blacklist.reasons.${key}`)}
				</NTag>
			)
			if (!row.dsn && !row.last_response) return tag
			return (
				<NTooltip>
					{{
						trigger: () => tag,
						default: () => [row.dsn, row.last_response].filter(Boolean).join(' · '),
					}}
				</NTooltip>
			)
		},
	},
	{
		key: 'status',
		title: t('contacts.blacklist.columns.status'),
		width: 110,
		render: renderStatus,
	},
	{
		key: 'hit_count',
		title: t('contacts.blacklist.columns.hits'),
		width: 80,
	},
	{
		key: 'expires_at',
		title: t('contacts.blacklist.columns.expiresAt'),
		width: 170,
		render: row =>
			row.expires_at === 0 ? t('contacts.blacklist.duration.permanent') : formatTime(row.expires_at),
	},
	{
		key: 'source',
		title: t('contacts.blacklist.columns.source'),
		width: 100,
		render: row => t(`contacts.blacklist.sources.${sourceKey(row.source)}`),
	},
	{
		key: 'last_seen_at',
		title: t('contacts.blacklist.columns.lastSeen'),
		width: 170,
		render: row => formatTime(row.last_seen_at),
	},
	{
		key: 'note',
		title: t('contacts.blacklist.columns.note'),
		minWidth: 140,
		ellipsis: { tooltip: true },
		render: row => row.note || '--',
	},
	{
		key: 'actions',
		title: t('common.columns.actions'),
		align: 'right',
		width: 180,
		fixed: 'right',
		render: row => (
			<NFlex inline={true} justify="end">
				<NButton type="primary" text={true} onClick={() => handleEdit(row)}>
					{t('common.actions.edit')}
				</NButton>
				{row.status === BlacklistStatus.Released ? (
					<NButton type="primary" text={true} onClick={() => handleBlockAgain(row)}>
						{t('contacts.blacklist.buttons.blockAgain')}
					</NButton>
				) : (
					<NButton type="primary" text={true} onClick={() => handleRelease([row.id], row.value)}>
						{t('contacts.blacklist.buttons.release')}
					</NButton>
				)}
				<NButton type="error" text={true} onClick={() => handleDelete([row.id], row.value)}>
					{t('common.actions.delete')}
				</NButton>
			</NFlex>
		),
	},
])

const [FormModal, formModalApi] = useModal({
	component: BlacklistForm,
	state: { row: null, refresh },
})

const [ImportModal, importModalApi] = useModal({
	component: BlacklistImport,
	state: { refresh },
})

const handleAdd = () => {
	formModalApi.setState({ row: null })
	formModalApi.open()
}

const handleEdit = (row: BlacklistEntry) => {
	formModalApi.setState({ row })
	formModalApi.open()
}

const handleImport = () => {
	importModalApi.open()
}

const selectedIds = () => tableKeys.value.map(key => Number(key))

const handleRelease = (ids: number[], value?: string) => {
	confirm({
		title: t('contacts.blacklist.confirm.release.title'),
		content: value
			? t('contacts.blacklist.confirm.release.single', { value })
			: t('contacts.blacklist.confirm.release.batch', { count: ids.length }),
		onConfirm: async () => {
			await releaseBlacklist({ ids })
			refresh()
		},
	})
}

const handleBatchRelease = () => handleRelease(selectedIds())

const handleDelete = (ids: number[], value?: string) => {
	confirm({
		title: t('contacts.blacklist.confirm.delete.title'),
		content: value
			? t('contacts.blacklist.confirm.delete.single', { value })
			: t('contacts.blacklist.confirm.delete.batch', { count: ids.length }),
		confirmText: t('common.actions.delete'),
		confirmType: 'error',
		onConfirm: async () => {
			await deleteBlacklist({ ids })
			refresh()
		},
	})
}

const handleBatchDelete = () => handleDelete(selectedIds())

const handleBlockAgain = async (row: BlacklistEntry) => {
	await createBlacklist({
		scope: row.scope,
		value: row.value,
		reason: row.reason,
		note: row.note,
		ttl_days: -1,
	})
	refresh()
}
</script>

<style lang="scss" scoped>
.stats-row {
	display: grid;
	grid-template-columns: repeat(3, minmax(140px, 1fr)) minmax(280px, 2fr);
	gap: 12px;
	margin-bottom: 16px;
}

.stat-card {
	padding: 14px 16px;
	border: 1px solid var(--color-border-1);
	border-radius: 6px;
	background-color: var(--color-bg-1);

	.stat-label {
		font-size: 13px;
		color: var(--color-text-3);
	}

	.stat-value {
		margin-top: 6px;
		font-size: 24px;
		font-weight: 600;
		line-height: 1.2;
	}
}
</style>
