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
import { CartPage } from '../pages/CartPage'
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
        <Route
          path="cart"
          element={<CartPage />}
        />
        <Route
          path="auth"
          element={
            <GuestRoute>
              <AuthPage />
            </GuestRoute>
          }
        />
        <Route
          path="profile"
          element={
            <ProtectedRoute>
              <ProfilePage />
            </ProtectedRoute>
          }
        />
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}

export default App
