export type RegisterRequest = {
  email: string
  password: string
  username: string
}

export type LoginRequest = {
  email: string
  password: string
}

export type User = {
  id: number
  email: string
  role: 'buyer' | 'seller' | 'admin'
  username: string
}

export type AuthResponse = {
  user: User
  access_token: string
}

export type APIErrorResponse = {
  error: {
    code: string
    message: string
    details?: Record<string, unknown>
  }
}

export type RefreshResponse = {
  access_token: string
}

export type UserProfile = User & {
  avatar_url: string | null
  phone: string | null
  created_at: string
}
