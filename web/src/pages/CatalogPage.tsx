//import { useProducts } from '../features/products/hooks/useProducts'
import { ProductCard } from '../features/products/component/ProductCard/ProductCard'
import { mockProducts } from '../features/products/mocks/product'
import cls from "./CatalogPage.module.scss"

export function CatalogPage() {
  //const { products, total, isLoading, error, } = useProducts()

  //if (isLoading) {
  //  return <p role="status">Загружаем товары...</p>
  //}

  //if (error) {
  //  return <p role="alert">{error}</p>
  //}

  const products = mockProducts
  //const total = products.length
  
  return (
    <main className={cls.wrapperCatalog}>
      {/*<p>Найдено товаро: {total}</p>*/}

      {products.length === 0 ? (
        <p>Товары пока отсутствуют</p>
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