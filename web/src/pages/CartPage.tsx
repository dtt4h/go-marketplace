import { useCartStore } from '../features/cart/store/cartStore'

export function CartPage() {
  const items = useCartStore((state) => state.items)

  return (
    <main>
      <h1>Корзина</h1>

      {items.length === 0 ? (
        <p>Корзина пуста</p>
      ) : (
        <p>Товаров в корзине: {items.length}</p>
      )}
    </main>
  )
}
