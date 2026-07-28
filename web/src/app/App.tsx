import { Route, Routes } from 'react-router-dom'

import { useAuthInitialization } from '../features/auth/hooks/useAuthInitialization'
import { useAuthStore } from '../features/auth/store/authStore'
import { AuthPage } from '../pages/AuthPage'
import { ProfilePage } from '../pages/ProfilePage'

function App() {
  useAuthInitialization()

  const isInitialized = useAuthStore(
    (state) => state.isInitialized,
  )

  if (!isInitialized) {
    return <p role="status">Проверяем сессию...</p>
  }

  return (
    <Routes>
      <Route path="/auth" element={<AuthPage />} />
      <Route path="/profile" element={<ProfilePage />} />
    </Routes>
  )
}

export default App
