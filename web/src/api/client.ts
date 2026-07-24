import axios from 'axios'

import { useAuthStore } from '../store/authStore'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-type': 'application/json',
  },
  timeout: 10_000,
  withCredentials: true,
})

apiClient.interceptors.request.use((config) => {
  const accessToken = useAuthStore.getState().accessToken

  if (accessToken) {
    config.headers.set(
      'Authorization',
      `Bearer ${accessToken}`,
    )
  }

  return config
})