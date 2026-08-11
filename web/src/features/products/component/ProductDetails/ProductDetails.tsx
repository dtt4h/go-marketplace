import { useState } from 'react'

import { Button } from '../../../../shared/ui/Button/Button'
import { NoImage } from '../../../../shared/assets/icons/NoImage'
import type { ProductDetails as ProductDetailsType } from '../../type'
import { Cart } from '../../../../shared/assets/icons/Cart'

import cls from './ProductDetails.module.scss'

type ProductDetailsProps = {
  product: ProductDetailsType
  onAddToCart: () => void
}

export function ProductDetails({
  product,
  onAddToCart,
}: ProductDetailsProps) {
  const images = [...(product.images ?? [])].sort(
    (firstImage, secondImage) =>
      firstImage.position - secondImage.position,
  )
  const [selectedImageId, setSelectedImageId] =
    useState<number | null>(images[0]?.id ?? null)

  const selectedImage =
    images.find((image) => image.id === selectedImageId)
    ?? images[0]

  return (
    <article className={cls.productMainContainer}>
      <div className={cls.leftSide}>
        <div className={cls.imgsContainer}>
          {selectedImage ? (
            <>
              <div className={cls.mainImageContainer}>
                <img
                  className={cls.mainImage}
                  src={selectedImage.url}
                  alt={product.title}
                />
              </div>

              {images.length > 1 && (
                <div className={cls.thumbnails}>
                  {images.map((image, index) => (
                    <button
                      className={cls.thumbnailButton}
                      type="button"
                      key={image.id}
                      onClick={() => setSelectedImageId(image.id)}
                      aria-label={`Показать изображение ${index + 1}`}
                      aria-pressed={image.id === selectedImage.id}
                    >
                      <img
                        className={cls.thumbnailImage}
                        src={image.url}
                        alt=""
                      />
                    </button>
                  ))}
                </div>
              )}
            </>
          ) : (
            <div className={cls.noImageContainer}>
              <NoImage aria-hidden="true" />
              <p>Изображение отсутствует</p>
            </div>
          )}
        </div>
        <section aria-labelledby="product-description-heading" className={cls.descContainer}>
          <h2 id="product-description-heading" className={cls.descTitle}>
            Описание
          </h2>
          <p className={cls.productDescription}>
            {product.description ?? 'Описание отсутствует'}
          </p>
        </section>
        <section
          aria-labelledby="product-specifications-heading"
          className={cls.specContainer}
        >
          <h2
            id="product-specifications-heading"
            className={cls.descTitle}
          >
            Характеристики
          </h2>
          <div className={cls.specRow}>
            <p className={cls.catTitle}>Категория</p>
            <p className={cls.prodTitle}>
              {product.category?.name ?? 'Не указана'}
            </p>
          </div>
        </section>
      </div>

      <div className={cls.rightSide}>
        <div className={cls.rightSideUpper}>
          <h1 className={cls.productTitle}>{product.title}</h1>
          <div className={cls.priceNcount}>
            <p className={cls.productPrice}>{product.price} ₽</p>
            <p className={cls.productCount}>
              {product.stock > 0
                ? `В наличии: ${product.stock}`
                : 'Нет в наличии'}
            </p>
          </div>
          <Button
            type="button"
            disabled={product.stock === 0}
            className={cls.buyButton}
          >
            Купить сейчас
          </Button>
          <Button
            type="button"
            disabled={product.stock === 0}
            className={cls.cartButton}
            onClick={onAddToCart}
          >
            <Cart
              className={cls.cartSvg}
              aria-hidden="true"
              focusable="false"
            />
            В корзину
          </Button>
        </div>
        {product.store && (
          <section
            aria-labelledby="product-store-heading"
            className={cls.rightSideLower}
          >
            <div className={cls.storeContainer}>
              <h2
                id="product-store-heading"
                className={cls.storeTitle}
              >
                {product.store.name}
              </h2>
            </div>
            <Button type="button" className={cls.storeButton}>
              Магазин
            </Button>
          </section>
        )}
      </div>
    </article>
  )
}
