import { useCallback, useEffect, useState } from 'react'

import { getMockProductById } from '../mocks/getMockProductById'
import type { ProductDetails } from '../type'

export function useProduct(
  productId: string | undefined,
) {
  const [product, setProduct] =
    useState<ProductDetails | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [requestVersion, setRequestVersion] = useState(0)

  const retry = useCallback(() => {
    setRequestVersion((version) => version + 1)
  }, [])

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
        const response = await getMockProductById(productId)

        if (isActive) {
          if (!response) {
            setError('Товар не найден')
            return
          }

          setProduct(response)
        }
      } catch {
        if (!isActive) {
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
  }, [productId, requestVersion])

  return {
    product,
    isLoading,
    error,
    retry,
  }
}