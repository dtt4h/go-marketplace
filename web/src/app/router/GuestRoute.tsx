import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'

import { useAuthStore } from '../../features/auth/store/authStore'

type GuestRouteProps = {
  children: ReactNode
}

export function GuestRoute({
  children,
}: GuestRouteProps) {
  const user = useAuthStore((state) => state.user)

  if (user?.role === 'seller') {
    return <Navigate to="/profile" replace />
  }

  if (user?.role === 'buyer') {
    return <Navigate to="/seller/application" replace />
  }

  if (user?.role === 'admin') {
    return <Navigate to="/" replace />
  }

  return children
}
