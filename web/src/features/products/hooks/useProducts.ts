import { useEffect, useState } from "react"

import { getProducts } from "../api/products"
import type { ProductListResponse } from "../type"

export function useProducts() {
  const [data, setData] = useState<ProductListResponse | null>(null,)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let isActive = true

    async function loadProducts() {
      try {
        setIsLoading(true)
        setError(null)

        const products = await getProducts()

        if (isActive) {
          setData(products)
        }
      } catch {
        if (isActive) {
          setError('Не удалось загрузить товар')
        }
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void loadProducts()

    return () => {
      isActive = false
    }
  }, [])

  return {
    products: data?.items ?? [],
    total: data?.total ?? 0,
    page: data?.page ?? 1,
    limit: data?. limit ?? 20,
    isLoading,
    error,
  }
}