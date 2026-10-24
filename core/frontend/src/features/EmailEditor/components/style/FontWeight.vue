<template>
	<container :label="$t('editor.fontWeight')">
		<template #value>
			<div class="w-100px">
				<n-select v-model:value="value" :options="options.map(o => ({ ...o, label: $t(o.label) }))"></n-select>
			</div>
		</template>
	</container>
</template>

<script lang="ts" setup>
import Container from './BaseContainer.vue'

interface Props {
	label?: string
}

withDefaults(defineProps<Props>(), {
	label: '',
})

const value = defineModel<string>('value', {
	default: '',
})

const options = [
	{ label: 'editor.normal', value: 'normal' },
	{ label: 'editor.bold', value: 'bold' },
]

if (!value.value) {
	value.value = options[0].value
}
</script>
