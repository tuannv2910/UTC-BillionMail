<template>
	<modal :title="$t('template.inboxPlacement.seeds.addTitle')" width="540">
		<div class="pt-8px pr-12px">
			<n-alert type="warning" :show-icon="false" class="mb-16px">
				{{ $t('template.inboxPlacement.seeds.addTip') }}
			</n-alert>
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="$t('template.inboxPlacement.seeds.email')" path="email">
					<n-input v-model:value="form.email" placeholder="seed01@gmail.com"></n-input>
				</n-form-item>
				<n-form-item :label="$t('template.inboxPlacement.seeds.refreshToken')" path="refresh_token">
					<n-input
						v-model:value="form.refresh_token"
						type="password"
						show-password-on="click"
						:placeholder="$t('template.inboxPlacement.seeds.refreshTokenPlaceholder')"></n-input>
				</n-form-item>
				<n-form-item :label="$t('template.inboxPlacement.create.seedGroup')" path="seed_group">
					<n-input v-model:value="form.seed_group" placeholder="default"></n-input>
				</n-form-item>
			</bt-form>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { addSeed } from '@/api/modules/inbox-placement'

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const createForm = () => ({
	email: '',
	refresh_token: '',
	seed_group: 'default',
})

const form = reactive(createForm())

const rules: FormRules = {
	email: {
		required: true,
		type: 'email',
		trigger: 'blur',
		message: t('template.inboxPlacement.seeds.emailRequired'),
	},
	refresh_token: {
		required: true,
		trigger: 'blur',
		message: t('template.inboxPlacement.seeds.refreshTokenPlaceholder'),
	},
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (!isOpen) Object.assign(form, createForm())
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		await addSeed({
			email: form.email.trim(),
			refresh_token: form.refresh_token.trim(),
			seed_group: form.seed_group.trim() || 'default',
		})
		modalApi.getState<{ refresh: () => void }>().refresh()
	},
})
</script>
