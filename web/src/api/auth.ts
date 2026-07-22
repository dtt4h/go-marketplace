import { apiClient } from './client'
import type { AuthResponse, RegisterRequest } from '../types/auth'

export async function register(
  data: RegisterRequest,): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>(
      '/auth/register',
      data,
    )
  return response.data
}
