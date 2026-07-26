import { apiClient } from './client'
import type { UserProfile } from '../types/auth'

export async function getCurrentUser(): Promise<UserProfile> {
  const response = await apiClient.get<UserProfile>(
    '/users/me',
  )

  return response.data
}