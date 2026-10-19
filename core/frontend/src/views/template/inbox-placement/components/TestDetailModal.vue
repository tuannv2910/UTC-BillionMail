<template>
	<modal
		:title="$t('template.inboxPlacement.detail.title')"
		width="960"
		:footer="false"
		:on-after-leave="handleAfterLeave">
		<n-spin :show="loading && !test">
			<div v-if="test" class="detail">
				<div class="detail__head">
					<div class="min-w-0">
						<div class="detail__subject">{{ test.subject }}</div>
						<div class="detail__meta">
							<span>{{ test.from }}</span>
							<span>·</span>
							<span>{{ formatTime(test.created_at) }}</span>
							<span>·</span>
							<span>{{ $t('template.inboxPlacement.create.seedGroup') }}: {{ test.seed_group }}</span>
						</div>
					</div>
					<n-flex :size="12" class="items-center shrink-0">
						<n-tag :type="testStatusTag[test.status]" :bordered="false">
							<template v-if="running" #icon>
								<n-spin :size="12" />
							</template>
							{{ $t(`template.inboxPlacement.status.${test.status}`) }}
						</n-tag>
						<n-button v-if="running" size="small" secondary type="error" @click="handleCancel">
							{{ $t('template.inboxPlacement.actions.cancel') }}
						</n-button>
					</n-flex>
				</div>

				<n-alert v-if="test.error" type="error" :show-icon="false" class="mb-16px">
					{{ test.error }}
				</n-alert>
				<n-alert v-else-if="running" type="info" :show-icon="false" class="mb-16px">
					{{ $t('template.inboxPlacement.detail.runningTip', { min: Math.round(test.timeout_sec / 60) }) }}
				</n-alert>

				<div class="detail__stats">
					<div class="stat">
						<div class="stat__label">{{ $t('template.inboxPlacement.columns.inboxRate') }}</div>
						<div class="stat__value" :class="`is-${rateLevel(test.summary.inbox_rate)}`">
							{{ test.summary.delivered ? formatRate(test.summary.inbox_rate) : '--' }}
						</div>
						<div class="stat__hint">{{ $t('template.inboxPlacement.detail.inboxRateHint') }}</div>
					</div>
					<div class="stat">
						<div class="stat__label">{{ $t('template.inboxPlacement.detail.inboxAnyRate') }}</div>
						<div class="stat__value">
							{{ test.summary.delivered ? formatRate(test.summary.inbox_any_rate) : '--' }}
						</div>
						<div class="stat__hint">{{ $t('template.inboxPlacement.detail.inboxAnyRateHint') }}</div>
					</div>
					<div class="stat">
						<div class="stat__label">{{ $t('template.inboxPlacement.detail.spamRate') }}</div>
						<div class="stat__value" :class="{ 'is-error': test.summary.spam > 0 }">
							{{ test.summary.delivered ? formatRate(test.summary.spam_rate) : '--' }}
						</div>
						<div class="stat__hint">{{ test.summary.spam }} / {{ test.summary.delivered }}</div>
					</div>
					<div class="stat">
						<div class="stat__label">{{ $t('template.inboxPlacement.detail.delivered') }}</div>
						<div class="stat__value">{{ test.summary.delivered }} / {{ test.summary.total }}</div>
						<div class="stat__hint">
							{{ $t('template.inboxPlacement.placement.missing') }} {{ test.summary.missing }} ·
							{{ $t('template.inboxPlacement.placement.pending') }} {{ test.summary.pending }}
						</div>
					</div>
				</div>

				<placement-bar :summary="test.summary" :height="12" legend class="mb-20px" />

				<n-data-table :columns="columns" :data="test.results" :max-height="360" size="small">
				</n-data-table>
			</div>
		</n-spin>
	</modal>
</template>

<script lang="tsx" setup>
import { DataTableColumns, NTag, NTooltip } from 'naive-ui'
import { confirm, formatTime, isObject } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { cancelPlacementTest, getPlacementTestDetail } from '@/api/modules/inbox-placement'
import {
	authTag,
	formatLatency,
	formatRate,
	isRunning,
	placementMeta,
	rateLevel,
	testStatusTag,
} from '../placement'
import type { PlacementTest, SeedResult } from '../interface'

import PlacementBar from './PlacementBar.vue'

interface ModalState {
	testId: string
	refresh?: () => void
}

const POLL_MS = 5000

const { t } = useI18n()

const loading = ref(false)

const test = ref<PlacementTest | null>(null)

