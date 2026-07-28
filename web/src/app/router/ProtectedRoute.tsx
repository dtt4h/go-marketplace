import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'

import { useAuthStore } from '../../features/auth/store/authStore'

type ProtectedRouteProps = {
  children: ReactNode
}

export function ProtectedRoute({
  children,
}: ProtectedRouteProps) {
  const user = useAuthStore((state) => state.user)

  if (!user) {
    return <Navigate to="/auth" replace />
  }
  return children
}