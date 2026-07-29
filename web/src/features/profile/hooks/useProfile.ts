import { useEffect, useState } from 'react'
import { getCurrentUser } from '../api/users'
import type { UserProfile } from '../types'

export function useProfile() {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let isActive = true
    async function loadProfile() {
      setError(null)
      try {
        const response = await getCurrentUser()
        if (isActive) {
          setProfile(response)
        }
      } catch {
        if (isActive) {
          setError('Не удалось загрузить профиль')
        }
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }
    void loadProfile()
    return () => {
      isActive = false
    }
  }, [])

    function replaceProfile(
    updatedProfile: UserProfile,
  ): void {
    setProfile(updatedProfile)
  }

  return {
    profile,
    isLoading,
    error,
    replaceProfile,
  }
}