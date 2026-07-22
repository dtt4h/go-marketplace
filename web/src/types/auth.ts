export type RegisterRequest = {
  email: string
  password: string
  username: string
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
  refresh_token: string
}

export type APIErrorResponse = {
  error: {
    code: string
    message: string
    details?: Record<string, unknown>
  }
}