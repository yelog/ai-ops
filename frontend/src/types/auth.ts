export interface User {
  id: string
  username: string
  email: string
  role: 'admin' | 'operator' | 'viewer'
  created_at?: string
  last_login?: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  role?: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
}