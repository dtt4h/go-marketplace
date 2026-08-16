import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'

import { ProductDetails } from '../features/products/component/ProductDetails/ProductDetails'
import { useProduct } from '../features/products/hooks/useProduct'
import { useCartStore } from '../features/cart/store/cartStore'
import { Button } from '../shared/ui/Button/Button'

import cls from './ProductPage.module.scss'

export function ProductPage() {
  const { productId } = useParams()
  const { product, isLoading, error, retry } =
    useProduct(productId)
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

  if (isLoading) {
    return (
      <main className={cls.stateContainer}>
        <p className={cls.stateMessage} role="status">
          Загружаем товар...
        </p>
      </main>
    )
  }

  if (error || !product) {
    const canRetry = error !== 'Товар не найден'

    return (
      <main className={cls.stateContainer}>
        <p className={cls.errorMessage} role="alert">
          {error ?? 'Товар не найден'}
        </p>
        {canRetry && (
          <Button type="button" onClick={retry}>
            Попробовать снова
          </Button>
        )}
      </main>
    )
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
