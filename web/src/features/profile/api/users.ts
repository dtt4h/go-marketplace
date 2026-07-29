import { apiClient } from '../../../shared/api/client'
import type { UpdateProfileRequest, UserProfile } from '../types'

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

export async function updateCurrentUser(
  data: UpdateProfileRequest,
  ): Promise<UserProfile> {
    const response = await apiClient.patch<UserProfile>(
      '/users/me', 
      data,
    )

    return response.data
}
