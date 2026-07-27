import { useEffect } from 'react'

import { refreshSession } from '../api/auth'
import { getCurrentUser } from '../api/users'
import { useAuthStore } from '../store/authStore'

let initializationPromise: Promise<void> | null = null

async function initializeAuth(): Promise<void> {
  const {
    isInitialized,
    setSession,
    clearSession,
  } = useAuthStore.getState()

  if (isInitialized) {
    return
  }

  try {
    const { access_token: accessToken } =
      await refreshSession()

    const user = await getCurrentUser(accessToken)

    setSession(user, accessToken)
  } catch {
    clearSession()
  }
}

export function useAuthInitialization(): void {
  useEffect(() => {
    initializationPromise ??= initializeAuth().finally(() => {
      initializationPromise = null
    })
  }, [])
}
