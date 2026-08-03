import { useProducts } from '../features/products/hooks/useProducts'
import { ProductCard } from '../features/products/component/ProductCard/ProductCard'

export function CatalogPage() {
  const { products, total, isLoading, error, } = useProducts()

  if (isLoading) {
    return <p role="status">Загружаем товары...</p>
  }

  if (error) {
    return <p role="alert">{error}</p>
  }

  return (
    <main>
      <h1>Каталог</h1>

      <p>Найдено товаро: {total}</p>

      {products.length === 0 ? (
        <p>Товары пока отсутствуют</p>
      ) : (
        <div>
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