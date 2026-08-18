<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../services/api'
import { useWorkspaceStore } from '../stores/workspace'
const store = useWorkspaceStore()
const message = ref('')
async function acknowledge(id: string) { try { await api.handoffAck(id); message.value = '交接已确认，责任人变更已记录。'; await store.loadDashboard() } catch (cause) { message.value = cause instanceof Error ? cause.message : '确认失败' } }
onMounted(() => store.loadDashboard())
</script>

<template><p v-if="message" class="notice">{{ message }}</p><div class="item-list"><article v-for="item in store.dashboard?.overdue_handoffs" :key="item.id" class="item-row warning"><div><span class="eyebrow">已过期</span><h2>风险交接待确认</h2><p>{{ item.note }} · 截止 {{ new Date(item.due_at).toLocaleString() }}</p></div><button @click="acknowledge(item.id)">确认接手</button></article><p v-if="!store.dashboard?.overdue_handoffs.length" class="empty-panel">当前没有过期交接</p></div></template>
