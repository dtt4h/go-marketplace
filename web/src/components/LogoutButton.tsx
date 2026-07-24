import { useState } from 'react'
import axios from 'axios'

import { logout as logoutUser } from '../api/auth'
import { useAuthStore } from '../store/authStore'

export function LogoutButton() {
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const [logoutError, setLogoutError] = useState<string | null>(null)

  const clearSession = useAuthStore(
    (state) => state.clearSession,
  )

  async function handleLogout() {
    setIsLoggingOut(true)
    setLogoutError(null)

    try {
      await logoutUser()
      clearSession()
    } catch (error) {
      if (
        axios.isAxiosError(error) &&
        error.response?.status === 401
      ) {
        // уже недействительная сессия
        clearSession()
        return
      }

      setLogoutError('Не удалось выполнить вход')
    } finally {
      setIsLoggingOut(false)
    }
  }

  return (
    <div>
      {logoutError && <p role="alert">{logoutError}</p>}

      <button
        type="button"
        onClick={handleLogout}
        disabled={isLoggingOut}
      >
        {isLoggingOut ? 'Выход...' : 'Выйти'}
      </button>
    </div>
  )
}