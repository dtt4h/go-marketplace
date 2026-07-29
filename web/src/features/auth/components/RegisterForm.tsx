import { useState } from 'react'
import axios from 'axios'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm, useWatch } from 'react-hook-form'
import { z } from 'zod'

import cls from './RegisterForm.module.scss'

import { register as registerUser} from '../api/auth'
import { useAuthStore } from '../store/authStore'
import type { APIErrorResponse } from '../../../shared/api/types'

const registerSchema = z.object({
  username: z.string().trim().min(1, 'Введите имя пользователя'),
  email: z.string().trim().email('Введите корректный email'),
  password: z.string().min(8, 'Пароль должен содержать минимум 8 символов'),

  acceptedTerms: z
    .boolean()
    .refine((value) => value, {
      message: 'Необходимо принять условие'
    }),
})

type RegisterFormData = z.infer<typeof registerSchema>


export function RegisterForm() {
  const [serverError, setServerError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const setSession = useAuthStore((state) => state.setSession)

  const {
    register,
    handleSubmit,
    setError,
    control,
    formState: { errors, isSubmitting},
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    
    defaultValues: {
      acceptedTerms: false,
    },
  })

  const acceptedTerms = useWatch({
    control,
    name: 'acceptedTerms',
  })

  async function onSubmit(data: RegisterFormData) {
    setServerError(null)
    setSuccessMessage(null)

    try {
      const response = await registerUser(data)

      setSession(
        response.user,
        response.access_token,
      )

      setSuccessMessage(
        `Аккаунт ${response.user.username} успешно создан`,
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
        const conflictField = apiError.details?.field

        if (conflictField === 'username'){ 
          setError('username', {
            type: 'server',
            message: 'Это имя пользователя уже занято',
          })
          return
        }

        if (conflictField === 'email') {
          setError('email', {
            type: 'server',
            message: 'Пользователь с таким email уже существует',
          })
        return
        }

        setServerError(
          apiError.message
            ?? 'Email или имя пользователя уже заняты',
        )
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
      <form className={cls.formClass} onSubmit={handleSubmit(onSubmit)} noValidate>
        <div className={cls.container}>
          <label className={cls.labelClass} htmlFor="username">Имя пользователя</label>
          <input
            placeholder='ivan_petrov'
            className={cls.inputClass} 
            id="username" 
            type="text"
            {...register('username')}
          />
          {errors.username && <p className={cls.errorMes}>{errors.username.message}</p>}
        </div>

        <div className={cls.container}>
          <label className={cls.labelClass} htmlFor="email">Email</label>
          <input 
            placeholder='you@example.com'
            className={cls.inputClass}
            id="email" 
            type="email"
            {...register('email')}
          />
          {errors.email && <p className={cls.errorMes}>{errors.email.message}</p>}
        </div>

        <div className={cls.container}>
          <label className={cls.labelClass} htmlFor="password">Пароль</label>
          <input
            placeholder='минимум 8 символов'
            className={cls.inputClass} 
            id="password" 
            type="password"
            {...register('password')} 
          />
          {errors.password && <p className={cls.errorMes}>{errors.password.message}</p>}
        </div>

        <div className={cls.agreeContainer}>
          <input 
          className={cls.agreeCheck} 
          id='acceptedTerms'
          type='checkbox'
          {...register('acceptedTerms')}
          />
          <a 
            className={cls.agreeText} 
            href="https://google.com" 
            target='_blank'
            rel="noopener noreferrer"
          >
            Соглашаюсь с условиями оферты и политикой <br></br> конфиденциальности
          </a>
        </div>

        {serverError && <p className={cls.serverAlert} role="alert">{serverError}</p>}
        {successMessage && <p role="status">{successMessage}</p>}
        <button className={cls.enterButton} type="submit" disabled={isSubmitting || !acceptedTerms}>
          {isSubmitting ? 'Регистрация...' : 'Зарегистрироваться'}
        </button>
      </form>
  )
}
