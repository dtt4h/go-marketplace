import { useEffect, useState } from 'react'
import axios from 'axios'

import { getProductById } from '../api/products'
import type { ProductDetails } from '../type'

export function useProduct(
  productId: string | undefined,
) {
  const [product, setProduct] =
    useState<ProductDetails | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let isActive = true

    async function loadProduct() {
      setProduct(null)
      setError(null)
      setIsLoading(true)

      if (!productId) {
        setError('Некорректный идентификатор товара')
        setIsLoading(false)
        return
      }

      try {
        const response = await getProductById(productId)

        if (isActive) {
          setProduct(response)
        }
      } catch (error) {
        if (!isActive) {
          return
        }

        if (
          axios.isAxiosError(error) &&
          error.response?.status === 404
        ) {
          setError('Товар не найден')
          return
        }

        setError('Не удалось загрузить товар')
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void loadProduct()

    return () => {
      isActive = false
    }
  }, [productId])

  return {
    product,
    isLoading,
    error,
  }
}