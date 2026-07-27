import { useAuthStore } from '../store/authStore';
import { LogoutButton } from '../components/LogoutButton';

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