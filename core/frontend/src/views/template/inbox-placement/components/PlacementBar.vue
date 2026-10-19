<template>
	<div class="placement-bar">
		<div class="placement-bar__track" :style="{ height: `${height}px` }">
			<template v-if="segments.length">
				<n-tooltip v-for="seg in segments" :key="seg.key" trigger="hover">
					<template #trigger>
						<div
							class="placement-bar__segment"
							:style="{ width: `${seg.percent}%`, backgroundColor: seg.color }"></div>
					</template>
					{{ $t(`template.inboxPlacement.placement.${seg.key}`) }}: {{ seg.count }}
				</n-tooltip>
			</template>
		</div>
		<div v-if="legend && segments.length" class="placement-bar__legend">
			<span v-for="seg in segments" :key="seg.key" class="placement-bar__legend-item">
				<i class="placement-bar__dot" :style="{ backgroundColor: seg.color }"></i>
				{{ $t(`template.inboxPlacement.placement.${seg.key}`) }}
				<b>{{ seg.count }}</b>
			</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { buildSegments } from '../placement'
import type { PlacementSummary } from '../interface'

const props = defineProps({
	summary: {
		type: Object as PropType<PlacementSummary>,
		default: undefined,
	},
	height: {
		type: Number,
		default: 8,
	},
	legend: {
		type: Boolean,
		default: false,
	},
})

const segments = computed(() => buildSegments(props.summary))
</script>

<style lang="scss" scoped>
.placement-bar {
	&__track {
		display: flex;
		width: 100%;
		overflow: hidden;
		border-radius: 999px;
		background-color: var(--color-border-1, #ebeef5);
	}

	&__segment {
		height: 100%;
		transition: width 0.3s ease;

		& + & {
			border-left: 1px solid rgba(255, 255, 255, 0.6);
		}
	}

	&__legend {
		display: flex;
		flex-wrap: wrap;
		gap: 6px 16px;
		margin-top: 10px;
		font-size: 12px;
		color: var(--color-text-2, #606266);
	}

	&__legend-item {
		display: inline-flex;
		align-items: center;
		gap: 6px;

		b {
			color: var(--color-text-1, #303133);
		}
	}

	&__dot {
		display: inline-block;
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}
}
</style>
