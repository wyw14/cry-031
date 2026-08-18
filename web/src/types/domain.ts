export type Team = {
  id: string
  community_id: string
  name: string
  description: string
  status: 'recruiting' | 'paused' | 'closed'
  visibility: 'public' | 'community' | 'private'
  service_area_ids: string[]
}

export type Activity = {
  id: string
  team_id: string
  title: string
  description: string
  status: 'draft' | 'published' | 'cancelled' | 'completed'
  start_at: string
  end_at: string
  location: string
  material_needs: string[]
}

export type RoleSlot = {
  id: string
  activity_id: string
  name: string
  description: string
  capacity: number
  status: string
}

export type Risk = {
  id: string
  title: string
  severity: 'low' | 'medium' | 'high'
  description: string
  owner_id: string
  status: string
  due_at: string
}

export type ActivityDetail = { activity: Activity; slots: RoleSlot[]; claims: unknown[]; risks: Risk[] }

export type Page<T> = { items: T[]; page: number; page_size: number; total: number }

export type Dashboard = {
  pending_reviews: number
  missing_positions: number
  overdue_handoffs: Array<{ id: string; note: string; due_at: string; to_user_id: string }>
  open_risks: Risk[]
  completed_recaps: Activity[]
}
