import { mockSellerOrders } from '../../mocks/seller'
import type { SellerOrderStatus } from '../../types'

import cls from './SellerOrders.module.scss'

const statusLabels: Record<SellerOrderStatus, string> = {
  paid: 'Оплачен',
  shipped: 'Отправлен',
  delivered: 'Доставлен',
}

const moneyFormatter = new Intl.NumberFormat('ru-RU')

export function SellerOrders() {
  if (mockSellerOrders.length === 0) {
    return (
      <p className={cls.emptyState} role="status">
        У магазина пока нет заказов.
      </p>
    )
  }

  return (
    <section className={cls.table} aria-label="Заказы магазина">
      <div className={cls.tableHeader} aria-hidden="true">
        <span>Заказ</span>
        <span>Товар</span>
        <span>Покупатель</span>
        <span>Сумма</span>
        <span>Статус</span>
      </div>

      {mockSellerOrders.map((order) => (
        <article className={cls.row} key={order.id}>
          <span className={cls.orderNumber}>№ {order.number}</span>
          <span>{order.productName}</span>
          <span className={cls.secondary}>{order.customerName}</span>
          <span className={cls.total}>
            {moneyFormatter.format(order.total)} ₽
          </span>
          <span
            className={`${cls.status} ${cls[order.status]}`}
          >
            {statusLabels[order.status]}
          </span>
        </article>
      ))}
    </section>
  )
}
