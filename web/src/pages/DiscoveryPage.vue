<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../services/api'
import { useWorkspaceStore } from '../stores/workspace'

const store = useWorkspaceStore()
const query = ref('')
const message = ref('')

async function join(teamID: string) {
  message.value = ''
  try { await api.join(teamID); message.value = '申请已提交，负责人会在工作台处理。' } catch (error) { message.value = error instanceof Error ? error.message : '提交失败' }
}
onMounted(() => store.loadTeams())
</script>

<template>
  <section class="toolbar"><input v-model="query" placeholder="按小队名称或服务领域查找" @keyup.enter="store.loadTeams(query)" /><button class="primary" @click="store.loadTeams(query)">查找</button></section>
  <p v-if="message" class="notice">{{ message }}</p>
  <p v-if="store.error" class="error">{{ store.error }}</p>
  <div class="item-list">
    <article v-for="team in store.teams" :key="team.id" class="item-row">
      <div><span class="eyebrow">{{ team.status === 'recruiting' ? '正在招募' : '暂缓招募' }}</span><h2>{{ team.name }}</h2><p>{{ team.description }}</p></div>
      <button :disabled="team.status !== 'recruiting'" @click="join(team.id)">申请加入</button>
    </article>
  </div>
</template>
