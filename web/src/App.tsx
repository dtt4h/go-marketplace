import { useAuthInitialization } from './hooks/useAuthInitialization'
import { AuthPage } from './pages/AuthPage'
import { useAuthStore } from './store/authStore'

function App() {
  useAuthInitialization()

  const isInitialized = useAuthStore(
    (state) => state.isInitialized,
  )

  if (!isInitialized) {
    return <p role="status">Проверяем сессию...</p>
  }

  return (
    <AuthPage />
  )
}

export default App
