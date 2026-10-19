<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" :disabled="!enabled" @click="handleCreate()">
				{{ $t('template.inboxPlacement.actions.newTest') }}
			</n-button>
		</template>
		<template #toolsRight>
			<n-select
				v-model:value="tableParams.status"
				class="w-160px!"
				clearable
				:options="statusOptions"
				:placeholder="$t('template.inboxPlacement.columns.status')"
				@update:value="() => getTableData(true)">
			</n-select>
			<n-select
				v-model:value="tableParams.template_id"
				class="w-240px!"
				clearable
				filterable
				:options="templateOptions"
				:placeholder="$t('template.inboxPlacement.filterTemplate')"
				@update:value="() => getTableData(true)">
			</n-select>
		</template>
		<template #table>
			<n-data-table :loading="loading" :columns="columns" :data="tableList" :row-key="row => row.test_id">
				<template #empty>
					<n-empty :description="$t('template.inboxPlacement.empty')">
						<template #extra>
							<n-button size="small" type="primary" :disabled="!enabled" @click="handleCreate()">
								{{ $t('template.inboxPlacement.actions.newTest') }}
							</n-button>
						</template>
					</n-empty>
				</template>
			</n-data-table>
		</template>
		<template #pageRight>
			<bt-table-page
				v-model:page="tableParams.page"
				v-model:page-size="tableParams.page_size"
				:item-count="tableTotal"
				@refresh="getTableData">
			</bt-table-page>
		</template>
		<template #modal>
			<create-modal />
			<detail-modal />
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NFlex, NSpin, NTag } from 'naive-ui'
import { confirm, formatTime, isArray } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { useTableData } from '@/hooks/useTableData'
import { cancelPlacementTest, getPlacementTestList } from '@/api/modules/inbox-placement'
import { getTemplateAll } from '@/api/modules/market/template'
import { formatRate, isRunning, rateLevel, testStatusTag } from '../placement'
import type { PlacementTest, TestStatus } from '../interface'
import type { Template } from '@/views/template/interface'

import PlacementBar from './PlacementBar.vue'
import CreateTestModal from './CreateTestModal.vue'
import TestDetailModal from './TestDetailModal.vue'

const props = defineProps({
	enabled: {
		type: Boolean,
		default: true,
	},
})

const { t } = useI18n()

const route = useRoute()

const REFRESH_MS = 10000

interface ListParams {
	page: number
	page_size: number
	template_id: number | null
	status: string | null
}

const { tableParams, tableList, loading, tableTotal, getTableData } = useTableData<
	PlacementTest,
	ListParams
>({
	loading: true,
	params: {
		page: 1,
		page_size: 10,
		template_id: Number(route.query.template_id) || null,
		status: null,
	},
	fetchFn: params =>
		getPlacementTestList({
			page: params.page,
			page_size: params.page_size,
			template_id: params.template_id || undefined,
			status: params.status || undefined,
		}),
})

const statusOptions = computed(() =>
	(['pending', 'sending', 'checking', 'completed', 'failed', 'canceled'] as TestStatus[]).map(value => ({
		label: t(`template.inboxPlacement.status.${value}`),
		value,
	}))
)

const templateNames = ref<Record<number, string>>({})

const templateOptions = computed(() =>
	Object.entries(templateNames.value).map(([id, name]) => ({ label: name, value: Number(id) }))
)

const loadTemplates = async () => {
	const res = await getTemplateAll()
	if (isArray<Template>(res)) {
		templateNames.value = Object.fromEntries(res.map(item => [item.id, item.temp_name]))
	}
}

