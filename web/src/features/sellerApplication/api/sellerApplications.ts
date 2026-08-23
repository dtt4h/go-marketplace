import { apiClient } from '../../../shared/api/client'
import type {
  CreateSellerApplicationRequest,
  SellerApplication,
} from '../types'

export async function createSellerApplication(
  data: CreateSellerApplicationRequest,
  accessToken?: string,
): Promise<SellerApplication> {
  const response = await apiClient.post<SellerApplication>(
    '/seller-applications',
    data,
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

export async function getMySellerApplications(): Promise<
  SellerApplication[]
> {
  const response = await apiClient.get<SellerApplication[]>(
    '/seller-applications/me',
  )

  return response.data
}
