import type { Activity, ActivityDetail, Dashboard, Page, Team } from '../types/domain'

const base = import.meta.env.VITE_API_BASE ?? '/api/v1'

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  headers.set('X-User-ID', localStorage.getItem('demo_user') ?? 'u-volunteer')
  headers.set('X-User-Role', localStorage.getItem('demo_role') ?? 'member')
  const response = await fetch(`${base}${path}`, { ...init, headers })
  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: { message: '请求失败' } }))
    throw new Error(payload.error?.message ?? `请求失败 (${response.status})`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const api = {
  teams: (query = '') => request<Page<Team>>(`/discovery/teams?page=1&page_size=50&query=${encodeURIComponent(query)}`),
  activities: (teamID = '') => request<Page<Activity>>(`/activities?page=1&page_size=50&team_id=${encodeURIComponent(teamID)}`),
  activity: (id: string) => request<ActivityDetail>(`/activities/${id}`),
  claim: (slotID: string) => request(`/slots/${slotID}/claims`, { method: 'POST', body: '{}' }),
  join: (teamID: string, inviteCode = '') => request(`/teams/${teamID}/join`, { method: 'POST', body: JSON.stringify({ invite_code: inviteCode }) }),
  notices: () => request<{ items: Array<{ id: string; kind: string; message: string; created_at: string }> }>('/notices'),
  records: (userID: string) => request<Page<{ id: string; activity_id: string; status: string; duration_minutes: number; updated_at: string }>>(`/profiles/${userID}/records?page=1&page_size=50`),
  dashboard: (teamID: string) => request<Dashboard>(`/teams/${teamID}/dashboard`),
  handoffAck: (id: string) => request(`/handoffs/${id}/acknowledge`, { method: 'POST', body: '{}' }),
  downloadCertificate: async (userID: string) => {
    const headers = new Headers()
    headers.set('X-User-ID', localStorage.getItem('demo_user') ?? 'u-volunteer')
    headers.set('X-User-Role', localStorage.getItem('demo_role') ?? 'member')
    const response = await fetch(`${base}/profiles/${userID}/certificate`, { headers })
    if (!response.ok) throw new Error('服务证明导出失败')
    const url = URL.createObjectURL(await response.blob())
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'service-certificate.csv'
    anchor.click()
    URL.revokeObjectURL(url)
  },
}
