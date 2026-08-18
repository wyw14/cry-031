<script setup lang="ts">
import { computed, ref } from 'vue'
import DiscoveryPage from './pages/DiscoveryPage.vue'
import TeamPage from './pages/TeamPage.vue'
import ActivityPage from './pages/ActivityPage.vue'
import ServiceRecordsPage from './pages/ServiceRecordsPage.vue'
import HandoffPage from './pages/HandoffPage.vue'
import ProfilePage from './pages/ProfilePage.vue'
import LeaderDashboardPage from './pages/LeaderDashboardPage.vue'

const sections = [
  { id: 'discovery', label: '社区发现', component: DiscoveryPage },
  { id: 'team', label: '小队主页', component: TeamPage },
  { id: 'activities', label: '活动与岗位', component: ActivityPage },
  { id: 'records', label: '服务记录', component: ServiceRecordsPage },
  { id: 'handoff', label: '交接清单', component: HandoffPage },
  { id: 'profile', label: '个人档案', component: ProfilePage },
  { id: 'leader', label: '负责人工作台', component: LeaderDashboardPage },
]
const active = ref('discovery')
const current = computed(() => sections.find((section) => section.id === active.value)?.component ?? DiscoveryPage)

function switchRole(role: 'member' | 'captain') {
  localStorage.setItem('demo_role', role)
  localStorage.setItem('demo_user', role === 'captain' ? 'u-captain' : 'u-volunteer')
  window.location.reload()
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark">邻</span><div><strong>邻里同行</strong><small>社区志愿协作</small></div></div>
      <nav>
        <button v-for="section in sections" :key="section.id" :class="{ active: active === section.id }" @click="active = section.id">
          {{ section.label }}
        </button>
      </nav>
      <div class="role-switch">
        <span>演示身份</span>
        <button @click="switchRole('member')">居民成员</button>
        <button @click="switchRole('captain')">小队负责人</button>
      </div>
    </aside>
    <main>
      <header><div><p>滨河社区 · 2026 夏季服务</p><h1>{{ sections.find((item) => item.id === active)?.label }}</h1></div><span class="status-dot">本地服务可用</span></header>
      <component :is="current" />
    </main>
  </div>
</template>
