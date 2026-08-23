export type SellerApplicationStatus =
  | 'pending'
  | 'approved'
  | 'rejected'

export type CreateSellerApplicationRequest = {
  store_name: string
  description?: string
}

export type SellerApplication = {
  id: number
  user_id: number
  store_name: string
  description?: string
  status: SellerApplicationStatus
  created_at: string
  updated_at: string
}
