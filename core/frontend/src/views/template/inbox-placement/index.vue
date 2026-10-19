<template>
	<div class="p-24px">
		<div class="bt-title mb-12px!">{{ $t('template.inboxPlacement.title') }}</div>
		<n-breadcrumb class="mb-12px">
			<n-breadcrumb-item>
				<router-link :to="{ name: 'template' }">{{ $t('template.template') }}</router-link>
			</n-breadcrumb-item>
			<n-breadcrumb-item>{{ $t('template.inboxPlacement.title') }}</n-breadcrumb-item>
		</n-breadcrumb>

		<n-alert v-if="checked && !status.enabled" type="warning" class="mb-16px">
			<template #header>{{ $t('template.inboxPlacement.notConfigured.title') }}</template>
			{{ $t('template.inboxPlacement.notConfigured.description') }}
			<code class="block mt-6px">INBOX_PLACEMENT_URL / INBOX_PLACEMENT_API_KEY / INBOX_PLACEMENT_ADMIN_TOKEN</code>
		</n-alert>
		<n-alert v-else-if="checked && !status.reachable" type="error" class="mb-16px">
			<template #header>{{ $t('template.inboxPlacement.unreachable.title') }}</template>
			{{ $t('template.inboxPlacement.unreachable.description') }}
		</n-alert>

		<n-tabs v-model:value="activeTab" type="line" animated @update:value="handleTabChange">
			<n-tab-pane name="tests" :tab="$t('template.inboxPlacement.tabs.tests')">
				<test-list v-if="checked" :enabled="status.enabled" />
			</n-tab-pane>
			<n-tab-pane name="seeds" :tab="$t('template.inboxPlacement.tabs.seeds')" display-directive="show:lazy">
				<seed-list v-if="checked" :admin-enabled="status.admin_seed && status.reachable" />
			</n-tab-pane>
		</n-tabs>
	</div>
</template>

<script lang="ts" setup>
import { isObject } from '@/utils'
import { getPlacementStatus } from '@/api/modules/inbox-placement'
import type { PlacementServiceStatus } from './interface'

import TestList from './components/TestList.vue'
import SeedList from './components/SeedList.vue'

const route = useRoute()

const router = useRouter()

const activeTab = ref(route.query.tab === 'seeds' ? 'seeds' : 'tests')

const checked = ref(false)

const status = reactive<PlacementServiceStatus>({
	enabled: false,
	reachable: false,
	admin_seed: false,
})

const handleTabChange = (tab: string) => {
	router.replace({ query: { ...route.query, tab } })
}

const getStatus = async () => {
	try {
		const res = await getPlacementStatus()
		if (isObject<PlacementServiceStatus>(res)) {
			Object.assign(status, res)
		}
	} finally {
		checked.value = true
	}
}

getStatus()
</script>
