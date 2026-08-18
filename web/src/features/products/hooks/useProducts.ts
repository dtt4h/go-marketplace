import { useCallback, useEffect, useRef, useState } from 'react'

import { getMockProducts } from '../mocks/getMockProducts'
import type { ProductListItem } from '../type'

const PRODUCTS_PER_PAGE = 12

export function useProducts(search = '') {
  const [products, setProducts] = useState<ProductListItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [requestVersion, setRequestVersion] = useState(0)
  const requestIdRef = useRef(0)

  useEffect(() => {
    const requestId = ++requestIdRef.current

    async function loadFirstPage() {
      setProducts([])
      setTotal(0)
      setPage(1)
      setIsLoading(true)
      setIsLoadingMore(false)
      setError(null)

      try {
        const response = await getMockProducts({
          search,
          page: 1,
          limit: PRODUCTS_PER_PAGE,
        })

        if (requestId !== requestIdRef.current) {
          return
        }

        setProducts(response.items)
        setTotal(response.total)
        setPage(response.page)
      } catch {
        if (requestId === requestIdRef.current) {
          setError('Не удалось загрузить товары')
        }
      } finally {
        if (requestId === requestIdRef.current) {
          setIsLoading(false)
        }
      }
    }

    void loadFirstPage()

    return () => {
      if (requestId === requestIdRef.current) {
        requestIdRef.current += 1
      }
    }
  }, [search, requestVersion])

  const hasMore = products.length < total

  const loadMore = useCallback(async () => {
    if (isLoading || isLoadingMore || !hasMore || error) {
      return
    }

    const requestId = ++requestIdRef.current
    const nextPage = page + 1

    setIsLoadingMore(true)

    try {
      const response = await getMockProducts({
        search,
        page: nextPage,
        limit: PRODUCTS_PER_PAGE,
      })

      if (requestId !== requestIdRef.current) {
        return
      }

      setProducts((currentProducts) => [
        ...currentProducts,
        ...response.items.filter(
          (newProduct) =>
            !currentProducts.some(
              (product) => product.id === newProduct.id,
            ),
        ),
      ])
      setTotal(response.total)
      setPage(response.page)
    } catch {
      if (requestId === requestIdRef.current) {
        setError('Не удалось загрузить следующую страницу')
      }
    } finally {
      if (requestId === requestIdRef.current) {
        setIsLoadingMore(false)
      }
    }
  }, [error, hasMore, isLoading, isLoadingMore, page, search])

  const retry = useCallback(() => {
    setRequestVersion((version) => version + 1)
  }, [])

  return {
    products,
    total,
    page,
    limit: PRODUCTS_PER_PAGE,
    hasMore,
    isLoading,
    isLoadingMore,
    error,
    loadMore,
    retry,
  }
}
