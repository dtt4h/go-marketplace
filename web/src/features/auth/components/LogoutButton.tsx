import { useState } from 'react'
import axios from 'axios'

import { logout as logoutUser } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import { Logout } from '../../../shared/assets/icons/Logout'
import cls from './LogoutButton.module.scss'

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

      setLogoutError('Не удалось выйти из аккаунта')
    } finally {
      setIsLoggingOut(false)
    }
  }

  return (
    <div>
      {logoutError && <p role="alert">{logoutError}</p>}
      <button
        className={cls.logoutButton}
        type="button"
        onClick={handleLogout}
        disabled={isLoggingOut}
        aria-label={
          isLoggingOut
            ? 'Выполняется выход'
            : 'Выйти из аккаунта'
        }
        title="Выйти из аккаунта"
      >
        {isLoggingOut ? 'Выход...' : <Logout
            aria-hidden="true"
            focusable="false"
        />}
      </button>
    </div>
  )
}