const columns = computed<DataTableColumns<PlacementTest>>(() => [
	{
		key: 'subject',
		title: t('template.inboxPlacement.columns.test'),
		minWidth: 240,
		render: row => (
			<div class="min-w-0">
				<NButton text type="primary" class="max-w-full" onClick={() => handleDetail(row)}>
					<span class="truncate">{row.subject}</span>
				</NButton>
				<div class="text-12px text-[var(--color-text-3,#909399)] truncate mt-2px">
					{row.from}
					{row.template_id && templateNames.value[row.template_id]
						? ` · ${templateNames.value[row.template_id]}`
						: ''}
				</div>
			</div>
		),
	},
	{
		key: 'status',
		title: t('template.inboxPlacement.columns.status'),
		width: 130,
		render: row => (
			<NTag size="small" bordered={false} type={testStatusTag[row.status]}>
				{{
					icon: isRunning(row.status) ? () => <NSpin size={10} /> : undefined,
					default: () => t(`template.inboxPlacement.status.${row.status}`),
				}}
			</NTag>
		),
	},
	{
		key: 'placement',
		title: t('template.inboxPlacement.columns.placement'),
		minWidth: 200,
		render: row => (
			<div class="max-w-260px">
				<PlacementBar summary={row.summary} />
				<div class="text-12px text-[var(--color-text-3,#909399)] mt-4px">
					{t('template.inboxPlacement.placement.inbox')} {row.summary.inbox} ·{' '}
					{t('template.inboxPlacement.columns.tabs')}{' '}
					{row.summary.promotions + row.summary.social + row.summary.updates + row.summary.forums} ·{' '}
					{t('template.inboxPlacement.placement.spam')} {row.summary.spam} ·{' '}
					{t('template.inboxPlacement.placement.missing')} {row.summary.missing}
				</div>
			</div>
		),
	},
	{
		key: 'inbox_rate',
		title: t('template.inboxPlacement.columns.inboxRate'),
		width: 110,
		render: row => {
			if (!row.summary.delivered) return '--'
			return (
				<NTag size="small" bordered={false} type={rateLevel(row.summary.inbox_rate)}>
					{formatRate(row.summary.inbox_rate)}
				</NTag>
			)
		},
	},
	{
		key: 'created_at',
		title: t('template.inboxPlacement.columns.createdAt'),
		width: 170,
		render: row => formatTime(row.created_at),
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 200,
		render: row => (
			<NFlex inline={true} justify="end">
				<NButton type="primary" text={true} onClick={() => handleDetail(row)}>
					{t('template.inboxPlacement.actions.detail')}
				</NButton>
				<NButton type="primary" text={true} onClick={() => handleRetest(row)}>
					{t('template.inboxPlacement.actions.retest')}
				</NButton>
				{isRunning(row.status) && (
					<NButton type="error" text={true} onClick={() => handleCancel(row)}>
						{t('template.inboxPlacement.actions.cancel')}
					</NButton>
				)}
			</NFlex>
		),
	},
])

const [DetailModal, detailModalApi] = useModal({
	component: TestDetailModal,
	state: {
		testId: '',
		refresh: () => getTableData(),
	},
})

const handleDetail = (row: Pick<PlacementTest, 'test_id'>) => {
	detailModalApi.setState({ testId: row.test_id })
	detailModalApi.open()
}

const [CreateModal, createModalApi] = useModal({
	component: CreateTestModal,
	state: {
		templateId: null,
		subject: '',
		onCreated: async (testId: string) => {
			await getTableData(true)
			handleDetail({ test_id: testId })
		},
	},
})

const handleCreate = (templateId: number | null = null, subject = '') => {
	createModalApi.setState({ templateId, subject })
	createModalApi.open()
}

const handleRetest = (row: PlacementTest) => {
	handleCreate(row.template_id || null, row.subject)
}

const handleCancel = (row: PlacementTest) => {
	confirm({
		title: t('template.inboxPlacement.actions.cancel'),
		content: t('template.inboxPlacement.cancelConfirm'),
		confirmType: 'error',
		onConfirm: async () => {
			await cancelPlacementTest({ test_id: row.test_id })
			getTableData()
		},
	})
}

// Refresh silently while some test on the page is still running.
let timer: ReturnType<typeof setInterval> | null = null

watch(
	tableList,
	list => {
		const anyRunning = list.some(item => isRunning(item.status))
		if (anyRunning && !timer) {
			timer = setInterval(() => {
				if (!loading.value) getTableData()
			}, REFRESH_MS)
		} else if (!anyRunning && timer) {
			clearInterval(timer)
			timer = null
		}
	},
	{ deep: false }
)

onBeforeUnmount(() => {
	if (timer) clearInterval(timer)
})

const init = async () => {
	await Promise.all([loadTemplates(), props.enabled ? getTableData() : Promise.resolve()])
	loading.value = false
	if (typeof route.query.test_id === 'string' && route.query.test_id) {
		handleDetail({ test_id: route.query.test_id })
	}
}

init()

defineExpose({ create: handleCreate, refresh: getTableData })
</script>
