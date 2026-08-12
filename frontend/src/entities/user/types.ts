export type UserRole = 'user' | 'volunteer' | 'moderator' | 'business_owner' | 'admin'

export interface User {
  id: string
  email: string
  nickname: string
  role: UserRole
  is_active: boolean
  created_at: string
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface RefreshRequest {
  refresh_token: string
}

export interface ValidateResponse {
  user_id: string
  email: string
  role: string
  exp: number
}

export type RoleApplicationStatus = 'pending' | 'approved' | 'rejected'

export interface RoleApplication {
  id: string
  user_id: string
  requested_role: UserRole
  comment: string
  status: RoleApplicationStatus
  created_at: string
  reviewed_at: string
}