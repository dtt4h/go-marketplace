import { useParams } from 'react-router-dom'

import { ProductDetails } from '../features/products/component/ProductDetails/ProductDetails'
import { mockProductDetails } from '../features/products/mocks/productDetails'

export function ProductPage() {
  const { productId } = useParams()

  const product = mockProductDetails.find(
    (item) => item.id === Number(productId),
  )

  if (!product) {
    return <p role="alert">Товар не найден</p>
  }

  return (
    <main>
      <ProductDetails key={product.id} product={product} />
    </main>
  )
}
