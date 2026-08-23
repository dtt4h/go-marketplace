import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'

import type { User } from '../../features/auth/types'
import { useAuthStore } from '../../features/auth/store/authStore'

type ProtectedRouteProps = {
  children: ReactNode
  allowedRoles?: User['role'][]
  forbiddenRedirect?: string
}

export function ProtectedRoute({
  children,
  allowedRoles,
  forbiddenRedirect = '/',
}: ProtectedRouteProps) {
  const user = useAuthStore((state) => state.user)

  if (!user) {
    return <Navigate to="/auth" replace />
  }

  if (
    allowedRoles &&
    !allowedRoles.includes(user.role)
  ) {
    return <Navigate to={forbiddenRedirect} replace />
  }

  return children
}
