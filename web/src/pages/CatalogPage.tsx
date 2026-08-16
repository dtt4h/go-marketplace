import { useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'

import { ProductCard } from '../features/products/component/ProductCard/ProductCard'
import { useProducts } from '../features/products/hooks/useProducts'
import { Button } from '../shared/ui/Button/Button'

import cls from './CatalogPage.module.scss'

export function CatalogPage() {
  const [searchParams] = useSearchParams()
  const searchQuery = searchParams.get('search')?.trim() ?? ''
  const {
    products,
    isLoading,
    isLoadingMore,
    hasMore,
    error,
    loadMore,
    retry,
  } = useProducts(searchQuery)
  const loadMoreTriggerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const trigger = loadMoreTriggerRef.current

    if (!trigger || !hasMore || isLoading || error) {
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          void loadMore()
        }
      },
      { rootMargin: '300px 0px' },
    )

    observer.observe(trigger)

    return () => observer.disconnect()
  }, [error, hasMore, isLoading, loadMore])

  if (isLoading) {
    return (
      <main className={cls.wrapperCatalog}>
        <p className={cls.stateMessage} role="status">
          Загружаем товары...
        </p>
      </main>
    )
  }

  if (error && products.length === 0) {
    return (
      <main className={cls.wrapperCatalog}>
        <div className={cls.errorContainer}>
          <p className={cls.errorMessage} role="alert">
            {error}
          </p>
          <Button type="button" onClick={retry}>
            Попробовать снова
          </Button>
        </div>
      </main>
    )
  }
  
  return (
    <main className={cls.wrapperCatalog}>
      {/*<p>Найдено товаро: {total}</p>*/}

      {products.length === 0 ? (
        <p className={cls.stateMessage} role="status">
          {searchQuery
            ? `По запросу «${searchQuery}» ничего не найдено`
            : 'Товары пока отсутствуют'}
        </p>
      ) : (
        <>
          <div className={cls.cardsContainer}>
            {products.map((product) => (
              <ProductCard
                key={product.id}
                product={product}
              />
            ))}
          </div>

          {error && (
            <div className={cls.paginationState}>
              <p className={cls.errorMessage} role="alert">
                {error}
              </p>
              <Button type="button" onClick={retry}>
                Попробовать снова
              </Button>
            </div>
          )}

          {isLoadingMore && (
            <p className={cls.paginationState} role="status">
              Загружаем ещё товары...
            </p>
          )}

          <div
            ref={loadMoreTriggerRef}
            className={cls.loadMoreTrigger}
            aria-hidden="true"
          />
        </>
      )}
    </main>
  )
}
