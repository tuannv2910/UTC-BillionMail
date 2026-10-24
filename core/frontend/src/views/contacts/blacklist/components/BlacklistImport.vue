<template>
	<modal
		:title="$t('contacts.blacklist.import.title')"
		width="600"
		:confirm-text="result ? $t('contacts.blacklist.import.done') : $t('contacts.blacklist.import.submit')">
		<div class="pt-12px">
			<template v-if="!result">
				<bt-form label-placement="top">
					<n-grid :cols="2" :x-gap="16">
						<n-form-item-gi :label="$t('contacts.blacklist.form.scope')">
							<n-radio-group v-model:value="form.scope">
								<n-radio v-for="option in scopeOptions" :key="option.value" :value="option.value">
									{{ option.label }}
								</n-radio>
							</n-radio-group>
						</n-form-item-gi>
						<n-form-item-gi :label="$t('contacts.blacklist.form.duration')">
							<n-select v-model:value="form.ttl_days" :options="durationOptions"></n-select>
						</n-form-item-gi>
					</n-grid>
					<n-form-item :label="$t('contacts.blacklist.columns.reason')">
						<n-select v-model:value="form.reason" :options="editableReasonOptions"></n-select>
					</n-form-item>
					<n-form-item :label="$t('contacts.blacklist.import.values')" :show-feedback="false">
						<div class="flex flex-col gap-8px w-full">
							<n-input
								v-model:value="text"
								type="textarea"
								:rows="8"
								:placeholder="$t('contacts.blacklist.import.valuesPlaceholder')">
							</n-input>
							<div class="flex justify-between items-center">
								<bt-file-upload mode="button" :is-upload="false" :accept="['txt', 'csv']" @change="handleFile">
									{{ $t('contacts.blacklist.import.upload') }}
								</bt-file-upload>
								<span class="text-desc">
									{{ $t('contacts.blacklist.import.count', { count: values.length, max: MAX_VALUES }) }}
								</span>
							</div>
						</div>
					</n-form-item>
					<n-form-item class="mt-12px" :label="$t('contacts.blacklist.columns.note')">
						<n-input v-model:value="form.note" :maxlength="255"></n-input>
					</n-form-item>
				</bt-form>
			</template>

			<template v-else>
				<n-alert :type="result.invalid.length ? 'warning' : 'success'" :show-icon="true">
					{{
						$t('contacts.blacklist.import.result', {
							inserted: result.inserted,
							updated: result.updated,
							invalid: result.invalid.length,
						})
					}}
				</n-alert>
				<div v-if="result.invalid.length" class="invalid-list">
					<div v-for="item in result.invalid" :key="item.value" class="invalid-item">
						<span class="font-600">{{ item.value || '--' }}</span>
						<span class="text-desc">{{ item.error }}</span>
					</div>
				</div>
			</template>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { readTextFile } from '@/utils/file'
import { UploadFileInfo } from 'naive-ui'
import { isObject, Message } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { importBlacklist } from '@/api/modules/contacts/blacklist'
import { BlacklistScope, parseValues, type BlacklistImportResult } from '../types/base'
import { useBlacklistOptions } from '../useBlacklistOptions'

const MAX_VALUES = 10000

const { t } = useI18n()

const { scopeOptions, editableReasonOptions, durationOptions } = useBlacklistOptions()

const form = reactive({
	scope: BlacklistScope.Email,
	reason: 1,
	ttl_days: -1,
	note: '',
})

const text = ref('')

const values = computed(() => parseValues(text.value))

const result = ref<BlacklistImportResult | null>(null)

const handleFile = async (file: UploadFileInfo) => {
	if (!file.file) return
	const content = await readTextFile(file.file)
	text.value = text.value ? `${text.value}\n${content}` : content
}

const reset = () => {
	form.scope = BlacklistScope.Email
	form.reason = 1
	form.ttl_days = -1
	form.note = ''
	text.value = ''
	result.value = null
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (!isOpen) reset()
	},
	onConfirm: async () => {
		// Second click closes the result view
		if (result.value) return true

		if (values.value.length === 0) {
			Message.error(t('contacts.blacklist.import.empty'))
			return false
		}
		if (values.value.length > MAX_VALUES) {
			Message.error(t('contacts.blacklist.import.tooMany', { max: MAX_VALUES }))
			return false
		}

		const res = await importBlacklist({ ...form, values: values.value })
		modalApi.getState<{ refresh: () => void }>().refresh()
		if (isObject<BlacklistImportResult>(res) && res.invalid.length > 0) {
			result.value = res
			return false
		}
		return true
	},
})
</script>

<style lang="scss" scoped>
.invalid-list {
	max-height: 280px;
	margin-top: 12px;
	overflow-y: auto;
	border: 1px solid var(--color-border-1);
	border-radius: 4px;
}

.invalid-item {
	display: flex;
	justify-content: space-between;
	gap: 12px;
	padding: 8px 12px;
	border-bottom: 1px solid var(--color-border-1);

	&:last-child {
		border-bottom: none;
	}
}
</style>
