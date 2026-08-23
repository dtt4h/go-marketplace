import { useMemo, useState } from 'react'
import { Link, Navigate } from 'react-router-dom'

import { useAuthStore } from '../features/auth/store/authStore'
import { SellerApplicationForm } from '../features/sellerApplication/components/SellerApplicationForm/SellerApplicationForm'
import { SellerRegistrationForm } from '../features/sellerApplication/components/SellerApplicationForm/SellerRegistrationForm'
import { useSellerApplications } from '../features/sellerApplication/hooks/useSellerApplications'
import type { SellerApplication } from '../features/sellerApplication/types'
import { Button } from '../shared/ui/Button/Button'

import cls from './SellerApplicationPage.module.scss'

export function SellerApplicationPage() {
  const user = useAuthStore((state) => state.user)
  const [submittedApplication, setSubmittedApplication] =
    useState<SellerApplication | null>(null)
  const [partialFailure, setPartialFailure] =
    useState<string | null>(null)
  const [isReapplying, setIsReapplying] = useState(false)
  const {
    applications,
    isLoading,
    error,
    retry,
  } = useSellerApplications(user?.role === 'buyer')

  const latestApplication = useMemo(() => {
    if (submittedApplication) {
      return submittedApplication
    }

    return [...applications].sort(
      (firstApplication, secondApplication) =>
        new Date(secondApplication.created_at).getTime() -
        new Date(firstApplication.created_at).getTime(),
    )[0] ?? null
  }, [applications, submittedApplication])

  if (user?.role === 'seller') {
    return <Navigate to="/profile" replace />
  }

  if (user?.role === 'admin') {
    return <Navigate to="/" replace />
  }

  function handleCreated(application: SellerApplication) {
    setSubmittedApplication(application)
    setPartialFailure(null)
    setIsReapplying(false)
  }

  return (
    <main className={cls.page}>
      <header className={cls.header}>
        <p className={cls.eyebrow}>Для продавцов</p>
        <h1 className={cls.title}>Стать продавцом</h1>
        <p className={cls.description}>
          Заполните данные магазина. После проверки заявки вы получите
          доступ к кабинету продавца и публикации товаров.
        </p>
      </header>

      <section className={cls.card}>
        {!user && (
          <>
            <SellerRegistrationForm
              onCreated={handleCreated}
              onPartialFailure={setPartialFailure}
            />
            <p className={cls.loginHint}>
              Уже зарегистрированы?{' '}
              <Link className={cls.loginLink} to="/auth">
                Войти
              </Link>
            </p>
          </>
        )}

        {user?.role === 'buyer' && isLoading && (
          <p className={cls.stateMessage} role="status">
            Проверяем статус заявки...
          </p>
        )}

        {user?.role === 'buyer' && error && !isLoading && (
          <div className={cls.stateContainer}>
            <p className={cls.errorMessage} role="alert">
              {error}
            </p>
            <Button type="button" onClick={retry}>
              Попробовать снова
            </Button>
          </div>
        )}

        {user?.role === 'buyer' && !isLoading && !error && (
          <>
            {partialFailure && (
              <p className={cls.errorMessage} role="alert">
                {partialFailure}
              </p>
            )}

            {!latestApplication && (
              <SellerApplicationForm onCreated={handleCreated} />
            )}

            {latestApplication?.status === 'pending' && (
              <div className={cls.stateContainer}>
                <span className={cls.pendingBadge}>На рассмотрении</span>
                <h2 className={cls.stateTitle}>Заявка отправлена</h2>
                <p className={cls.stateMessage}>
                  Мы сообщим результат проверки на email, указанный при
                  регистрации.
                </p>
              </div>
            )}

            {latestApplication?.status === 'approved' && (
              <div className={cls.stateContainer}>
                <span className={cls.approvedBadge}>Одобрена</span>
                <h2 className={cls.stateTitle}>Магазин одобрен</h2>
                <p className={cls.stateMessage}>
                  Обновите сессию, чтобы перейти в кабинет продавца.
                </p>
                <Button
                  type="button"
                  onClick={() => window.location.reload()}
                >
                  Перейти в кабинет
                </Button>
              </div>
            )}

            {latestApplication?.status === 'rejected' && !isReapplying && (
              <div className={cls.stateContainer}>
                <span className={cls.rejectedBadge}>Отклонена</span>
                <h2 className={cls.stateTitle}>Заявка не прошла проверку</h2>
                <p className={cls.stateMessage}>
                  Причина отклонения будет отправлена на email. Вы можете
                  исправить данные и подать новую заявку.
                </p>
                <Button
                  type="button"
                  onClick={() => setIsReapplying(true)}
                >
                  Подать новую заявку
                </Button>
              </div>
            )}

            {latestApplication?.status === 'rejected' && isReapplying && (
              <SellerApplicationForm onCreated={handleCreated} />
            )}
          </>
        )}
      </section>
    </main>
  )
}
