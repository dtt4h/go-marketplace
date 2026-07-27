import { apiClient } from './client'
import type { UserProfile } from '../types/auth'

export async function getCurrentUser(
  accessToken?: string,
): Promise<UserProfile> {
  const response = await apiClient.get<UserProfile>(
    '/users/me',
    accessToken
      ? {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        }
      : undefined,
  )

  return response.data
}
