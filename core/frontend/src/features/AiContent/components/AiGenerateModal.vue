<template>
	<modal :title="$t('template.aiContent.generate.title')" width="960" :footer="false">
		<div class="pt-12px generate-body">
			<div class="generate-form">
				<bt-form label-placement="top">
					<n-form-item v-if="hasCurrent" :label="$t('template.aiContent.generate.mode')" :show-feedback="false">
						<n-radio-group v-model:value="mode">
							<n-radio value="revise">{{ $t('template.aiContent.generate.modeRevise') }}</n-radio>
							<n-radio value="new">{{ $t('template.aiContent.generate.modeNew') }}</n-radio>
						</n-radio-group>
					</n-form-item>
					<n-form-item
						:label="mode === 'revise' ? $t('template.aiContent.generate.instruction') : $t('template.aiContent.generate.brief')">
						<n-input
							v-model:value="brief"
							type="textarea"
							:rows="7"
							:maxlength="4000"
							show-count
							:placeholder="
								mode === 'revise'
									? $t('template.aiContent.generate.instructionPlaceholder')
									: $t('template.aiContent.generate.briefPlaceholder')
							">
						</n-input>
					</n-form-item>
					<n-form-item :label="$t('template.aiContent.review.subject')">
						<n-input
							v-model:value="subject"
							:placeholder="$t('template.aiContent.generate.subjectPlaceholder')"
							clearable>
						</n-input>
					</n-form-item>
					<n-grid :cols="2" :x-gap="12">
						<n-form-item-gi :label="$t('template.aiContent.tone.label')">
							<n-select v-model:value="tone" :options="toneOptions" clearable></n-select>
						</n-form-item-gi>
						<n-form-item-gi :label="$t('template.aiContent.language.label')">
							<n-select v-model:value="language" :options="languageOptions"></n-select>
						</n-form-item-gi>
					</n-grid>
					<n-form-item :label="$t('template.aiContent.model.label')">
						<ai-model-select v-model:value="model"></ai-model-select>
					</n-form-item>
				</bt-form>
				<n-button type="primary" block :loading="generating" @click="generate">
					{{ result ? $t('template.aiContent.generate.regenerate') : $t('template.aiContent.generate.run') }}
				</n-button>
			</div>

			<div class="generate-preview">
				<n-spin :show="generating" class="h-full" content-class="h-full">
					<template #description>{{ $t('template.aiContent.generate.running') }}</template>
					<template v-if="result">
						<div class="preview-subject">
							<span class="text-desc">{{ $t('template.aiContent.review.subject') }}:</span>
							{{ result.subject || '--' }}
						</div>
						<div class="preview-frame">
							<bt-preview :value="result.html"></bt-preview>
						</div>
					</template>
					<div v-else class="placeholder">{{ $t('template.aiContent.generate.hint') }}</div>
				</n-spin>
			</div>
		</div>
		<div class="flex justify-between items-center mt-20px">
			<span class="text-desc text-12px">
				<template v-if="result">{{ result.model }} · {{ result.input_tokens + result.output_tokens }} tokens</template>
			</span>
			<div class="flex gap-12px">
				<n-button secondary @click="modalApi.close()">{{ $t('common.actions.cancel') }}</n-button>
				<n-button type="primary" :disabled="!result || generating" @click="apply">
					{{ $t('template.aiContent.generate.apply') }}
				</n-button>
			</div>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { confirm, isObject, Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { aiGenerateContent } from '@/api/modules/ai'
import { defaultAiLanguage, modelParams } from '../utils'
import { useAiLanguageOptions, useAiToneOptions } from '../useAiOptions'
import type { GenerateResult } from '../types'

import AiModelSelect from './AiModelSelect.vue'

/**
 * Modal state:
 * - subject: current subject, sent as a hint
 * - currentHtml: when not empty the user can revise it instead of writing from scratch
 * - onApply: receives the generated email
 */
interface GenerateState {
	subject?: string
	currentHtml?: string
	// eslint-disable-next-line no-unused-vars
	onApply?: (result: GenerateResult) => void
}

const { t, locale } = useI18n()

const languageOptions = useAiLanguageOptions()

const toneOptions = useAiToneOptions()

const mode = ref<'new' | 'revise'>('new')

const brief = ref('')

const subject = ref('')

const tone = ref<string | null>('professional')

const language = ref(defaultAiLanguage(locale.value))

const model = ref<string | null>(null)

const result = ref<GenerateResult | null>(null)

const generating = ref(false)

const getState = () => modalApi.getState<GenerateState>()

// Modal state is a plain object, so capture what the template needs into a ref on open.
const hasCurrent = ref(false)

const generate = async () => {
	if (!brief.value.trim()) {
		Message.error(
			mode.value === 'revise'
				? t('template.aiContent.generate.instructionRequired')
				: t('template.aiContent.generate.briefRequired')
		)
		return
	}
	generating.value = true
	try {
		const res = await aiGenerateContent({
			brief: brief.value,
			subject: subject.value,
			current_html: mode.value === 'revise' ? getState().currentHtml || '' : '',
			tone: tone.value || '',
			language: language.value,
			...modelParams(model.value),
		})
		if (isObject<GenerateResult>(res)) {
			result.value = res
		}
	} finally {
		generating.value = false
	}
}

const apply = () => {
	if (!result.value) return
	const doApply = () => {
		getState().onApply?.(result.value as GenerateResult)
		modalApi.close()
	}
	if (hasCurrent.value) {
		confirm({
			title: t('template.aiContent.generate.apply'),
			content: t('template.aiContent.generate.replaceConfirm'),
			onConfirm: doApply,
		})
	} else {
		doApply()
	}
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (!isOpen) return
		const state = getState()
		hasCurrent.value = !!state.currentHtml?.trim()
		mode.value = hasCurrent.value ? 'revise' : 'new'
		subject.value = state.subject || ''
		brief.value = ''
		result.value = null
	},
})
</script>

<style lang="scss" scoped>
.generate-body {
	display: flex;
	gap: 20px;
	height: 600px;
}

.generate-form {
	width: 340px;
	flex-shrink: 0;
	overflow-y: auto;
}

.generate-preview {
	flex: 1;
	min-width: 0;
	padding: 12px;
	border: 1px solid var(--color-border-1);
	border-radius: 6px;
	background-color: var(--color-bg-2);

	.preview-subject {
		margin-bottom: 8px;
		font-weight: 600;
		color: var(--color-text-1);
	}

	.preview-frame {
		height: calc(100% - 30px);
		background-color: #fff;
		border-radius: 4px;
	}

	.placeholder {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		padding: 0 24px;
		text-align: center;
		color: var(--color-text-3);
	}
}
</style>
