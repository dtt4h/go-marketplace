export type ProductImage = {
  id: number
  url: string
  position: number
}

export type ProductStore = {
  id: number
  name: string
  description?: string
}

export type ProductCategory = {
  id: number
  name: string
  slug: string
  parent_id?: number
  children?: ProductCategory[]
}

export type ProductListItem = {
  id: number
  title: string
  price: string
  stock: number
  images?: ProductImage[]
  store?: ProductStore
  category?: ProductCategory
  created_at: string
}

export type ProductSort = 
  | 'price_asc'
  | 'price_desc'
  | 'created_desc'

export type ProductListParams = {
  page?: number
  limit?: number
  store_id?: number
  category_id?: number
  min_price?: string
  max_price?: string
  search?: string
  sort?: ProductSort
}

export type ProductListResponse = {
  items: ProductListItem[]
  total: number
  page: number
  limit: number
}

export type ProductDetails = {
  id: number
  title: string
  description?: string
  price: string
  stock: number
  status: string
  images?: ProductImage[]
  store?: ProductStore
  category?: ProductCategory
  created_at: string
  updated_at: string
}