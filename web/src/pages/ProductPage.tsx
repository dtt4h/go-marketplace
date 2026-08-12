import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'

import { ProductDetails } from '../features/products/component/ProductDetails/ProductDetails'
import { mockProductDetails } from '../features/products/mocks/productDetails'
import { useCartStore } from '../features/cart/store/cartStore'

import cls from './ProductPage.module.scss'

export function ProductPage() {
  const { productId } = useParams()
  const addItem = useCartStore((state) => state.addItem)
  const [isCartNoticeVisible, setIsCartNoticeVisible] =
    useState(false)
  const noticeTimeoutRef = useRef<number | null>(null)

  useEffect(() => {
    return () => {
      if (noticeTimeoutRef.current !== null) {
        window.clearTimeout(noticeTimeoutRef.current)
      }
    }
  }, [])

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

    setIsCartNoticeVisible(true)

    if (noticeTimeoutRef.current !== null) {
      window.clearTimeout(noticeTimeoutRef.current)
    }

    noticeTimeoutRef.current = window.setTimeout(() => {
      setIsCartNoticeVisible(false)
      noticeTimeoutRef.current = null
    }, 3000)
  }

  return (
    <main>
      <ProductDetails
        key={product.id}
        product={product}
        onAddToCart={handleAddToCart}
      />

      {isCartNoticeVisible && (
        <div
          className={cls.cartNotice}
          role="status"
          aria-live="polite"
        >
          Товар добавлен в корзину!
        </div>
      )}
    </main>
  )
}
