<template>
	<modal :title="title" :width="1200">
		<bt-form class="pt-8px" :model="form">
			<n-grid :cols="24" :x-gap="24">
				<n-form-item-gi :span="18" label="Name" path="temp_name">
					<n-input
						v-model:value="form.temp_name"
						:placeholder="$t('market.template.add.namePlaceholder')">
					</n-input>
				</n-form-item-gi>
				<n-form-item-gi :span="6" label="Type" path="add_type">
					<n-select v-model:value="form.add_type" :disabled="isEdit" :options="typeOptions">
					</n-select>
				</n-form-item-gi>
			</n-grid>
			<n-form-item :label="typeLabel" :show-feedback="false">
				<div class="flex flex-col flex-1 h-580px">
					<div class="flex justify-end gap-12px mb-8px">
						<n-button v-if="form.add_type === 0" type="primary" ghost @click="handleAiGenerate">
							<i class="i-mdi:creation-outline mr-4px"></i>
							{{ $t('template.aiContent.actions.generate') }}
						</n-button>
						<n-button type="primary" ghost @click="handleAiReview">
							<i class="i-mdi:shield-check-outline mr-4px"></i>
							{{ $t('template.aiContent.actions.review') }}
						</n-button>
						<div>
							<bt-file-upload
								mode="button"
								button-type="primary"
								:is-upload="false"
								:accept="['html']"
								@change="handleFileUpload">
								{{ $t('market.template.uploadHtml') }}
							</bt-file-upload>
						</div>
						<n-button type="primary" ghost @click="handlePreview">
							{{ $t('common.actions.preview') }}
						</n-button>
					</div>
					<div class="flex-1">
						<template v-if="form.add_type === 1">
							<email-editor v-model:config="form.drag_data" v-model:html="form.html_content">
							</email-editor>
						</template>
						<template v-if="form.add_type === 0">
							<bt-editor v-model:value="form.html_content" language="html"> </bt-editor>
						</template>
					</div>
				</div>
			</n-form-item>
		</bt-form>
		<ai-review-modal />
		<ai-generate-modal />
	</modal>
</template>

<script lang="ts" setup>
import { UploadFileInfo } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { addTemplate, updateTemplate } from '@/api/modules/market/template'
import AiReviewModalComponent from '@/features/AiContent/components/AiReviewModal.vue'
import AiGenerateModalComponent from '@/features/AiContent/components/AiGenerateModal.vue'
import type { GenerateResult } from '@/features/AiContent/types'
import type { Template } from '../interface'

const EmailEditor = defineAsyncComponent(() => import('@/features/EmailEditor/index.vue'))

const isEdit = ref(false)

const title = computed(() => {
	return isEdit.value ? 'Edit template' : 'New template'
})

const id = ref(0)

const form = reactive({
	temp_name: '',
	add_type: 0,
	html_content: '',
	drag_data: '',
})

const typeOptions = [
	{ label: 'HTML', value: 0 },
	{ label: 'Drag', value: 1 },
]

const typeLabel = computed(() => {
	return typeOptions.find(item => item.value === form.add_type)?.label || ''
})

const [AiReviewModal, aiReviewModalApi] = useModal({ component: AiReviewModalComponent })

const [AiGenerateModal, aiGenerateModalApi] = useModal({ component: AiGenerateModalComponent })

const applyAiContent = (result: GenerateResult) => {
	form.html_content = result.html
}

// Templates have no subject; the platform appends the unsubscribe link for tasks by default.
const handleAiReview = () => {
	aiReviewModalApi.setState({
		subject: '',
		content: form.html_content,
		unsubscribeManaged: true,
		// Drag templates are rebuilt from their design data, so AI rewrites only apply to HTML templates.
		onApply: form.add_type === 0 ? applyAiContent : undefined,
	})
	aiReviewModalApi.open()
}

const handleAiGenerate = () => {
	aiGenerateModalApi.setState({
		subject: '',
		currentHtml: form.html_content,
		onApply: applyAiContent,
	})
	aiGenerateModalApi.open()
}

const handleFileUpload = (file: UploadFileInfo) => {
	const reader = new FileReader()
	reader.onload = e => {
		form.html_content = e.target?.result as string
	}
	if (file.file) {
		reader.readAsText(file.file)
	}
}

const handlePreview = () => {
	// eslint-disable-next-line no-unused-vars
	const state = modalApi.getState<{ preview: (content: string) => void }>()
	if (state.preview) {
		state.preview(form.html_content)
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ isEdit: boolean; type: number; row: Template }>()
			const { row } = state
			isEdit.value = state.isEdit
			if (state.type)  {
				form.add_type = state.type
			}
			if (row) {
				id.value = row.id
				form.temp_name = row.temp_name
				form.add_type = row.add_type
				form.html_content = row.html_content
				form.drag_data = row.drag_data
			}
		} else {
			id.value = 0
			form.temp_name = ''
			form.add_type = 0
			form.html_content = ''
			form.drag_data = ''
		}
	},
	onConfirm: async () => {
		if (isEdit.value) {
			await updateTemplate({ id: id.value, ...toRaw(form) })
		} else {
			await addTemplate(toRaw(form))
		}
		const state = modalApi.getState<{ refresh: () => void }>()
		state.refresh()
	},
})
</script>

<style lang="scss" scoped></style>
