import { Link } from 'react-router-dom'

import { CartList } from '../features/cart/components/CartList/CartList'
import { CartSummary } from '../features/cart/components/CartSummary/CartSummary'
import { useCartStore } from '../features/cart/store/cartStore'

import cls from './CartPage.module.scss'

export function CartPage() {
  const hasItems = useCartStore(
    (state) => state.items.length > 0,
  )

  return (
    <main className={cls.page}>
      <h1 className={cls.title}>Корзина</h1>
      <div className={cls.content}>
        <div className={cls.leftColumn}>
          <CartList />
          <Link className={cls.continueLink} to="/catalog">
            ← Продолжить покупки
          </Link>
        </div>
        {hasItems && <CartSummary />}
      </div>
    </main>
  )
}
