import axios from 'axios'

import type { APIErrorResponse } from '../../../shared/api/types'

export function getApplicationErrorMessage(
  error: unknown,
  fallback: string,
): string {
  if (!axios.isAxiosError<APIErrorResponse>(error)) {
    return 'Произошла неизвестная ошибка'
  }

  if (!error.response) {
    return 'Не удалось подключиться к серверу'
  }

  return error.response.data?.error?.message ?? fallback
}