const running = computed(() => !!test.value && isRunning(test.value.status))

let timer: ReturnType<typeof setTimeout> | null = null

// False once the modal is closed, so an in-flight request does not re-arm polling.
let active = false

const stopPolling = () => {
	if (timer) clearTimeout(timer)
	timer = null
}

const load = async () => {
	const { testId } = modalApi.getState<ModalState>()
	if (!testId) return
	loading.value = true
	try {
		const res = await getPlacementTestDetail({ test_id: testId })
		if (isObject<PlacementTest>(res)) {
			const wasRunning = running.value
			test.value = res
			// Refresh the list once the test settles.
			if (wasRunning && !isRunning(res.status)) {
				modalApi.getState<ModalState>().refresh?.()
			}
		}
	} finally {
		loading.value = false
	}
	stopPolling()
	if (active && running.value) {
		timer = setTimeout(load, POLL_MS)
	}
}

const authCell = (verdict?: string) =>
	verdict ? (
		<NTag size="small" bordered={false} type={authTag(verdict)}>
			{verdict}
		</NTag>
	) : (
		'--'
	)

const columns: DataTableColumns<SeedResult> = [
	{
		key: 'seed',
		title: t('template.inboxPlacement.columns.seed'),
		minWidth: 180,
		ellipsis: { tooltip: true },
		render: row => (
			<div class="flex items-center gap-6px">
				<span>{row.seed}</span>
				<span class="text-12px text-[var(--color-text-3,#909399)]">{row.provider}</span>
			</div>
		),
	},
	{
		key: 'status',
		title: t('template.inboxPlacement.columns.placement'),
		width: 140,
		render: row => {
			const tag = (
				<NTag size="small" bordered={false} type={placementMeta[row.status]?.tag || 'default'}>
					{t(`template.inboxPlacement.placement.${row.status}`)}
				</NTag>
			)
			if (!row.error) return tag
			return (
				<NTooltip>
					{{
						trigger: () => tag,
						default: () => row.error,
					}}
				</NTooltip>
			)
		},
	},
	{
		key: 'received_at',
		title: t('template.inboxPlacement.columns.receivedAt'),
		width: 170,
		render: row => formatTime(row.received_at),
	},
	{
		key: 'latency_ms',
		title: t('template.inboxPlacement.columns.latency'),
		width: 100,
		render: row => formatLatency(row.latency_ms),
	},
	{ key: 'spf', title: 'SPF', width: 90, render: row => authCell(row.spf) },
	{ key: 'dkim', title: 'DKIM', width: 90, render: row => authCell(row.dkim) },
	{ key: 'dmarc', title: 'DMARC', width: 90, render: row => authCell(row.dmarc) },
]

const handleCancel = () => {
	if (!test.value) return
	const testId = test.value.test_id
	confirm({
		title: t('template.inboxPlacement.actions.cancel'),
		content: t('template.inboxPlacement.cancelConfirm'),
		confirmType: 'error',
		onConfirm: async () => {
			await cancelPlacementTest({ test_id: testId })
			await load()
			modalApi.getState<ModalState>().refresh?.()
		},
	})
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		stopPolling()
		if (isOpen) {
			active = true
			test.value = null
			load()
		}
	},
})

const handleAfterLeave = () => {
	active = false
	stopPolling()
}

onBeforeUnmount(handleAfterLeave)
</script>

<style lang="scss" scoped>
.detail {
	padding-top: 4px;

	&__head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 16px;
	}

	&__subject {
		font-size: 16px;
		font-weight: 600;
		color: var(--color-text-1, #303133);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	&__meta {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 6px;
		font-size: 12px;
		color: var(--color-text-3, #909399);
	}

	&__stats {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 12px;
		margin-bottom: 16px;
	}
}

.stat {
	padding: 14px 16px;
	border: 1px solid var(--color-border-1, #ebeef5);
	border-radius: 6px;

	&__label {
		font-size: 12px;
		color: var(--color-text-3, #909399);
	}

	&__value {
		margin-top: 6px;
		font-size: 24px;
		font-weight: 600;
		line-height: 1.2;
		color: var(--color-text-1, #303133);

		&.is-success {
			color: var(--color-primary-1);
		}

		&.is-warning {
			color: var(--color-warning-1);
		}

		&.is-error {
			color: var(--color-error-1);
		}
	}

	&__hint {
		margin-top: 4px;
		font-size: 12px;
		color: var(--color-text-3, #909399);
	}
}
</style>
