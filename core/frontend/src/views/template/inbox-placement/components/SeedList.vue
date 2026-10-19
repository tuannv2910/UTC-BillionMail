<template>
	<div>
		<n-alert v-if="!adminEnabled" type="warning" :show-icon="false" class="mb-16px">
			{{ $t('template.inboxPlacement.seeds.adminDisabled') }}
		</n-alert>
		<bt-table-layout>
			<template #toolsLeft>
				<n-input-group class="w-auto!">
					<n-input
						v-model:value="connectGroup"
						class="w-140px!"
						placeholder="default"
						:disabled="!adminEnabled"></n-input>
					<n-button type="primary" :disabled="!adminEnabled" :loading="connecting" @click="handleConnect">
						{{ $t('template.inboxPlacement.seeds.connectGmail') }}
					</n-button>
				</n-input-group>
				<n-button secondary :disabled="!adminEnabled" @click="handleAdd">
					{{ $t('template.inboxPlacement.seeds.addWithToken') }}
				</n-button>
			</template>
			<template #toolsRight>
				<span class="text-12px text-[var(--color-text-3,#909399)]">
					{{ $t('template.inboxPlacement.seeds.activeCount', { active: activeCount, total: list.length }) }}
				</span>
				<n-button secondary :disabled="!adminEnabled" @click="getList">
					{{ $t('template.inboxPlacement.actions.refresh') }}
				</n-button>
			</template>
			<template #table>
				<n-data-table :loading="loading" :columns="columns" :data="list" :row-key="row => row.id">
					<template #empty>
						<n-empty :description="$t('template.inboxPlacement.seeds.empty')"></n-empty>
					</template>
				</n-data-table>
			</template>
			<template #modal>
				<add-modal />
			</template>
		</bt-table-layout>

		<bt-tips class="mt-16px">
			<li>{{ $t('template.inboxPlacement.seeds.tips.production') }}</li>
			<li>{{ $t('template.inboxPlacement.seeds.tips.noInteract') }}</li>
			<li>{{ $t('template.inboxPlacement.seeds.tips.autoDisable') }}</li>
		</bt-tips>
	</div>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NButton, NSwitch, NTag, NTooltip } from 'naive-ui'
import { formatTime, isObject, Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { checkSeed, getSeedList, getSeedOAuthUrl, updateSeed } from '@/api/modules/inbox-placement'
import type { Seed } from '../interface'

import AddSeedModal from './AddSeedModal.vue'

const props = defineProps({
	adminEnabled: {
		type: Boolean,
		default: true,
	},
})

const { t } = useI18n()

const loading = ref(false)

const list = ref<Seed[]>([])

const connectGroup = ref('default')

const connecting = ref(false)

const activeCount = computed(() => list.value.filter(item => item.is_active).length)

const getList = async () => {
	if (!props.adminEnabled) return
	loading.value = true
	try {
		const res = await getSeedList()
		if (isObject<{ list: Seed[] }>(res)) {
			list.value = res.list || []
		}
	} finally {
		loading.value = false
	}
}

const healthTag: Record<Seed['health_status'], 'success' | 'error' | 'warning' | 'default'> = {
	ok: 'success',
	auth_revoked: 'error',
	error: 'warning',
	unknown: 'default',
}

const handleToggle = async (row: Seed, value: boolean) => {
	await updateSeed({ id: row.id, is_active: value })
	row.is_active = value
}

const handleCheck = async (row: Seed) => {
	const res = await checkSeed({ id: row.id })
	if (isObject<Seed>(res)) {
		Object.assign(row, res)
		if (res.health_status === 'ok') {
			Message.success(t('template.inboxPlacement.seeds.checkOk'))
		} else {
			Message.error(res.last_error || t('template.inboxPlacement.seeds.checkFailed'))
		}
	}
}

const columns = computed<DataTableColumns<Seed>>(() => [
	{
		key: 'email',
		title: t('template.inboxPlacement.seeds.email'),
		minWidth: 220,
		ellipsis: { tooltip: true },
	},
	{
		key: 'provider',
		title: t('template.inboxPlacement.seeds.provider'),
		width: 100,
		render: row => (
			<NTag size="small" bordered={false}>
				{row.provider}
			</NTag>
		),
	},
	{
		key: 'seed_group',
		title: t('template.inboxPlacement.create.seedGroup'),
		width: 120,
	},
	{
		key: 'health_status',
		title: t('template.inboxPlacement.seeds.health'),
		width: 150,
		render: row => {
			const tag = (
				<NTag size="small" bordered={false} type={healthTag[row.health_status] || 'default'}>
					{t(`template.inboxPlacement.seeds.healthStatus.${row.health_status}`)}
				</NTag>
			)
			if (!row.last_error) return tag
			return (
				<NTooltip style={{ maxWidth: '360px' }}>
					{{
						trigger: () => tag,
						default: () => row.last_error,
					}}
				</NTooltip>
			)
		},
	},
	{
		key: 'last_health_at',
		title: t('template.inboxPlacement.seeds.lastCheck'),
		width: 170,
		render: row => formatTime(row.last_health_at),
	},
	{
		key: 'is_active',
		title: t('template.inboxPlacement.seeds.active'),
		width: 90,
		render: row => (
			<NSwitch
				size="small"
				value={row.is_active}
				onUpdateValue={(val: boolean) => handleToggle(row, val)}
			/>
		),
	},
	{
		title: t('common.columns.actions'),
		key: 'actions',
		align: 'right',
		width: 120,
		render: row => (
			<NButton type="primary" text={true} onClick={() => handleCheck(row)}>
				{t('template.inboxPlacement.seeds.check')}
			</NButton>
		),
	},
])

const [AddModal, addModalApi] = useModal({
	component: AddSeedModal,
	state: {
		refresh: getList,
	},
})

const handleAdd = () => {
	addModalApi.open()
}

// Google consent runs in a popup; the callback page posts a message back.
const onMessage = (event: MessageEvent) => {
	if (event.origin !== window.location.origin) return
	if (isObject<{ type: string; ok: boolean }>(event.data) && event.data.type === 'inbox-placement-seed') {
		if (event.data.ok) {
			Message.success(t('template.inboxPlacement.seeds.connected'))
		}
		getList()
	}
}

const handleConnect = async () => {
	connecting.value = true
	// Open synchronously so the popup is not blocked, then navigate it.
	const popup = window.open('', 'inbox-placement-oauth', 'width=520,height=680')
	try {
		const res = await getSeedOAuthUrl({ seed_group: connectGroup.value.trim() || 'default' })
		if (isObject<{ url: string }>(res) && res.url) {
			if (popup) {
				popup.location.href = res.url
			} else {
				window.open(res.url, '_blank')
			}
		} else {
			popup?.close()
		}
	} catch {
		popup?.close()
	} finally {
		connecting.value = false
	}
}

onMounted(() => window.addEventListener('message', onMessage))
onBeforeUnmount(() => window.removeEventListener('message', onMessage))

watch(
	() => props.adminEnabled,
	val => {
		if (val) getList()
	},
	{ immediate: true }
)
</script>
