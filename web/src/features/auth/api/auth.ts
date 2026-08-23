import { apiClient } from '../../../shared/api/client'
import type {
  AuthResponse,
  LoginRequest,
  RefreshResponse,
  RegisterRequest,
} from '../types'

export async function register(
  data: RegisterRequest,): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>(
      '/auth/register',
      data,
    )
  return response.data
}

export async function login(
  data: LoginRequest,): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>(
      '/auth/login',
      data,
    )
  return response.data
}

export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout')
}

export async function refreshSession(): Promise<RefreshResponse> {
  const response = await apiClient.post<RefreshResponse>(
    '/auth/refresh',
  )

  return response.data
}
