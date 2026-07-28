import { LogoutButton } from '../features/auth/components/LogoutButton'
import { useAuthStore } from '../features/auth/store/authStore'

export function ProfilePage() {
  const user = useAuthStore((state) => state.user)
  return (
    <main>
      <h1>Профиль</h1>
      <p>Имя пользователя: {user?.username}</p>
      <p>Email: {user?.email}</p>
      <LogoutButton />
    </main>
  )
}
