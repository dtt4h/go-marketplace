import { Link } from 'react-router-dom'

import { NoImage } from '../../../../shared/assets/icons/NoImage'
import type { CartItem as CartItemType } from '../../types'
import {
  calculateCartItemTotal,
  formatMoney,
} from '../../utils/cartCalculations'

import cls from './CartItem.module.scss'

type CartItemProps = {
  item: CartItemType
  onIncrease: () => void
  onDecrease: () => void
  onRemove: () => void
}

export function CartItem({
  item,
  onIncrease,
  onDecrease,
  onRemove,
}: CartItemProps) {
  return (
    <article className={cls.item}>
      <div className={cls.imageContainer}>
        {item.imageUrl ? (
          <img
            className={cls.image}
            src={item.imageUrl}
            alt={item.title}
          />
        ) : (
          <NoImage
            className={cls.placeholderIcon}
            aria-hidden="true"
          />
        )}
      </div>

      <div className={cls.info}>
        {item.categoryName && (
          <p className={cls.category}>{item.categoryName}</p>
        )}
        <Link
          className={cls.titleLink}
          to={`/products/${item.productId}`}
        >
          <h2 className={cls.title}>{item.title}</h2>
        </Link>

        {item.storeName && (
          <p className={cls.store}>{item.storeName}</p>
        )}
      </div>

      <div
        className={cls.quantityControl}
        aria-label={`Количество товара ${item.title}`}
      >
        <button
          className={cls.quantityButton}
          type="button"
          onClick={onDecrease}
          disabled={item.quantity <= 1}
          aria-label="Уменьшить количество"
        >
          −
        </button>

        <span className={cls.quantityValue} aria-live="polite">
          {item.quantity}
        </span>

        <button
          className={cls.quantityButton}
          type="button"
          onClick={onIncrease}
          disabled={item.quantity >= item.stock}
          aria-label="Увеличить количество"
        >
          +
        </button>
      </div>

      <div className={cls.priceBlock}>
        <p className={cls.price}>
          {formatMoney(calculateCartItemTotal(item))}
        </p>

        <button
          className={cls.removeButton}
          type="button"
          onClick={onRemove}
        >
          Удалить
        </button>
      </div>
    </article>
  )
}
