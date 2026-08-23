import { NoImage } from '../../../../shared/assets/icons/NoImage'
import { useCartStore } from '../../../cart/store/cartStore'
import {
  calculateCartItemTotal,
  calculateCartTotals,
  formatMoney,
} from '../../../cart/utils/cartCalculations'

import cls from './CheckoutSummary.module.scss'

export function CheckoutSummary() {
  const items = useCartStore((state) => state.items)
  const totals = calculateCartTotals(items)

  return (
    <aside
      className={cls.summary}
      aria-labelledby="checkout-summary-heading"
    >
      <h2 className={cls.title} id="checkout-summary-heading">
        Ваш заказ
      </h2>

      <ul className={cls.items}>
        {items.map((item) => (
          <li className={cls.item} key={item.productId}>
            <div className={cls.imageContainer}>
              {item.imageUrl ? (
                <img
                  className={cls.image}
                  src={item.imageUrl}
                  alt=""
                />
              ) : (
                <NoImage
                  className={cls.placeholderIcon}
                  aria-hidden="true"
                />
              )}
            </div>

            <div className={cls.itemInfo}>
              <p className={cls.itemTitle}>{item.title}</p>
              <p className={cls.itemQuantity}>
                {item.quantity} шт.
              </p>
            </div>

            <p className={cls.itemPrice}>
              {formatMoney(calculateCartItemTotal(item))}
            </p>
          </li>
        ))}
      </ul>

      <dl className={cls.totals}>
        <div className={cls.totalRow}>
          <dt>Товары</dt>
          <dd>{formatMoney(totals.subtotal)}</dd>
        </div>
        <div className={cls.totalRow}>
          <dt>Доставка</dt>
          <dd>{formatMoney(totals.delivery)}</dd>
        </div>
      </dl>

      <div className={cls.grandTotal}>
        <p className={cls.grandTotalLabel}>Итого</p>
        <p className={cls.grandTotalValue}>
          {formatMoney(totals.total)}
        </p>
      </div>
    </aside>
  )
}
