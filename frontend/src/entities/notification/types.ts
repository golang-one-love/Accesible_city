export type NotificationType = 
  | 'barrier_approved' 
  | 'barrier_resolved' 
  | 'barrier_nearby' 
  | 'new_poi'

export interface Notification {
  id: string
  type: NotificationType
  title: string
  message: string
  payload: Record<string, string>
  is_read: boolean
  created_at: string
  read_at: string | null
}

export interface NotificationsListResponse {
  notifications: Notification[]
  total: number
  limit: number
  offset: number
}

export interface UnreadCountResponse {
  count: number
}

export interface MarkAsReadRequest {
  notification_id: string
}