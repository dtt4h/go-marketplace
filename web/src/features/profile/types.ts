import type { User } from '../auth/types'

export type UserProfile = User & {
  avatar_url: string | null
  phone: string | null
  created_at: string
}
