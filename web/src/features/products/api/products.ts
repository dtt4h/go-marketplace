import { apiClient } from "../../../shared/api/client"
import type { ProductListParams, ProductListResponse } from "../type"

export async function getProducts(
  params: ProductListParams = {},
): Promise<ProductListResponse> {
  const response = await apiClient.get<ProductListResponse>(
    '/products',
    { params },
  )

  return response.data
}