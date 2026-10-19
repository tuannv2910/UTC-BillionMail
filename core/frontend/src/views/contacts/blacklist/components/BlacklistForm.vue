<template>
	<modal :title="isEdit ? $t('contacts.blacklist.form.editTitle') : $t('contacts.blacklist.form.addTitle')" width="560">
		<div class="pt-12px">
			<bt-form ref="formRef" :model="form" :rules="rules">
				<n-form-item v-if="!isEdit" :label="$t('contacts.blacklist.form.scope')">
					<n-radio-group v-model:value="form.scope">
						<n-radio v-for="option in scopeOptions" :key="option.value" :value="option.value">
							{{ option.label }}
						</n-radio>
					</n-radio-group>
				</n-form-item>
				<n-form-item :label="valueLabel" path="value">
					<n-input
						v-model:value="form.value"
						:disabled="isEdit"
						:placeholder="
							form.scope === BlacklistScope.Domain
								? $t('contacts.blacklist.form.domainPlaceholder')
								: $t('contacts.blacklist.form.emailPlaceholder')
						">
					</n-input>
				</n-form-item>
				<n-form-item :label="$t('contacts.blacklist.columns.reason')" path="reason">
					<n-select v-model:value="form.reason" :options="editableReasonOptions"></n-select>
				</n-form-item>

				<n-form-item v-if="!isEdit" :label="$t('contacts.blacklist.form.duration')">
					<n-select v-model:value="form.ttl_days" :options="durationOptions"></n-select>
				</n-form-item>
				<n-form-item v-else :label="$t('contacts.blacklist.form.duration')" path="expires_at">
					<div class="flex flex-col gap-8px w-full">
						<n-radio-group v-model:value="form.expire_mode">
							<n-radio value="keep">{{ $t('contacts.blacklist.form.expireKeep') }}</n-radio>
							<n-radio value="permanent">{{ $t('contacts.blacklist.duration.permanent') }}</n-radio>
							<n-radio value="until">{{ $t('contacts.blacklist.form.expireUntil') }}</n-radio>
						</n-radio-group>
						<n-date-picker
							v-if="form.expire_mode === 'until'"
							v-model:value="form.expires_at_ms"
							type="datetime"
							class="w-full"
							:is-date-disabled="isPast">
						</n-date-picker>
						<span v-if="form.expire_mode === 'keep'" class="text-desc text-12px">{{ currentExpiry }}</span>
					</div>
				</n-form-item>

				<n-form-item :label="$t('contacts.blacklist.columns.note')">
					<n-input v-model:value="form.note" type="textarea" :rows="3" :maxlength="255" show-count></n-input>
				</n-form-item>
			</bt-form>
		</div>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { formatTime } from '@/utils'
import { useModal } from '@/hooks/modal/useModal'
import { createBlacklist, updateBlacklist } from '@/api/modules/contacts/blacklist'
import { BlacklistScope, type BlacklistEntry } from '../types/base'
import { useBlacklistOptions } from '../useBlacklistOptions'

const { t } = useI18n()

const { scopeOptions, editableReasonOptions, durationOptions } = useBlacklistOptions()

const formRef = useTemplateRef('formRef')

const isEdit = ref(false)

const id = ref(0)

const originalExpiresAt = ref(0)

const form = reactive({
	scope: BlacklistScope.Email,
	value: '',
	reason: 10,
	ttl_days: -1,
	note: '',
	expire_mode: 'keep' as 'keep' | 'permanent' | 'until',
	expires_at_ms: null as number | null,
})

const valueLabel = computed(() =>
	form.scope === BlacklistScope.Domain ? t('contacts.blacklist.scope.domain') : t('contacts.blacklist.scope.email')
)

const currentExpiry = computed(() =>
	originalExpiresAt.value === 0
		? t('contacts.blacklist.duration.permanent')
		: t('contacts.blacklist.form.currentExpiry', { time: formatTime(originalExpiresAt.value) })
)

const isPast = (ts: number) => ts < Date.now() - 24 * 60 * 60 * 1000

const rules: FormRules = {
	value: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (!form.value.trim()) return new Error(t('contacts.blacklist.form.valueRequired'))
			return true
		},
	},
	expires_at: {
		trigger: 'change',
		validator: () => {
			if (form.expire_mode !== 'until') return true
			if (!form.expires_at_ms || form.expires_at_ms <= Date.now()) {
				return new Error(t('contacts.blacklist.form.expiresFuture'))
			}
			return true
		},
	},
}

const reset = () => {
	id.value = 0
	originalExpiresAt.value = 0
	form.scope = BlacklistScope.Email
	form.value = ''
	form.reason = 10
	form.ttl_days = -1
	form.note = ''
	form.expire_mode = 'keep'
	form.expires_at_ms = null
}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (!isOpen) {
			reset()
			return
		}
		const { row } = modalApi.getState<{ row: BlacklistEntry | null }>()
		isEdit.value = !!row
		if (row) {
			id.value = row.id
			originalExpiresAt.value = row.expires_at
			form.scope = row.scope
			form.value = row.value
			form.reason = row.reason
			form.note = row.note
			form.expires_at_ms = row.expires_at > 0 ? row.expires_at * 1000 : null
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		if (isEdit.value) {
			await updateBlacklist({
				id: id.value,
				reason: form.reason,
				note: form.note,
				expire_mode: form.expire_mode,
				expires_at: form.expires_at_ms ? Math.floor(form.expires_at_ms / 1000) : 0,
			})
		} else {
			await createBlacklist({
				scope: form.scope,
				value: form.value.trim(),
				reason: form.reason,
				note: form.note,
				ttl_days: form.ttl_days,
			})
		}
		modalApi.getState<{ refresh: () => void }>().refresh()
	},
})
</script>
