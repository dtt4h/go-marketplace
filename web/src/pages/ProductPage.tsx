import { useParams } from 'react-router-dom'

import { ProductDetails } from '../features/products/component/ProductDetails/ProductDetails'
import { mockProductDetails } from '../features/products/mocks/productDetails'
import { useCartStore } from '../features/cart/store/cartStore'

export function ProductPage() {
  const { productId } = useParams()
  const addItem = useCartStore((state) => state.addItem)

  const product = mockProductDetails.find(
    (item) => item.id === Number(productId),
  )

  if (!product) {
    return <p role="alert">Товар не найден</p>
  }

  const handleAddToCart = () => {
    const sortedImages = [...(product.images ?? [])].sort(
      (firstImage, secondImage) =>
        firstImage.position - secondImage.position,
    )

    addItem({
      productId: product.id,
      title: product.title,
      price: product.price,
      stock: product.stock,
      imageUrl: sortedImages[0]?.url,
      storeName: product.store?.name,
      categoryName: product.category?.name,
    })
  }

  return (
    <main>
      <ProductDetails
        key={product.id}
        product={product}
        onAddToCart={handleAddToCart}
      />
    </main>
  )
}
