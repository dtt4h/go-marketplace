import { mockSellerProducts } from '../../mocks/seller'
import type { SellerProductStatus } from '../../types'

import cls from './SellerProducts.module.scss'

const statusLabels: Record<SellerProductStatus, string> = {
  active: 'Активен',
  moderation: 'На модерации',
  rejected: 'Отклонён',
}

const moneyFormatter = new Intl.NumberFormat('ru-RU')

export function SellerProducts() {
  if (mockSellerProducts.length === 0) {
    return (
      <p className={cls.emptyState} role="status">
        У магазина пока нет товаров.
      </p>
    )
  }

  return (
    <section className={cls.table} aria-label="Товары магазина">
      <div className={cls.tableHeader} aria-hidden="true">
        <span>Товар</span>
        <span>Статус</span>
        <span>Цена</span>
        <span>Остаток</span>
        <span />
      </div>

      {mockSellerProducts.map((product) => (
        <article className={cls.row} key={product.id}>
          <div className={cls.productCell}>
            <span className={cls.productPreview} aria-hidden="true" />
            <span className={cls.productName}>{product.name}</span>
          </div>
          <span
            className={`${cls.status} ${cls[product.status]}`}
          >
            {statusLabels[product.status]}
          </span>
          <span className={cls.price}>
            {moneyFormatter.format(product.price)} ₽
          </span>
          <span className={cls.stock}>{product.stock} шт.</span>
          <button
            className={cls.editButton}
            type="button"
            aria-label={`Редактировать ${product.name}`}
            title="Редактировать"
          >
            ✎
          </button>
        </article>
      ))}
    </section>
  )
}
