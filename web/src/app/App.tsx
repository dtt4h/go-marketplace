import { Route, Routes } from 'react-router-dom'

import { useAuthInitialization } from '../features/auth/hooks/useAuthInitialization'
import { ProtectedRoute } from './router/ProtectedRoute'
import { GuestRoute } from './router/GuestRoute'
import { useAuthStore } from '../features/auth/store/authStore'
import { AuthPage } from '../pages/AuthPage'
import { ProfilePage } from '../pages/ProfilePage'
import { CatalogPage } from '../pages/CatalogPage'
import { NotFoundPage } from '../pages/NotFoundPage'

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
      <Route path="/" element={<CatalogPage />} />
      <Route 
        path="/auth" 
        element={
          <GuestRoute>
            <AuthPage />
          </GuestRoute>
        } 
      />
      <Route 
        path="/profile" 
        element={
          <ProtectedRoute>
            <ProfilePage />
          </ProtectedRoute>
        } 
      />
      <Route
        path="/catalog"
        element = {<CatalogPage />}
      />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}

export default App
