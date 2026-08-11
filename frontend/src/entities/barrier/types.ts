export type BarrierType = 
  | 'high_curb'
  | 'broken_elevator'
  | 'closed_sidewalk'
  | 'stairs'
  | 'pothole'
  | 'uneven_surface'
  | 'parked_car'

export type BarrierStatus = 'pending' | 'approved' | 'rejected' | 'resolved'

export type Severity = 1 | 2 | 3 | 4 | 5

export interface Coordinates {
  latitude: number
  longitude: number
}

export interface BarrierPhoto {
  id: string
  barrier_id: string
  s3_key: string
  original_filename: string
  content_type: string
  size_bytes: number
  uploaded_by: string
  created_at: string
  presigned_url?: string
}

export interface Barrier {
  id: string
  type: BarrierType
  coordinates: Coordinates
  description: string
  severity: Severity
  status: BarrierStatus
  reporter_id: string | null
  moderator_id: string | null
  created_at: string
  updated_at: string
  approved_at: string | null
  resolved_at: string | null
  photos: BarrierPhoto[]
  confirmations_count: number
  complaints_count: number
}

export interface BarrierListResponse {
  barriers: Barrier[]
  total: number
  limit: number
  offset: number
}

export interface CreateBarrierRequest {
  type: BarrierType
  coordinates: Coordinates
  description?: string
  severity?: Severity
}

export interface ConfirmBarrierResponse {
  id: string
  barrier_id: string
  user_id: string
  created_at: string
}

export interface ComplainBarrierRequest {
  reason: string
}

export interface ComplainBarrierResponse {
  id: string
  barrier_id: string
  user_id: string
  reason: string
  created_at: string
}