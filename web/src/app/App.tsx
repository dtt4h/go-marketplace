import { Route, Routes } from 'react-router-dom'

import { useAuthInitialization } from '../features/auth/hooks/useAuthInitialization'
import { ProtectedRoute } from './router/ProtectedRoute'
import { GuestRoute } from './router/GuestRoute'
import { useAuthStore } from '../features/auth/store/authStore'
import { AuthPage } from '../pages/AuthPage'
import { ProfilePage } from '../pages/ProfilePage'
import { CatalogPage } from '../pages/CatalogPage'
import { NotFoundPage } from '../pages/NotFoundPage'
import { ProductPage } from '../pages/ProductPage'
import { MainLayout } from './layout/MainLayout'

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
      <Route element={<MainLayout />}>
        <Route index element={<CatalogPage />} />
        <Route
          path="catalog"
          element={<CatalogPage />}
        />
        <Route
          path="products/:productId"
          element={<ProductPage />}
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

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
    </Routes>
  )
}

export default App
