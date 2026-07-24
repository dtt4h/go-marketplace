import { create } from 'zustand'

import type { User } from '../types/auth'

type AuthState = {
  user: User | null
  accessToken: string | null
  isInitialized: boolean
  
  setSession: (user: User, accessToken: string) => void
  clearSession: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  isInitialized: false,

  setSession: (user, accessToken) => {
    set({
      user,
      accessToken,
      isInitialized: true,
    })
  },

  clearSession: () => {
    set({
      user: null,
      accessToken: null,
      isInitialized: true,
    })
  },
}))