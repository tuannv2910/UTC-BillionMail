<template>
	<div class="ai-review-result">
		<div class="overview">
			<n-progress
				type="circle"
				:percentage="(result.score / result.max_score) * 100"
				:color="scoreColor(result.score)"
				:stroke-width="8"
				class="score-circle">
				<div class="score-text" :style="{ color: scoreColor(result.score) }">
					{{ result.score }}<span class="score-max">/{{ result.max_score }}</span>
				</div>
			</n-progress>
			<div class="overview-body">
				<div class="flex items-center gap-8px">
					<n-tag :type="verdictTagType(result.verdict)" :bordered="false">
						{{ $t(`template.aiContent.verdict.${result.verdict}`) }}
					</n-tag>
					<span class="text-desc">
						{{ $t('template.aiContent.review.issueCount', { count: result.issues.length }) }}
					</span>
				</div>
				<div v-if="result.summary" class="summary">{{ result.summary }}</div>
				<div class="meta">
					{{ result.model }} · {{ result.input_tokens + result.output_tokens }} tokens
					<template v-if="result.dropped_issues > 0">
						· {{ $t('template.aiContent.review.dropped', { count: result.dropped_issues }) }}
					</template>
				</div>
			</div>
		</div>

		<n-empty
			v-if="result.issues.length === 0"
			class="py-24px"
			:description="$t('template.aiContent.review.noIssues')">
		</n-empty>

		<div v-else class="issue-list">
			<div v-for="(issue, index) in issues" :key="index" class="issue" :class="`issue-${issue.severity}`">
				<div class="issue-header">
					<n-tag size="small" :type="severityTagType(issue.severity)" :bordered="false">
						{{ $t(`template.aiContent.severity.${issue.severity}`) }}
					</n-tag>
					<span class="issue-category">{{ $t(`template.aiContent.category.${issue.category}`) }}</span>
					<n-tag size="small" :bordered="false">
						{{ $t(`template.aiContent.location.${issue.location}`) }}
					</n-tag>
				</div>
				<div v-if="issue.excerpt" class="issue-excerpt">“{{ issue.excerpt }}”</div>
				<div class="issue-reason">{{ issue.reason }}</div>
				<div v-if="issue.suggestion" class="issue-suggestion">
					<i class="i-mdi:lightbulb-on-outline"></i>
					<span>{{ issue.suggestion }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { scoreColor, severityTagType, sortIssues, verdictTagType } from '../utils'
import type { ReviewResult } from '../types'

const props = defineProps<{ result: ReviewResult }>()

const issues = computed(() => sortIssues(props.result.issues))
</script>

<style lang="scss" scoped>
.ai-review-result {
	display: flex;
	flex-direction: column;
	gap: 16px;
}

.overview {
	display: flex;
	align-items: center;
	gap: 20px;
	padding: 16px;
	border-radius: 6px;
	background-color: var(--color-bg-2);

	.score-circle {
		width: 96px;
		flex-shrink: 0;
	}

	.score-text {
		font-size: 26px;
		font-weight: 600;
		line-height: 1;
	}

	.score-max {
		font-size: 13px;
		color: var(--color-text-3);
	}

	.overview-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
	}

	.summary {
		color: var(--color-text-1);
		line-height: 1.5;
	}

	.meta {
		font-size: 12px;
		color: var(--color-text-3);
	}
}

.issue-list {
	display: flex;
	flex-direction: column;
	gap: 10px;
}

.issue {
	padding: 12px 14px;
	border: 1px solid var(--color-border-1);
	border-left-width: 3px;
	border-radius: 4px;

	&.issue-critical,
	&.issue-high {
		border-left-color: var(--color-error-1);
	}

	&.issue-medium {
		border-left-color: var(--color-warning-1);
	}

	&.issue-low {
		border-left-color: var(--color-border-2);
	}
}

.issue-header {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 8px;

	.issue-category {
		font-weight: 600;
		color: var(--color-text-1);
	}
}

.issue-excerpt {
	margin-bottom: 6px;
	padding: 6px 10px;
	border-radius: 4px;
	background-color: var(--color-bg-2);
	font-family: monospace;
	word-break: break-word;
	color: var(--color-text-1);
}

.issue-reason {
	line-height: 1.5;
	color: var(--color-text-2);
}

.issue-suggestion {
	display: flex;
	gap: 6px;
	margin-top: 6px;
	line-height: 1.5;
	color: var(--color-primary-1);

	i {
		flex-shrink: 0;
		margin-top: 3px;
	}
}
</style>
