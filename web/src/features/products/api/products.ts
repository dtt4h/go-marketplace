import { apiClient } from "../../../shared/api/client"
import type { ProductDetails, ProductListParams, ProductListResponse } from "../type"

export async function getProducts(
  params: ProductListParams = {},
): Promise<ProductListResponse> {
  const response = await apiClient.get<ProductListResponse>(
    '/products',
    { params },
  )

  return response.data
}

export async function getProductById(
  productId: string,
): Promise<ProductDetails> {
  const response = await apiClient.get<ProductDetails>(
    `/products/${productId}`,
  )

  return response.data
}