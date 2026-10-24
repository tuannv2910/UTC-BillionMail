<template>
	<container :label="label || $t('editor.textAlign')">
		<template #value>
			<n-radio-group v-model:value="value">
				<n-radio-button value="left">{{ $t('editor.alignLeft') }}</n-radio-button>
				<n-radio-button value="center">{{ $t('editor.alignCenter') }}</n-radio-button>
				<n-radio-button value="right">{{ $t('editor.alignRight') }}</n-radio-button>
			</n-radio-group>
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

const value = defineModel<string>('value')
</script>
