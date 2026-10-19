<template>
	<modal :title="$t('template.inboxPlacement.create.title')" width="580">
		<div class="pt-8px pr-12px">
			<n-alert type="info" :show-icon="false" class="mb-16px">
				{{ $t('template.inboxPlacement.create.description') }}
			</n-alert>
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item :label="$t('market.task.edit.template')" path="template_id">
					<div class="flex-1">
						<template-select v-model:value="form.template_id" v-model:content="form.template_content">
						</template-select>
					</div>
					<n-button
						text
						type="primary"
						class="ml-12px"
						:disabled="!form.template_content"
						@click="handlePreview">
						{{ $t('common.actions.preview') }}
					</n-button>
				</n-form-item>
				<n-form-item :label="$t('market.task.edit.from')" path="addresser">
					<from-select v-model:value="form.addresser" v-model:domain="form.domain" v-model:name="form.full_name">
					</from-select>
				</n-form-item>
				<n-form-item :label="$t('market.task.edit.displayName')" path="full_name">
					<n-input
						v-model:value="form.full_name"
						:placeholder="$t('market.task.edit.displayNamePlaceholder')"></n-input>
				</n-form-item>
				<n-form-item :label="$t('market.task.edit.subject')" path="subject">
					<n-input v-model:value="form.subject" :placeholder="$t('market.task.edit.subjectPlaceholder')">
					</n-input>
				</n-form-item>
				<bt-more>
					<template #title>
						<span>{{ $t('template.inboxPlacement.create.advanced') }}</span>
					</template>
					<n-form-item :label="$t('template.inboxPlacement.create.seedGroup')" path="seed_group">
						<n-input v-model:value="form.seed_group" placeholder="default"></n-input>
					</n-form-item>
					<n-form-item :label="$t('template.inboxPlacement.create.timeout')" path="timeout_sec">
						<n-select v-model:value="form.timeout_sec" :options="timeoutOptions"></n-select>
					</n-form-item>
				</bt-more>
			</bt-form>
		</div>
		<preview-modal />
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { isObject } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { createPlacementTest } from '@/api/modules/inbox-placement'

import FromSelect from '@/views/market/task/components/FromSelect.vue'
import TemplateSelect from '@/views/market/task/components/TemplateSelect.vue'
import TemplatePreview from '@/views/template/components/TemplatePreview.vue'

interface ModalState {
	templateId?: number | null
	subject?: string
	// eslint-disable-next-line no-unused-vars
	onCreated?: (testId: string) => void
}

const { t } = useI18n()

const formRef = useTemplateRef('formRef')

const createForm = () => ({
	template_id: null as number | null,
	template_content: '',
	domain: null as string | null,
	addresser: null as string | null,
	full_name: '',
	subject: '',
	seed_group: 'default',
	timeout_sec: 900,
})

const form = reactive(createForm())

const timeoutOptions = computed(() =>
	[300, 600, 900, 1800].map(sec => ({
		label: t('template.inboxPlacement.create.minutes', { n: sec / 60 }),
		value: sec,
	}))
)

const rules: FormRules = {
	template_id: {
		required: true,
		type: 'number',
		trigger: 'change',
		message: t('api.form.validation.templateRequired'),
	},
	addresser: {
		required: true,
		trigger: 'change',
		message: t('api.form.validation.addresserRequired'),
	},
	subject: {
		required: true,
		trigger: 'blur',
		message: t('market.task.edit.subjectPlaceholder'),
	},
}

const [PreviewModal, previewModalApi] = useModal({
	component: TemplatePreview,
})

const handlePreview = () => {
	previewModalApi.setState({ html: form.template_content })
	previewModalApi.open()
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<ModalState>()
			if (state.templateId) form.template_id = state.templateId
			if (state.subject) form.subject = state.subject
		} else {
			Object.assign(form, createForm())
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		const res = await createPlacementTest({
			template_id: form.template_id || 0,
			addresser: form.addresser || '',
			full_name: form.full_name,
			subject: form.subject,
			seed_group: form.seed_group.trim() || 'default',
			timeout_sec: form.timeout_sec,
		})
		if (isObject<{ test_id: string }>(res)) {
			modalApi.getState<ModalState>().onCreated?.(res.test_id)
		}
	},
})
</script>
