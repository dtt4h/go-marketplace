import type {
  SellerOrder,
  SellerProduct,
  SellerStats,
  SellerStore,
} from '../types'

export const mockSellerStore: SellerStore = {
  name: 'Архивный гардероб',
  description:
    'Архивная дизайнерская одежда и редкие вещи из частных коллекций.',
}

export const mockSellerStats: SellerStats = {
  activeProducts: 12,
  productsOnModeration: 1,
  activeOrders: 2,
  monthlyRevenue: 215_000,
}

export const mockSellerProducts: SellerProduct[] = [
  {
    id: 1,
    name: 'Кожаная куртка Helmut Lang',
    status: 'active',
    price: 145_000,
    stock: 1,
  },
  {
    id: 2,
    name: 'Пальто Maison Margiela',
    status: 'moderation',
    price: 82_000,
    stock: 1,
  },
  {
    id: 3,
    name: 'Брюки Yohji Yamamoto',
    status: 'active',
    price: 49_000,
    stock: 2,
  },
  {
    id: 4,
    name: 'Свитер Raf Simons',
    status: 'rejected',
    price: 58_000,
    stock: 0,
  },
]

export const mockSellerOrders: SellerOrder[] = [
  {
    id: 1,
    number: '10432',
    productName: 'Кожаная куртка Helmut Lang',
    customerName: 'И. Петров',
    total: 145_000,
    status: 'paid',
  },
  {
    id: 2,
    number: '10401',
    productName: 'Пальто Maison Margiela',
    customerName: 'А. Смирнова',
    total: 82_000,
    status: 'shipped',
  },
  {
    id: 3,
    number: '10388',
    productName: 'Брюки Yohji Yamamoto',
    customerName: 'Д. Волков',
    total: 49_000,
    status: 'delivered',
  },
]
