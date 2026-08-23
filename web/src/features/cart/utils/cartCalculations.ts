import type { CartItem } from '../types'

const DELIVERY_COST_KOPECKS = 90_000

const priceFormatter = new Intl.NumberFormat('ru-RU', {
  minimumFractionDigits: 0,
  maximumFractionDigits: 2,
})

function priceToKopecks(price: string): number {
  const priceInRubles = Number(price)

  if (!Number.isFinite(priceInRubles)) {
    return 0
  }

  return Math.round(priceInRubles * 100)
}

export function calculateCartItemTotal(
  item: CartItem,
): number {
  return priceToKopecks(item.price) * item.quantity
}

export function calculateCartTotals(items: CartItem[]) {
  const itemsCount = items.reduce(
    (total, item) => total + item.quantity,
    0,
  )

  const subtotal = items.reduce(
    (total, item) =>
      total + calculateCartItemTotal(item),
    0,
  )

  const delivery = items.length > 0
    ? DELIVERY_COST_KOPECKS
    : 0

  return {
    itemsCount,
    subtotal,
    delivery,
    total: subtotal + delivery,
  }
}

export function formatMoney(kopecks: number): string {
  return `${priceFormatter.format(kopecks / 100)} ₽`
}
