export type SellerProductStatus =
  | 'active'
  | 'moderation'
  | 'rejected'

export type SellerOrderStatus =
  | 'paid'
  | 'shipped'
  | 'delivered'

export type SellerProduct = {
  id: number
  name: string
  status: SellerProductStatus
  price: number
  stock: number
}

export type SellerOrder = {
  id: number
  number: string
  productName: string
  customerName: string
  total: number
  status: SellerOrderStatus
}

export type SellerStore = {
  name: string
  description: string
}

export type SellerStats = {
  activeProducts: number
  productsOnModeration: number
  activeOrders: number
  monthlyRevenue: number
}
