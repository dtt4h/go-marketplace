import { useState } from 'react'
import { Link } from 'react-router-dom'

import { CheckoutForm } from '../features/checkout/components/CheckoutForm/CheckoutForm'
import { CheckoutSummary } from '../features/checkout/components/CheckoutSummary/CheckoutSummary'
import type { CheckoutFormData } from '../features/checkout/schemas/checkoutSchema'
import { useCartStore } from '../features/cart/store/cartStore'

import cls from './CheckoutPage.module.scss'

export function CheckoutPage() {
  const hasItems = useCartStore(
    (state) => state.items.length > 0,
  )
  const [preparedEmail, setPreparedEmail] =
    useState<string | null>(null)

  function handlePrepared(data: CheckoutFormData) {
    setPreparedEmail(data.email)
  }

  if (!hasItems) {
    return (
      <main className={cls.emptyPage}>
        <h1 className={cls.title}>Оформление заказа</h1>
        <p className={cls.emptyMessage}>
          В корзине нет товаров для оформления.
        </p>
        <Link className={cls.catalogLink} to="/catalog">
          Перейти в каталог
        </Link>
      </main>
    )
  }

  return (
    <main className={cls.page}>
      <h1 className={cls.title}>Оформление заказа</h1>

      <div className={cls.content}>
        <div className={cls.formColumn}>
          <CheckoutForm onPrepared={handlePrepared} />

          {preparedEmail && (
            <p className={cls.preparedMessage} role="status">
              Данные заполнены. После подключения оплаты чек и
              трек-номер будут отправляться на {preparedEmail}.
            </p>
          )}
        </div>

        <CheckoutSummary />
      </div>
    </main>
  )
}
