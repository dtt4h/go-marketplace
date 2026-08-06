import { Outlet } from 'react-router-dom'

import { Header } from './Header/Header'
import { Footer } from './Footer/Footer'

import cls from './MainLayout.module.scss'

export function MainLayout() {
  return (
    <div className={cls.layout}>
      <Header />

      <div className={cls.content}>
        <Outlet />
      </div>

      <Footer />
    </div>
  )
}