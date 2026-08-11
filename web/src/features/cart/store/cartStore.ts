import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import type {
  AddToCartPayload,
  CartItem,
} from '../types'

type CartState = {
  items: CartItem[]

  addItem: (product: AddToCartPayload) => void
  removeItem: (productId: number) => void
  setQuantity: (
    productId: number,
    quantity: number,
  ) => void
  clearCart: () => void
}

export const useCartStore = create<CartState>()(
  persist(
    (set) => ({
      items: [],

      addItem: (product) => {
        if (product.stock <= 0) {
          return
        }

        set((state) => {
          const existingItem = state.items.find(
            (item) => item.productId === product.productId,
          )

          if (!existingItem) {
            return {
              items: [
                ...state.items,
                {
                  ...product,
                  quantity: 1,
                },
              ],
            }
          }

          return {
            items: state.items.map((item) =>
              item.productId === product.productId
                ? {
                  ...item,
                  quantity: Math.min(
                    item.quantity + 1,
                    item.stock,
                  ),
                }
              : item,
            ),
          }
        })
      },

      removeItem: (productId) => {
        set((state) => ({
          items: state.items.filter(
            (item) => item.productId !== productId,
          ),
        }))
      },

      setQuantity: (productId, quantity) => {
        set((state) => ({
          items: state.items.map((item) =>
          item.productId === productId
            ? {
              ...item,
              quantity: Math.min(
                Math.max(quantity, 1),
                item.stock,
              ),
            }
          : item,
          ),
        }))
      },

      clearCart: () => {
        set({ items: [] })
      },
    }),
    {
      name: 'marketplace-cart',
    },
  ),
)