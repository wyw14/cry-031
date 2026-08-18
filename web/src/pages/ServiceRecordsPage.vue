<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../services/api'
const records = ref<Array<{ id: string; activity_id: string; status: string; duration_minutes: number; updated_at: string }>>([])
const error = ref('')
onMounted(async () => { try { records.value = (await api.records(localStorage.getItem('demo_user') ?? 'u-volunteer')).items } catch (cause) { error.value = cause instanceof Error ? cause.message : '加载失败' } })
</script>

<template><p v-if="error" class="error">{{ error }}</p><div class="item-list"><article v-for="record in records" :key="record.id" class="item-row"><div><span class="eyebrow">{{ record.status }}</span><h2>{{ record.activity_id }}</h2><p>服务时长 {{ record.duration_minutes }} 分钟 · {{ new Date(record.updated_at).toLocaleString() }}</p></div><button disabled>查看版本</button></article><p v-if="!records.length" class="empty-panel">完成签到和服务后，记录会进入负责人复核。</p></div></template>
