import { apiClient } from '../../../shared/api/client'
import { useAuthStore } from '../store/authStore'

let interceptorId: number | null = null

export function configureAuthInterceptor(): void {
  if (interceptorId !== null) {
    return
  }

  interceptorId = apiClient.interceptors.request.use((config) => {
    const accessToken = useAuthStore.getState().accessToken

    if (
      accessToken &&
      !config.headers.has('Authorization')
    ) {
      config.headers.set(
        'Authorization',
        `Bearer ${accessToken}`,
      )
    }

    return config
  })
}
