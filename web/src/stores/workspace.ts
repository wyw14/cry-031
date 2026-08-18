import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '../services/api'
import type { Activity, Dashboard, Team } from '../types/domain'

export const useWorkspaceStore = defineStore('workspace', () => {
  const teams = ref<Team[]>([])
  const activities = ref<Activity[]>([])
  const dashboard = ref<Dashboard | null>(null)
  const selectedTeamID = ref('team-riverside')
  const loading = ref(false)
  const error = ref('')
  const selectedTeam = computed(() => teams.value.find((item) => item.id === selectedTeamID.value))

  async function run(action: () => Promise<void>) {
    loading.value = true
    error.value = ''
    try { await action() } catch (cause) { error.value = cause instanceof Error ? cause.message : '加载失败' } finally { loading.value = false }
  }

  async function loadTeams(query = '') {
    await run(async () => { teams.value = (await api.teams(query)).items })
  }

  async function loadActivities() {
    await run(async () => { activities.value = (await api.activities(selectedTeamID.value)).items })
  }

  async function loadDashboard() {
    await run(async () => { dashboard.value = await api.dashboard(selectedTeamID.value) })
  }

  return { teams, activities, dashboard, selectedTeamID, selectedTeam, loading, error, loadTeams, loadActivities, loadDashboard }
})
