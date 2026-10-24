<template>
	<bt-modal v-model:show="show" :title="$t('mailbox.form.importTitle')" :width="480" @confirm="onConfirm">
		<div class="py-16px">
			<bt-file-upload :is-upload="false" :accept="['csv']" @change="handleChangeFile">
			</bt-file-upload>
		</div>
	</bt-modal>
</template>

<script lang="ts" setup>
import { readTextFile } from '@/utils/file'
import { UploadFileInfo } from 'naive-ui'
import { Message } from '@/utils'
import { importMailbox } from '@/api/modules/mailbox'

const emit = defineEmits<{
	refresh: []
}>()

const { t } = useI18n()

const show = ref(false)

const form = reactive({
	file_data: '',
	file_type: '',
})

const handleChangeFile = async (file: UploadFileInfo) => {
	if (!file.file) return
	form.file_type = file.name.substring(file.name.lastIndexOf('.') + 1)
	// detects UTF-8/BOM, GB18030 and Windows-1252 (#153, #212)
	form.file_data = await readTextFile(file.file)
}

const onConfirm = async () => {
	if (!form.file_data || !form.file_type) {
		Message.error(t('mailbox.validation.fileRequired'))
		return false
	}
	await importMailbox(toRaw(form))
	emit('refresh')
}

const resetForm = () => {
	form.file_data = ''
	form.file_type = ''
}

const open = () => {
	resetForm()
	show.value = true
}

defineExpose({
	open,
})
</script>

<style lang="scss" scoped></style>
