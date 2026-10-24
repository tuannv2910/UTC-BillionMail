<template>
	<div class="p-24px">
		<bt-table-layout>
			<template #toolsRight>
				<n-select
					v-model:value="tableParams.tier"
					:options="tierOptions"
					:placeholder="$t('ui.videoOutreach.filterByTier')"
					clearable
					style="width: 160px; margin-right: 12px"
					@update:value="() => fetchTable(true)"
				/>
			</template>
			<template #table>
				<n-data-table v-bind="tableProps" :columns="columns" :row-key="(row: any) => row.contact_email">
					<template #empty>
						<n-empty :description="$t('ui.videoOutreach.empty')" />
					</template>
				</n-data-table>
			</template>
			<template #pageRight>
				<bt-table-page v-bind="pageProps" @refresh="fetchTable"> </bt-table-page>
			</template>
		</bt-table-layout>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NTag, NButton } from 'naive-ui'
import { useDataTable } from '@/hooks/useDataTable'
import { listVideoOutreach } from '@/api/modules/video-outreach'

const { t } = useI18n()

const tierOptions = [
	{ label: t('ui.videoOutreach.tier1'), value: 'tier_1' },
	{ label: t('ui.videoOutreach.tier2'), value: 'tier_2' },
]

const { tableProps, pageProps, tableParams, fetchTable } = useDataTable<any>({
	params: {
		page: 1,
		page_size: 20,
		tier: null as string | null,
	},
	fetchFn: listVideoOutreach,
})

const columns: DataTableColumns<any> = [
	{
		title: t('ui.videoOutreach.email'),
		key: 'contact_email',
		ellipsis: { tooltip: true },
	},
	{
		title: t('ui.videoOutreach.tier'),
		key: 'lead_tier',
		width: 100,
		render: (row) => {
			const typeMap: Record<string, string> = {
				tier_1: 'success',
				tier_2: 'info',
			}
			return <NTag type={typeMap[row.lead_tier] || 'default'} size="small">{row.lead_tier || '-'}</NTag>
		},
	},
	{
		title: t('ui.videoOutreach.score'),
		key: 'lead_score',
		width: 80,
	},
	{
		title: t('ui.videoOutreach.status'),
		key: 'status',
		width: 120,
		render: (row) => {
			const typeMap: Record<string, string> = {
				ready: 'success',
				pending: 'warning',
				not_eligible: 'default',
			}
			return <NTag type={typeMap[row.status] || 'default'} size="small">{row.status}</NTag>
		},
	},
	{
		title: t('ui.videoOutreach.video'),
		key: 'video_url',
		width: 100,
		render: (row) => {
			if (row.video_url) {
				return <NButton text type="primary" tag="a" href={row.landing_page_url} target="_blank">{t('common.actions.view')}</NButton>
			}
			return '-'
		},
	},
]
</script>
