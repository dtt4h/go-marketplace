import type {
  ProductListParams,
  ProductListResponse,
} from '../type'
import { mockProducts } from './product'

const MOCK_DELAY = 600
const DEFAULT_PAGE = 1
const DEFAULT_LIMIT = 12

export async function getMockProducts(
  params: ProductListParams = {},
): Promise<ProductListResponse> {
  await new Promise((resolve) => {
    window.setTimeout(resolve, MOCK_DELAY)
  })

  if (import.meta.env.VITE_MOCK_PRODUCTS_ERROR === 'true') {
    throw new Error('Mock products request failed')
  }

  const search = params.search?.trim().toLowerCase() ?? ''

  const filteredProducts = mockProducts.filter((product) => {
    if (!search) {
      return true
    }

    return [
      product.title,
      product.category?.name,
      product.store?.name,
    ].some((value) => value?.toLowerCase().includes(search))
  })

  const page = params.page ?? DEFAULT_PAGE
  const limit = params.limit ?? DEFAULT_LIMIT
  const startIndex = (page - 1) * limit
  const items = filteredProducts.slice(
    startIndex,
    startIndex + limit,
  )

  return {
    items,
    total: filteredProducts.length,
    page,
    limit,
  }
}
