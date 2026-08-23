import { CartItem } from '../CartItem/CartItem'
import { useCartStore } from '../../store/cartStore'

import cls from './CartList.module.scss'

export function CartList() {
  const items = useCartStore((state) => state.items)
  const setQuantity = useCartStore(
    (state) => state.setQuantity,
  )
  const removeItem = useCartStore(
    (state) => state.removeItem,
  )

  if (items.length === 0) {
    return (
      <section className={cls.emptyState} aria-label="Корзина">
        <p className={cls.emptyMessage}>Корзина пуста</p>
      </section>
    )
  }

  return (
    <section
      className={cls.list}
      aria-label="Товары в корзине"
    >
      {items.map((item) => (
        <CartItem
          key={item.productId}
          item={item}
          onDecrease={() =>
            setQuantity(
              item.productId,
              item.quantity - 1,
            )
          }
          onIncrease={() =>
            setQuantity(
              item.productId,
              item.quantity + 1,
            )
          }
          onRemove={() => removeItem(item.productId)}
        />
      ))}
    </section>
  )
}
