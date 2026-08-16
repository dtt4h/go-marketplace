import type { ProductDetails } from '../type'
import { mockProductDetails } from './productDetails'

const MOCK_DELAY = 600

export async function getMockProductById(
  productId: string,
): Promise<ProductDetails | null> {
  await new Promise((resolve) => {
    window.setTimeout(resolve, MOCK_DELAY)
  })

  if (import.meta.env.VITE_MOCK_PRODUCT_ERROR === 'true') {
    throw new Error('Mock product request failed')
  }

  const parsedProductId = Number(productId)

  if (!Number.isInteger(parsedProductId) || parsedProductId <= 0) {
    return null
  }

  return (
    mockProductDetails.find(
      (product) => product.id === parsedProductId,
    ) ?? null
  )
}
