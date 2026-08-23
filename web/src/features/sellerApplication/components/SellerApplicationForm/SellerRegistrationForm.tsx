import { useState } from 'react'
import axios from 'axios'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { register as registerUser } from '../../../auth/api/auth'
import { useAuthStore } from '../../../auth/store/authStore'
import { updateCurrentUser } from '../../../profile/api/users'
import { Button } from '../../../../shared/ui/Button/Button'
import { Input } from '../../../../shared/ui/Input/Input'
import type { APIErrorResponse } from '../../../../shared/api/types'
import { createSellerApplication } from '../../api/sellerApplications'
import {
  sellerRegistrationSchema,
  type SellerRegistrationFormData,
} from '../../schemas/sellerApplicationSchema'
import type { SellerApplication } from '../../types'
import { getApplicationErrorMessage } from '../../utils/getApplicationErrorMessage'

import cls from './SellerApplicationForm.module.scss'

type SellerRegistrationFormProps = {
  onCreated: (application: SellerApplication) => void
  onPartialFailure: (message: string) => void
}

export function SellerRegistrationForm({
  onCreated,
  onPartialFailure,
}: SellerRegistrationFormProps) {
  const [serverError, setServerError] = useState<string | null>(null)
  const setSession = useAuthStore((state) => state.setSession)
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<SellerRegistrationFormData>({
    resolver: zodResolver(sellerRegistrationSchema),
    defaultValues: {
      username: '',
      email: '',
      phone: '',
      password: '',
      storeName: '',
      description: '',
      acceptedTerms: false,
    },
  })

  async function onSubmit(data: SellerRegistrationFormData) {
    setServerError(null)

    let session: Awaited<ReturnType<typeof registerUser>> | null = null

    try {
      session = await registerUser({
        username: data.username,
        email: data.email,
        password: data.password,
      })

      await updateCurrentUser(
        { phone: data.phone },
        session.access_token,
      )

      const application = await createSellerApplication(
        {
          store_name: data.storeName,
          description: data.description || undefined,
        },
        session.access_token,
      )

      setSession(session.user, session.access_token)
      onCreated(application)
    } catch (error) {
      if (session) {
        setSession(session.user, session.access_token)
        onPartialFailure(
          'Аккаунт создан, но заявку не удалось отправить. Проверьте данные и повторите отправку.',
        )
        return
      }

      if (axios.isAxiosError<APIErrorResponse>(error)) {
        const apiError = error.response?.data?.error

        if (apiError?.code === 'CONFLICT') {
          const field = apiError.details?.field

          if (field === 'username') {
            setError('username', {
              type: 'server',
              message: 'Этот логин уже занят',
            })
            return
          }

          if (field === 'email') {
            setError('email', {
              type: 'server',
              message: 'Аккаунт с таким email уже существует',
            })
            return
          }
        }
      }

      setServerError(
        getApplicationErrorMessage(
          error,
          'Не удалось создать аккаунт продавца',
        ),
      )
    }
  }

  return (
    <form
      className={cls.form}
      onSubmit={handleSubmit(onSubmit)}
      noValidate
    >
      <fieldset className={cls.section}>
        <legend className={cls.sectionTitle}>Данные для входа</legend>
        <div className={cls.fields}>
          <Input
            id="seller-username"
            type="text"
            label="Логин"
            autoComplete="username"
            placeholder="archive_store"
            error={errors.username?.message}
            {...register('username')}
          />
          <Input
            id="seller-email"
            type="email"
            label="Email"
            autoComplete="email"
            placeholder="seller@example.com"
            error={errors.email?.message}
            {...register('email')}
          />
          <Input
            id="seller-phone"
            type="tel"
            label="Телефон"
            autoComplete="tel"
            placeholder="+7 900 000-00-00"
            error={errors.phone?.message}
            {...register('phone')}
          />
          <Input
            id="seller-password"
            type="password"
            label="Пароль"
            autoComplete="new-password"
            placeholder="Минимум 8 символов"
            error={errors.password?.message}
            {...register('password')}
          />
        </div>
      </fieldset>

      <fieldset className={cls.section}>
        <legend className={cls.sectionTitle}>Данные магазина</legend>
        <div className={cls.fields}>
          <Input
            id="seller-store-name"
            type="text"
            label="Название магазина"
            containerClassName={cls.fullWidthField}
            placeholder="Archive Store"
            error={errors.storeName?.message}
            {...register('storeName')}
          />
          <div className={cls.textareaField}>
            <label className={cls.label} htmlFor="seller-description">
              Описание магазина
            </label>
            <textarea
              id="seller-description"
              className={cls.textarea}
              rows={4}
              placeholder="Расскажите о магазине и ассортименте"
              aria-invalid={errors.description ? true : undefined}
              aria-describedby={
                errors.description ? 'seller-description-error' : undefined
              }
              {...register('description')}
            />
            {errors.description && (
              <p
                className={cls.error}
                id="seller-description-error"
                role="alert"
              >
                {errors.description.message}
              </p>
            )}
          </div>
        </div>
      </fieldset>

      <div>
        <label className={cls.termsContainer}>
          <input
            className={cls.checkbox}
            type="checkbox"
            {...register('acceptedTerms')}
          />
          <span>
            Соглашаюсь с условиями оферты и политикой конфиденциальности
          </span>
        </label>
        {errors.acceptedTerms && (
          <p className={cls.error} role="alert">
            {errors.acceptedTerms.message}
          </p>
        )}
      </div>

      {serverError && (
        <p className={cls.serverAlert} role="alert">
          {serverError}
        </p>
      )}

      <Button
        className={cls.submitButton}
        type="submit"
        disabled={isSubmitting}
      >
        {isSubmitting ? 'Отправляем...' : 'Отправить заявку'}
      </Button>
    </form>
  )
}
