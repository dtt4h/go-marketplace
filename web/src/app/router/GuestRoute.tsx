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

  if (user) {
    return <Navigate to="/profile" replace />
  }
  return children
}