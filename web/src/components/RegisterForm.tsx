import { useState } from 'react'
import axios from 'axios'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { register as registerUser} from '../api/auth'
import type { APIErrorResponse } from '../types/auth'

const registerSchema = z.object({
  username: z.string().trim().min(1, 'Введите имя пользователя'),
  email: z.string().trim().email('Введите корректный email'),
  password: z.string().min(8, 'Пароль должен содержать минимум 8 символов'),
})

type RegisterFormData = z.infer<typeof registerSchema>


export function RegisterForm() {
  const [serverError, setServerError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting},
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  })

  async function onSubmit(data: RegisterFormData) {
    setServerError(null)
    setSuccessMessage(null)

    try {
      const response = await registerUser(data)

      setSuccessMessage(
        `Аккаунт ${response.user.email} успешно создан`,
      )
      //save tokens here
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

      if (apiError?.code === 'CONFLICT') {
        setError('email', {
          type: 'server',
          message: 'Пользователь с таким email уже существует',
        })
        return
      }

      if (apiError?.code === 'VALIDATION_ERROR') {
        setServerError(apiError.message)
        return
      }

      setServerError(
        apiError?.message ?? 'Не удалось зарегистрироваться',
      )
    }
  }

  return (
      <form onSubmit={handleSubmit(onSubmit)} noValidate>
        <div>
          <label htmlFor="username">Имя пользователя</label>
          <input 
            id="username" 
            type="text"
            {...register('username')}
          />
          {errors.username && <p>{errors.username.message}</p>}
        </div>

        <div>
          <label htmlFor="email">Email</label>
          <input 
            id="email" 
            type="email"
            {...register('email')}
          />
          {errors.email && <p>{errors.email.message}</p>}
        </div>

        <div>
          <label htmlFor="password">Пароль</label>
          <input 
            id="password" 
            type="password"
            {...register('password')} 
          />
          {errors.password && <p>{errors.password.message}</p>}
        </div>

        {serverError && <p role="alert">{serverError}</p>}
        {successMessage && <p role="status">{successMessage}</p>}
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Регистрация...' : 'Зарегистрироваться'}
        </button>
      </form>
  )
}