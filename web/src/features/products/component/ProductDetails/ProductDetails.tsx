import { Button } from '../../../../shared/ui/Button/Button'
import { NoImage } from '../../../../shared/assets/icons/NoImage'
import type { ProductDetails as ProductDetailsType } from '../../type'

type ProductDetailsProps = {
  product: ProductDetailsType
}

export function ProductDetails({
  product,
}: ProductDetailsProps) {
  const hasImages = Boolean(product.images?.length)

  return (
    <article>
      <div>
        {hasImages ? (
          product.images?.map((image, index) => (
            <img
              key={image.id}
              src={image.url}
              alt={
                index === 0
                  ? product.title
                  : `${product.title}, изображение ${index + 1}`
              }
            />
          ))
        ) : (
          <div>
            <NoImage aria-hidden="true" />
            <p>Изображение отсутствует</p>
          </div>
        )}
      </div>

      <div>
        {product.category && (
          <p>{product.category.name}</p>
        )}

        <h1>{product.title}</h1>

        <p>{product.price} ₽</p>

        <p>
          {product.stock > 0
            ? `В наличии: ${product.stock}`
            : 'Нет в наличии'}
        </p>

        <Button
          type="button"
          disabled={product.stock === 0}
        >
          В корзину
        </Button>
      </div>

      <section aria-labelledby="product-description-heading">
        <h2 id="product-description-heading">
          Описание
        </h2>
        <p>
          {product.description ?? 'Описание отсутствует'}
        </p>
      </section>

      {product.store && (
        <section aria-labelledby="product-store-heading">
          <h2 id="product-store-heading">Продавец</h2>
          <p>{product.store.name}</p>
          {product.store.description && (
            <p>{product.store.description}</p>
          )}
        </section>
      )}
    </article>
  )
}
