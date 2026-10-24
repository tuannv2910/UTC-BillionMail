<template>
	<template v-if="!show">
		<base-label>{{ $t('editor.allSides') }}</base-label>
		<slot name="all"></slot>
	</template>
	<template v-else>
		<n-grid x-gap="8" y-gap="8" :cols="2">
			<n-gi>
				<base-label>{{ $t(topLabel) }}</base-label>
				<slot name="top"></slot>
			</n-gi>
			<n-gi>
				<base-label>{{ $t(rightLabel) }}</base-label>
				<slot name="right"></slot>
			</n-gi>
			<n-gi>
				<base-label>{{ $t(leftLabel) }}</base-label>
				<slot name="left"></slot>
			</n-gi>
			<n-gi>
				<base-label>{{ $t(bottomLabel) }}</base-label>
				<slot name="bottom"></slot>
			</n-gi>
		</n-grid>
	</template>
</template>

<script lang="ts" setup>
import BaseLabel from './BaseLabel.vue'

interface Props {
	show: boolean
	topLabel?: string
	rightLabel?: string
	leftLabel?: string
	bottomLabel?: string
}

withDefaults(defineProps<Props>(), {
	show: false,
	topLabel: 'editor.top',
	rightLabel: 'editor.right',
	leftLabel: 'editor.left',
	bottomLabel: 'editor.bottom',
})
</script>
