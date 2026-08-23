export interface Product {
  id: number
  title: string
  description?: string
  price: string
  stock: number
  status: string
  images: ProductImage[]
  store?: StoreInfo
  category?: CategoryInfo
  created_at: string
  updated_at: string
}

export interface ProductImage {
  id: number
  url: string
  position: number
}

export interface StoreInfo {
  id: number
  name: string
  description?: string
}

export interface CategoryInfo {
  id: number
  name: string
  slug: string
}

export interface CategoryNode {
  id: number
  name: string
  slug: string
  children?: CategoryNode[]
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface ProductListItem {
  id: number
  title: string
  price: string
  stock: number
  images: ProductImage[]
  store?: StoreInfo
  created_at: string
}

export interface CreateProductRequest {
  category_id?: number
  title: string
  description?: string
  price: string
  stock?: number
  images?: string[]
}

export interface UpdateProductRequest {
  title?: string
  description?: string
  price?: string
  stock?: number
}

export interface ModerateProductRequest {
  status: string
}

export interface ProductFilters {
  page?: number
  limit?: number
  category_id?: number
  store_id?: number
  min_price?: string
  max_price?: string
  search?: string
  sort?: 'price_asc' | 'price_desc' | 'created_desc'
}
