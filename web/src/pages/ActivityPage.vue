<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../services/api'
import { useWorkspaceStore } from '../stores/workspace'
import type { ActivityDetail } from '../types/domain'

const store = useWorkspaceStore()
const detail = ref<ActivityDetail | null>(null)
const message = ref('')
async function open(id: string) { detail.value = await api.activity(id) }
async function claim(id: string) { try { await api.claim(id); message.value = '岗位已认领；满员时系统会自动进入候补。'; if (detail.value) await open(detail.value.activity.id) } catch (error) { message.value = error instanceof Error ? error.message : '认领失败' } }
onMounted(() => store.loadActivities())
</script>

<template>
  <p v-if="message" class="notice">{{ message }}</p>
  <div class="split wide">
    <section class="item-list"><article v-for="activity in store.activities" :key="activity.id" class="item-row selectable" @click="open(activity.id)"><div><span class="eyebrow">{{ activity.status }}</span><h2>{{ activity.title }}</h2><p>{{ new Date(activity.start_at).toLocaleString() }} · {{ activity.location }}</p></div><span>查看岗位</span></article></section>
    <section v-if="detail" class="detail"><h2>{{ detail.activity.title }}</h2><p>{{ detail.activity.description }}</p><p><strong>物资：</strong>{{ detail.activity.material_needs.join('、') || '无需自备' }}</p><div v-for="slot in detail.slots" :key="slot.id" class="slot"><div><strong>{{ slot.name }}</strong><small>{{ slot.description }} · 名额 {{ slot.capacity }}</small></div><button @click="claim(slot.id)">认领</button></div></section>
    <section v-else class="empty-panel">选择一个活动查看岗位与现场风险</section>
  </div>
</template>
