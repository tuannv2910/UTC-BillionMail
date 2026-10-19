<template>
	<n-select
		v-model:value="value"
		filterable
		:options="options"
		:loading="loading"
		:placeholder="$t('template.aiContent.model.placeholder')">
		<template #empty>
			<div class="text-center text-desc">
				<div>{{ $t('template.aiContent.model.empty') }}</div>
				<n-button text type="primary" class="mt-4px" @click="goSettings">
					{{ $t('template.aiContent.model.configure') }}
				</n-button>
			</div>
		</template>
	</n-select>
</template>

<script lang="ts" setup>
import { isArray } from '@/utils'
import { getAiModels } from '@/api/modules/ai'
import { modelKey } from '../utils'
import type { AiModel } from '../types'

const emit = defineEmits<{ loaded: [count: number] }>()

const value = defineModel<string | null>('value', { default: null })

const router = useRouter()

const loading = ref(false)

const models = ref<AiModel[]>([])

const options = computed(() =>
	models.value.map(model => ({
		label: model.title || `${model.supplierName}/${model.modelId}`,
		value: modelKey(model),
	}))
)

const loadModels = async () => {
	loading.value = true
	try {
		const res = await getAiModels()
		models.value = isArray<AiModel>(res) ? res : []
		const keys = options.value.map(option => option.value)
		if (!value.value || !keys.includes(value.value)) {
			value.value = keys[0] ?? null
		}
		emit('loaded', models.value.length)
	} finally {
		loading.value = false
	}
}

const goSettings = () => {
	router.push({ name: 'AiModel' })
}

loadModels()
</script>
