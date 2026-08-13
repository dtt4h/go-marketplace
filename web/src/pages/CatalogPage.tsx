import { useSearchParams } from 'react-router-dom'

//import { useProducts } from '../features/products/hooks/useProducts'
import { ProductCard } from '../features/products/component/ProductCard/ProductCard'
import { mockProducts } from '../features/products/mocks/product'

import cls from './CatalogPage.module.scss'

export function CatalogPage() {
  const [searchParams] = useSearchParams()
  const searchQuery = searchParams.get('search')?.trim() ?? ''
  const normalizedSearchQuery = searchQuery.toLowerCase()

  //const { products, total, isLoading, error, } = useProducts()

  //if (isLoading) {
  //  return <p role="status">Загружаем товары...</p>
  //}

  //if (error) {
  //  return <p role="alert">{error}</p>
  //}

  const products = mockProducts.filter((product) => {
    if (!normalizedSearchQuery) {
      return true
    }

    return [
      product.title,
      product.category?.name,
      product.store?.name,
    ].some((value) =>
      value?.toLowerCase().includes(normalizedSearchQuery),
    )
  })
  //const total = products.length
  
  return (
    <main className={cls.wrapperCatalog}>
      {/*<p>Найдено товаро: {total}</p>*/}

      {products.length === 0 ? (
        <p className={cls.emptyMessage} role="status">
          {searchQuery
            ? `По запросу «${searchQuery}» ничего не найдено`
            : 'Товары пока отсутствуют'}
        </p>
      ) : (
        <div className={cls.cardsContainer}>
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
            />
          ))}
        </div>
      )}
    </main>
  )
}
