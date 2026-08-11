export type ModerationStatus = 'pending' | 'approved' | 'rejected'

export interface ModerationRequest {
  id: string
  barrier_id: string
  reporter_id: string
  status: ModerationStatus
  moderator_id: string | null
  moderator_comment: string
  created_at: string
  updated_at: string
  reviewed_at: string | null
}

export interface ModerationQueueResponse {
  requests: ModerationRequest[]
  total: number
  limit: number
  offset: number
}

export interface ModerationActionRequest {
  comment: string
}

export interface ModerationActionResponse {
  id: string
  barrier_id: string
  status: ModerationStatus
  moderator_id: string
  reviewed_at: string
}