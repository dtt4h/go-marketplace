import { useParams } from 'react-router-dom'

import { mockProducts } from '../features/products/mocks/product'

export function ProductPage() {
  const { productId } = useParams()

  const product = mockProducts.find(
    (item) => item.id === Number(productId),
  )

  if (!product) {
    return <p role="alert">Товар не найден</p>
  }

  return (
    <main>
      <h1>{product.title}</h1>

      <p>{product.price} ₽</p>

      <p>
        Магазин: {product.store?.name ?? 'Не указан'}
      </p>

      <p>
        Категория: {product.category?.name ?? 'Не указана'}
      </p>
    </main>
  )
}
