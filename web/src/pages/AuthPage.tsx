import { LoginForm } from '../components/LoginForm'
import { LogoutButton } from '../components/LogoutButton'
import { useAuthStore } from '../store/authStore'
import { RegisterForm } from '../components/RegisterForm'

import { useState } from 'react'

import cls from './AuthPage.module.scss'

type AuthMode = 'login' | 'register'

export function AuthPage() {
  const [mode, setMode] = useState<AuthMode>('login')

  const user = useAuthStore((state) => state.user)

  if (user) {
    return (
      <main>
        <p>Вы вошли как {user.username}</p>
        <LogoutButton />
      </main>
    )
  }

  return (
    <main className={cls.authWrapper} data-auth-mode={mode}>
      <div className={cls.logoContainer}>
        <p className={cls.logoText}>T</p>
      </div>
      <div className={cls.welcomeContainer}>
        <h1 className={cls.welcomeText}>Добро пожаловать в «Тарелку»</h1>
        <p className={cls.welcomeDesc}>Вход и регистрация покупателей</p>
      </div>
      <div className={cls.choiceContainer} aria-label="Выбор формы авторизации">
        <button
          className={cls.choiceButton}
          type="button"
          aria-pressed={mode === 'login'}
          onClick={() => setMode('login')}
        >
          <span className={cls.textButton}>Вход</span>
        </button>

        <button
          className={cls.choiceButton}
          type="button"
          aria-pressed={mode === 'register'}
          onClick={() => setMode('register')}
        >
          <span className={cls.textButton}>Регистрация</span>
        </button>
      </div>
      <section className={cls.sectionContainer}>
        {mode === 'login' ? (
          <LoginForm />
        ) : (
          <RegisterForm />
        )}
      </section>
    </main>
  )
}
