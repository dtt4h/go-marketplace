import { useNavigate } from 'react-router-dom'

import { Button } from '../../../../shared/ui/Button/Button'
import { useCartStore } from '../../store/cartStore'
import {
  calculateCartTotals,
  formatMoney,
} from '../../utils/cartCalculations'

import cls from './CartSummary.module.scss'

export function CartSummary() {
  const navigate = useNavigate()
  const items = useCartStore((state) => state.items)
  const totals = calculateCartTotals(items)

  return (
    <aside
      className={cls.summary}
      aria-labelledby="cart-summary-heading"
    >
      <h2 className={cls.title} id="cart-summary-heading">
        Итого
      </h2>

      <dl className={cls.details}>
        <div className={cls.detailsRow}>
          <dt className={cls.label}>
            Товары ({totals.itemsCount})
          </dt>
          <dd className={cls.value}>
            {formatMoney(totals.subtotal)}
          </dd>
        </div>

        <div className={cls.detailsRow}>
          <dt className={cls.label}>Доставка</dt>
          <dd className={cls.value}>
            {formatMoney(totals.delivery)}
          </dd>
        </div>

        <div className={cls.detailsRow}>
          <dt className={cls.label}>Комиссия эскроу</dt>
          <dd className={cls.included}>включена</dd>
        </div>
      </dl>

      <div className={cls.totalRow}>
        <p className={cls.totalLabel}>К оплате</p>
        <p className={cls.totalPrice}>
          {formatMoney(totals.total)}
        </p>
      </div>

      <Button
        className={cls.checkoutButton}
        type="button"
        onClick={() => navigate('/checkout')}
      >
        Оформить заказ
      </Button>

      <span className={cls.notice}>
        Деньги замораживаются на эскроу и переводятся
        продавцу только после подтверждения получения.
      </span>
    </aside>
  )
}
