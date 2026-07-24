import { LoginForm } from '../components/LoginForm'
import { LogoutButton } from '../components/LogoutButton'
import { useAuthStore } from '../store/authStore'
import { RegisterForm } from '../components/RegisterForm'

export function AuthPage() {
  const user = useAuthStore((state) => state.user)

  if (user) {
    return (
      <main>
        <p>Вы вошли как {user.username}</p>
        <LogoutButton />
      </main>
    )
  }
  return (
    <main>
      <LoginForm />
      <RegisterForm />
    </main>
  )
}
