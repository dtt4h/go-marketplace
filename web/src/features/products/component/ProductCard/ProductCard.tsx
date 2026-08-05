import type { ProductListItem } from '../../type'
import cls from './ProductCard.module.scss'
import { Link } from 'react-router-dom'
import { NoImage } from '../../../../shared/assets/icons/NoImage'

type ProductCardProps = {
  product: ProductListItem
}

export function ProductCard({
  product,
}: ProductCardProps) {
  const previewImage = product.images?.[0]

  return (
    <Link
      className={cls.cardLink}
      to={`/products/${product.id}`}
      aria-label={`Открыть товар ${product.title}`}
    >
      <article className={cls.mainContainer}>
        {previewImage ? (
          <div className={cls.imgCardContainer}>
            <img
              className={cls.imgCard}
              src={previewImage.url}
              alt={product.title}
            />
          </div>
        ) : (
          <div className={cls.imgCardContainer}>
            <NoImage className={cls.noImage}/>
          </div>
        )}
        <div className={cls.descContainer}>
          <p className={cls.categoryContainer}>{product.category?.name}</p>
          <h2 className={cls.titleContainer}>{product.title}</h2>
          <p className={cls.storeContainer}>
            {product.store?.name ?? 'Магазин не указан'}
          </p>
        </div>
        <p className={cls.priceContainer}>{product.price} ₽</p>
      </article>
    </Link>
  )
}