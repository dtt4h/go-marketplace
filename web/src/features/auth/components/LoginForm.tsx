import { useState } from 'react'
import axios from 'axios'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import cls from './LoginForm.module.scss'

import { login as loginUser } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import type { APIErrorResponse } from '../../../shared/api/types'

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

  const setSession = useAuthStore((state) => state.setSession)

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

      setSession(
        response.user,
        response.access_token,
      )

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
    <form className={cls.formClass} onSubmit={handleSubmit(onSubmit)} noValidate>
      <div className={cls.emailInput}>
        <label className={cls.authLabel} htmlFor="login-email">Email</label>
        <input
          className={cls.inputClass}
          placeholder='you@example.ru'
          id="login-email"
          type="email"
          autoComplete="email"
          {...register('email')}
        />
        {errors.email && <p className={cls.errorMes}>{errors.email.message}</p>}
      </div>

      <div className={cls.passwordInput}>
        <label className={cls.authLabel} htmlFor="login-password">Пароль</label>
        <input
          className={cls.inputClass}
          placeholder='********'
          id="login-password"
          type="password"
          autoComplete="current-password"
          {...register('password')}
        />
        {errors.password && <p className={cls.errorMes}>{errors.password.message}</p>}
      </div>

      {serverError && <p className={cls.serverAlert} role="alert">{serverError}</p>}
      {successMessage && <p role="status">{successMessage}</p>}

      <button className={cls.enterButton} type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Вход...' : 'Войти'}
      </button>
    </form>
  )
}
