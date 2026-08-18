<script setup lang="ts">
import { onMounted } from 'vue'
import { useWorkspaceStore } from '../stores/workspace'
const store = useWorkspaceStore()
onMounted(() => store.loadDashboard())
</script>

<template><p v-if="store.error" class="error">{{ store.error }}</p><div class="metrics"><article><span>待确认记录</span><strong>{{ store.dashboard?.pending_reviews ?? 0 }}</strong></article><article><span>缺岗人数</span><strong>{{ store.dashboard?.missing_positions ?? 0 }}</strong></article><article><span>过期交接</span><strong>{{ store.dashboard?.overdue_handoffs.length ?? 0 }}</strong></article><article><span>开放风险</span><strong>{{ store.dashboard?.open_risks.length ?? 0 }}</strong></article></div><section><h2>活动结束复盘</h2><div class="item-list"><article v-for="activity in store.dashboard?.completed_recaps" :key="activity.id" class="item-row"><div><h2>{{ activity.title }}</h2><p>{{ activity.location }} · {{ new Date(activity.end_at).toLocaleString() }}</p></div><button>查看复盘</button></article></div></section></template>
