import api from './api'
import type { AuthResponse, LoginRequest, RegisterRequest, User } from '@/types/auth'

export const authService = {
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/api/v1/auth/login', data)
    return response.data
  },

  async register(data: RegisterRequest): Promise<User> {
    const response = await api.post<User>('/api/v1/auth/register', data)
    return response.data
  },

  async getProfile(): Promise<User> {
    const response = await api.get<User>('/api/v1/auth/profile')
    return response.data
  },
}