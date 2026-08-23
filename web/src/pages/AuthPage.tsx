import { LoginForm } from '../features/auth/components/LoginForm'

import cls from './AuthPage.module.scss'

export function AuthPage() {
  return (
    <main className={cls.authWrapper}>
      <div className={cls.logoContainer}>
        <p className={cls.logoText}>T</p>
      </div>
      <div className={cls.welcomeContainer}>
        <h1 className={cls.welcomeText}>Добро пожаловать в «Тарелку»</h1>
        <p className={cls.welcomeDesc}>Вход для продавцов</p>
      </div>
      <section className={cls.sectionContainer}>
        <LoginForm />
      </section>
    </main>
  )
}
