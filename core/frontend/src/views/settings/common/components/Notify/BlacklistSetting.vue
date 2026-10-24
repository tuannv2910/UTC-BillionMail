<template>
	<div>
		<div class="block text-sm font-medium mb-3 text-basic">
			{{ $t('settings.common.notifySettings.blacklist.title') }}
		</div>
		<div class="flex items-center gap-16px">
			<n-switch :value="blacklistConfig.alert_enabled" @update:value="onUpdateValue" />
			<n-button @click="onShowAlarm">
				{{ $t('settings.common.notifySettings.blacklist.alarmSet') }}
			</n-button>
			<n-button @click="onShowWebhook">
				{{ $t('settings.common.notifySettings.blacklist.webhook.button') }}
			</n-button>
		</div>

		<AlarmModal />
		<WebhookModal />
	</div>
</template>

<script lang="ts" setup>
import { setBlacklistAlert } from '@/api/modules/settings/common'
import { useModal } from '@/hooks/modal/useModal'
import { getSettingsStore } from '../../store'

import AlarmSettings from './AlarmSettings.vue'
import WebhookSettings from './WebhookSettings.vue'

const { blacklistConfig } = getSettingsStore()

const onUpdateValue = async (value: boolean) => {
	await setBlacklistAlert({ enabled: value })
	blacklistConfig.value.alert_enabled = value
}

const [AlarmModal, modalApi] = useModal({
	component: AlarmSettings,
})

const [WebhookModal, webhookModalApi] = useModal({
	component: WebhookSettings,
})

const onShowAlarm = () => {
	modalApi.open()
}

const onShowWebhook = () => {
	webhookModalApi.open()
}
</script>

<style lang="scss" scoped></style>
