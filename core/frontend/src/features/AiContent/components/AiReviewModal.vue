<template>
	<modal :title="$t('template.aiContent.review.title')" width="760" :footer="false">
		<div class="pt-12px">
			<n-grid :cols="24" :x-gap="16">
				<n-form-item-gi :span="11" :label="$t('template.aiContent.model.label')" :show-feedback="false">
					<ai-model-select v-model:value="model"></ai-model-select>
				</n-form-item-gi>
				<n-form-item-gi :span="7" :label="$t('template.aiContent.language.label')" :show-feedback="false">
					<n-select v-model:value="language" :options="languageOptions"></n-select>
				</n-form-item-gi>
				<n-form-item-gi :span="6" label=" " :show-feedback="false">
					<n-button type="primary" block :loading="reviewing" :disabled="fixing" @click="runReview">
						{{ result ? $t('template.aiContent.review.rerun') : $t('template.aiContent.review.run') }}
					</n-button>
				</n-form-item-gi>
			</n-grid>
			<n-form-item class="mt-12px" :label="$t('template.aiContent.review.subject')" :show-feedback="false">
				<n-input
					v-model:value="subject"
					:placeholder="$t('template.aiContent.review.subjectPlaceholder')"
					clearable>
				</n-input>
			</n-form-item>

			<n-spin :show="reviewing || fixing" class="mt-16px">
				<template #description>
					{{ fixing ? $t('template.aiContent.review.fixing') : $t('template.aiContent.review.running') }}
				</template>
				<ai-review-result v-if="result" :result="result"></ai-review-result>
				<div v-else class="placeholder">{{ $t('template.aiContent.review.hint') }}</div>
			</n-spin>

			<div class="flex justify-between items-center mt-20px">
				<span class="text-desc text-12px">{{ $t('template.aiContent.review.disclaimer') }}</span>
				<div class="flex gap-12px">
					<n-button
						v-if="canFix"
						type="primary"
						ghost
						:loading="fixing"
						:disabled="reviewing"
						@click="fixWithAi">
						{{ $t('template.aiContent.review.fix') }}
					</n-button>
					<n-button secondary @click="modalApi.close()">{{ $t('template.aiContent.close') }}</n-button>
				</div>
			</div>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { isObject, Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { aiGenerateContent, aiReviewContent } from '@/api/modules/ai'
import { buildFixInstruction, defaultAiLanguage, modelParams } from '../utils'
import { useAiLanguageOptions } from '../useAiOptions'
import type { GenerateResult, ReviewResult } from '../types'

import AiModelSelect from './AiModelSelect.vue'
import AiReviewResult from './AiReviewResult.vue'

/**
 * Modal state:
 * - subject, content: the email to review
 * - unsubscribeManaged: the platform appends the unsubscribe link at send time
 * - onApply: when set, "Fix with AI" rewrites the email and passes the result back
 */
interface ReviewState {
	subject?: string
	content?: string
	unsubscribeManaged?: boolean
	// eslint-disable-next-line no-unused-vars
	onApply?: (result: GenerateResult) => void
}

const { t, locale } = useI18n()

const languageOptions = useAiLanguageOptions()

const model = ref<string | null>(null)

const language = ref(defaultAiLanguage(locale.value))

const subject = ref('')

const content = ref('')

const result = ref<ReviewResult | null>(null)

const reviewing = ref(false)

const fixing = ref(false)

// Modal state is a plain object, so capture what the template needs into refs on open.
const canApply = ref(false)

const getState = () => modalApi.getState<ReviewState>()

const canFix = computed(() => canApply.value && !!result.value && result.value.issues.length > 0)

const runReview = async () => {
	if (!subject.value.trim() && !content.value.trim()) {
		Message.error(t('template.aiContent.review.empty'))
		return
	}
	reviewing.value = true
	try {
		const res = await aiReviewContent({
			subject: subject.value,
			content: content.value,
			language: language.value,
			unsubscribe_managed: !!getState().unsubscribeManaged,
			...modelParams(model.value),
		})
		if (isObject<ReviewResult>(res)) {
			result.value = res
		}
	} finally {
		reviewing.value = false
	}
}

const fixWithAi = async () => {
	if (!result.value) return
	fixing.value = true
	try {
		const res = await aiGenerateContent({
			brief: buildFixInstruction(result.value.issues),
			subject: subject.value,
			current_html: content.value,
			tone: '',
			language: language.value,
			...modelParams(model.value),
		})
		if (!isObject<GenerateResult>(res)) return
		getState().onApply?.(res)
		content.value = res.html
		subject.value = res.subject || subject.value
		Message.success(t('template.aiContent.review.fixed'))
	} finally {
		fixing.value = false
	}
	// Score the rewritten email so the user sees whether the fix worked.
	await runReview()
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = getState()
			subject.value = state.subject || ''
			content.value = state.content || ''
			canApply.value = !!state.onApply
			result.value = null
			runReview()
		} else {
			result.value = null
		}
	},
})
</script>

<style lang="scss" scoped>
.placeholder {
	padding: 48px 0;
	text-align: center;
	color: var(--color-text-3);
}
</style>
