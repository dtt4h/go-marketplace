import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import axios from 'axios'

import { updateProfileSchema, type UpdateProfileFormData } from '../schemas/updateProfileSchema'
import type { UserProfile } from '../types'
import { updateCurrentUser } from '../api/users'
import type { APIErrorResponse } from '../../../shared/api/types'

type ProfileFormProps = {
  profile: UserProfile
  onUpdated: (profile: UserProfile) => void
  onCancel: () => void
}

export function ProfileForm({
  profile,
  onUpdated,
  onCancel,
}: ProfileFormProps) {
  const [serverError, setServerError] =
    useState<string | null>(null)
  const [successMessage, setSuccessMessage] =
    useState<string | null>(null)
  const {
    register,
    handleSubmit,
    setError,
    formState: {
      errors,
      isSubmitting,
      isDirty,
    },
  } = useForm<UpdateProfileFormData>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      username: profile.username,
      phone: profile.phone ?? '',
    },
  })
  async function onSubmit(
    data: UpdateProfileFormData,
  ) {
    setServerError(null)
    setSuccessMessage(null)
    try {
      const updatedProfile = 
        await updateCurrentUser(data)

      onUpdated(updatedProfile)
      setSuccessMessage('Профиль успешно обновлен')
    } catch (error) {
      if (!axios.isAxiosError<APIErrorResponse>(error)) {
        setServerError('Произошла неизвестная ошибка')
        return
      }
      if (!error.response) {
        setServerError('Не удалось подключиться')
        return
      }

      const apiError = error.response.data?.error

      if (apiError?.code === 'CONFLICT') {
        setError('username', {
          type: 'server',
          message: 'Это имя пользователя уже занято',
        })
        return
      }

      setServerError(
        apiError?.message ?? 'Не удалось обновить профиль',
      )
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate>
      <div>
        <label htmlFor="profile-username">
          Имя пользователя
        </label>
        <input
          id="profile-username"
          type="text"
          {...register('username')}
        />
        {errors.username && (
          <p role="alert">
            {errors.username.message}
          </p>
        )}
      </div>
      <div>
        <label htmlFor="profile-phone">
          Телефон
        </label>
        <input
          id="profile-phone"
          type="tel"
          {...register('phone')}
        />
      </div>
      <div>
      </div>
      {serverError && (
        <p role="alert">{serverError}</p>
      )}
      {successMessage && (
        <p role="status">{successMessage}</p>
      )}
      <button
        type="submit"
        disabled={isSubmitting || !isDirty}
      >
        {isSubmitting
          ? 'Сохраняем...'
          : 'Сохранить'}
      </button>
      <button
        type="button"
        onClick={onCancel}
        disabled={isSubmitting}
        >
        Отменить
      </button>
    </form>
  )
}