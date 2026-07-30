import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import axios from 'axios'

import cls from './ProfileForm.module.scss'

import { updateProfileSchema, type UpdateProfileFormData } from '../schemas/updateProfileSchema'
import type { UserProfile } from '../types'
import { updateCurrentUser } from '../api/users'
import type { APIErrorResponse } from '../../../shared/api/types'

import { Input } from '../../../shared/ui/Input/Input'
import { Button } from '../../../shared/ui/Button/Button'

type ProfileFormProps = {
  profile: UserProfile
  isEditing: boolean
  onUpdated: (profile: UserProfile) => void
  onCancel: () => void
}

export function ProfileForm({
  profile,
  isEditing,
  onUpdated,
  onCancel,
}: ProfileFormProps) {
  const [serverError, setServerError] =
    useState<string | null>(null)
  const {
    register,
    handleSubmit,
    setError,
    reset,
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
  function handleCancel(): void {
    reset({
      username: profile.username,
      phone: profile.phone ?? '',
    })

    setServerError(null)
    onCancel()
  }
  async function onSubmit(
    data: UpdateProfileFormData,
  ) {
    if (!isEditing) {
      return
    }
    setServerError(null)
    try {
      const updatedProfile = 
        await updateCurrentUser(data)
      
      reset({
        username: updatedProfile.username,
        phone: updatedProfile.phone ?? '',
      })

      onUpdated(updatedProfile)
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
    <form className={cls.formContainer} onSubmit={handleSubmit(onSubmit)} noValidate>
      <Input
        id="profile-username"
        type="text"
        label="Имя"
        readOnly = {!isEditing}
        error={errors.username?.message}
        {...register('username')}
      />
      <Input
        id="profile-phone"
        type="tel"
        label="Телефон"
        readOnly={!isEditing}
        error={errors.phone?.message}
        {...register('phone')}
      />
      {serverError && (
        <p role="alert">{serverError}</p>
      )}
      {isEditing && (
        <>
          <div className={cls.buttonContainer}>
            <Button
              type="submit"
              disabled={isSubmitting || !isDirty}
              variant="primary"
              className={cls.saveButton}
            >
            {isSubmitting ? 'Сохраняем...' : 'Сохранить'}
            </Button>
            <Button
              type="button"
              onClick={handleCancel}
              disabled={isSubmitting}
              variant="secondary"
              className={cls.cancelButton}
            >
              Отменить
            </Button>
          </div>
        </>
      )}
    </form>
  )
}