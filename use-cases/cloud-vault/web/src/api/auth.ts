import { client } from './client'

export interface User {
  id: string
  username: string
  email: string
  display_name: string
  role: string
  locale: string
  theme: string
  created_at: string
  updated_at: string
}

export interface Session {
  id: string
  user_id: string
  user_agent: string
  ip_address: string
  expires_at: string
  revoked_at?: string
  last_used_at: string
  created_at: string
}

export interface SetupRequest {
  username: string
  email: string
  password: string
}

export interface SetupResponse {
  user: User
}

export interface AuthStatusResponse {
  initialized: boolean
  authenticated: boolean
  user?: User
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  user: User
  session: Session
}

export interface UpdateProfileRequest {
  display_name?: string
  email?: string
  locale?: string
  theme?: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export const authAPI = {
  setup: (req: SetupRequest) => client.post<SetupResponse>('/api/v1/auth/setup', req),
  getStatus: () => client.get<AuthStatusResponse>('/api/v1/auth/status'),
  login: (req: LoginRequest) => client.post<LoginResponse>('/api/v1/auth/login', req),
  logout: () => client.post<void>('/api/v1/auth/logout', {}),
  getMe: () => client.get<User>('/api/v1/auth/me'),
  updateProfile: (req: UpdateProfileRequest) => client.put<User>('/api/v1/auth/me', req),
  changePassword: (req: ChangePasswordRequest) => client.post<void>('/api/v1/auth/change-password', req),
  listSessions: () => client.get<{ sessions: Session[] }>('/api/v1/auth/sessions'),
  revokeAllSessions: () => client.post<void>('/api/v1/auth/sessions/revoke-all', {}),
}
