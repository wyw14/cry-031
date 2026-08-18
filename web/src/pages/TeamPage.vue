<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../services/api'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const notices = ref<Array<{ id: string; kind: string; message: string; created_at: string }>>([])
onMounted(async () => { await store.loadTeams(); notices.value = (await api.notices()).items })
</script>

<template>
  <div class="split">
    <section><span class="eyebrow">{{ store.selectedTeam?.status }}</span><h2>{{ store.selectedTeam?.name ?? '滨河同行小队' }}</h2><p>{{ store.selectedTeam?.description }}</p><dl><dt>服务领域</dt><dd>助老服务、环境维护</dd><dt>公开范围</dt><dd>社区公开</dd><dt>招募方式</dt><dd>公开申请或邀请码</dd></dl></section>
    <section><h2>公告与提醒</h2><ul class="timeline"><li v-for="notice in notices" :key="notice.id"><strong>{{ notice.kind }}</strong><span>{{ notice.message }}</span><time>{{ new Date(notice.created_at).toLocaleString() }}</time></li><li v-if="!notices.length" class="empty">暂无新的站内提醒</li></ul></section>
  </div>
</template>
