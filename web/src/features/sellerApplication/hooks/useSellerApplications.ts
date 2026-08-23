import { useCallback, useEffect, useState } from 'react'

import { getMySellerApplications } from '../api/sellerApplications'
import type { SellerApplication } from '../types'

export function useSellerApplications(enabled: boolean) {
  const [applications, setApplications] = useState<
    SellerApplication[]
  >([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [requestVersion, setRequestVersion] = useState(0)

  const retry = useCallback(() => {
    setRequestVersion((version) => version + 1)
  }, [])

  useEffect(() => {
    let isActive = true

    if (!enabled) {
      return
    }

    async function loadApplications() {
      setIsLoading(true)
      setError(null)

      try {
        const response = await getMySellerApplications()

        if (isActive) {
          setApplications(response)
        }
      } catch {
        if (isActive) {
          setError('Не удалось получить статус заявки')
        }
      } finally {
        if (isActive) {
          setIsLoading(false)
        }
      }
    }

    void loadApplications()

    return () => {
      isActive = false
    }
  }, [enabled, requestVersion])

  return {
    applications,
    isLoading,
    error,
    retry,
  }
}
