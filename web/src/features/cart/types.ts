export type CartItem = {
  productId: number
  title: string
  price: string
  stock: number
  quantity: number
  imageUrl?: string
  storeName?: string
  categoryName?: string
}

export type AddToCartPayload = Omit<CartItem, 'quantity'>