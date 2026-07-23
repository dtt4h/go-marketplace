import { useState } from 'react'
import axios from 'axios'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { login as loginUser } from '../api/auth'
import type { APIErrorResponse } from '../types/auth'

const loginSchema = z.object({
  email: z
    .string()
    .trim()
    .email('Введите корректный email'),

  password: z
    .string()
    .min(1, 'Введите пароль'),
})

type LoginFormData = z.infer<typeof loginSchema>

export function LoginForm() {
  const [serverError, setServerError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  async function onSubmit(data: LoginFormData) {
    setServerError(null)
    setSuccessMessage(null)

    try {
      const response = await loginUser(data)

      setSuccessMessage(
        `Вы вошли как ${response.user.username}`,
      )

      //позже тут будет сохранение юзера и токены (как и в регистерформ)
    } catch (error) {
      if (!axios.isAxiosError<APIErrorResponse>(error)) {
        setServerError('Произошла неизвестная ошибка')
        return
      }

      if (!error.response) {
        setServerError('Не удалось подключиться к серверу')
        return
      }

      const apiError = error.response.data?.error

      if (apiError?.code === 'VALIDATION_ERROR'){
        setServerError(apiError.message)
        return
      }

      if (apiError?.code === 'UNAUTHORIZED') {
        setServerError('Неверный email или пароль')
        return
      }

      setServerError(
        apiError?.message ?? 'Не удалось выполнить вход',
      )
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate>
      <div>
        <label htmlFor="login-email">Email</label>
        <input
          id="login-email"
          type="email"
          autoComplete="email"
          {...register('email')}
        />
        {errors.email && <p>{errors.email.message}</p>}
      </div>

      <div>
        <label htmlFor="login-password">Пароль</label>
        <input
          id="login-password"
          type="password"
          autoComplete="current-password"
          {...register('password')}
        />
        {errors.password && <p>{errors.password.message}</p>}
      </div>

      {serverError && <p role="alert">{serverError}</p>}
      {successMessage && <p role="status">{successMessage}</p>}

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Вход...' : 'Войти'}
      </button>
    </form>
  )
}