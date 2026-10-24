<template>
	<modal :title="$t('settings.common.notifySettings.blacklist.webhook.title')" :width="520">
		<bt-form class="pt-8px">
			<n-form-item :label="$t('settings.common.notifySettings.blacklist.webhook.enabled')">
				<n-switch v-model:value="form.enabled" />
			</n-form-item>
			<n-form-item :label="$t('settings.common.notifySettings.blacklist.webhook.format')">
				<n-select v-model:value="form.format" :options="formatOptions" />
			</n-form-item>
			<n-form-item :label="$t('settings.common.notifySettings.blacklist.webhook.url')">
				<n-input
					v-model:value="form.url"
					:placeholder="$t('settings.common.notifySettings.blacklist.webhook.urlPlaceholder')" />
			</n-form-item>
			<n-form-item :show-label="false">
				<n-checkbox v-model:checked="form.send_test">
					{{ $t('settings.common.notifySettings.blacklist.webhook.sendTest') }}
				</n-checkbox>
			</n-form-item>
		</bt-form>
	</modal>
</template>

<script lang="ts" setup>
import { useModal } from '@/hooks/modal/useModal'
import {
	getAlertWebhook,
	setAlertWebhook,
	type AlertWebhookFormat,
	type AlertWebhookSettings,
} from '@/api/modules/settings/common'

const form = reactive({
	enabled: false,
	url: '',
	format: 'json' as AlertWebhookFormat,
	send_test: true,
})

const formatOptions = [
	{ label: 'JSON', value: 'json' },
	{ label: 'Feishu / Lark', value: 'feishu' },
	{ label: 'WeCom', value: 'wecom' },
	{ label: 'DingTalk', value: 'dingtalk' },
	{ label: 'Slack', value: 'slack' },
	{ label: 'Gotify', value: 'gotify' },
]

const loadSettings = async () => {
	const res = (await getAlertWebhook()) as AlertWebhookSettings | undefined
	if (res) {
		form.enabled = !!res.enabled
		form.url = res.url || ''
		form.format = res.format || 'json'
	}
}

const [Modal] = useModal({
	onChangeState: val => {
		if (val) loadSettings()
	},
	onConfirm: async () => {
		await setAlertWebhook({ ...form })
	},
})
</script>

<style lang="scss" scoped></style>
