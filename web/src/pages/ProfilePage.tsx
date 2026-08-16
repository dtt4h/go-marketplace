import { useState } from 'react'

import { LogoutButton } from '../features/auth/components/LogoutButton'
import { SellerOrders } from '../features/seller/components/SellerOrders/SellerOrders'
import { SellerProducts } from '../features/seller/components/SellerProducts/SellerProducts'
import { StoreSettingsForm } from '../features/seller/components/StoreSettingsForm/StoreSettingsForm'
import {
  mockSellerStats,
  mockSellerStore,
} from '../features/seller/mocks/seller'
import type { SellerStore } from '../features/seller/types'
import { useProfile } from '../features/profile/hooks/useProfile'
import { Button } from '../shared/ui/Button/Button'

import cls from './ProfilePage.module.scss'

type SellerSection = 'products' | 'orders' | 'settings'

const moneyFormatter = new Intl.NumberFormat('ru-RU')

export function ProfilePage() {
  const [activeSection, setActiveSection] =
    useState<SellerSection>('products')
  const [store, setStore] =
    useState<SellerStore>(mockSellerStore)
  const { profile, isLoading, error } = useProfile()

  if (isLoading) {
    return <p role="status">Загрузка кабинета продавца</p>
  }

  if (error) {
    return <p role="alert">{error}</p>
  }

  if (!profile) {
    return <p role="alert">Профиль не найден</p>
  }

  const avatarLetter = store.name.charAt(0).toUpperCase()

  return (
    <main className={cls.wrapper}>
      <header className={cls.profileHeader}>
        <div className={cls.identityContainer}>
          <div className={cls.avatarContainer} aria-hidden="true">
            <span className={cls.avatarMok}>{avatarLetter}</span>
          </div>
          <div className={cls.nameContainer}>
            <h1 className={cls.profileName}>{store.name}</h1>
            <p className={cls.profileDesc}>
              Панель продавца · {profile.email}
            </p>
          </div>
        </div>

        <div className={cls.actionContainer}>
          <Button className={cls.addProductButton} type="button">
            <span className={cls.addIcon} aria-hidden="true">+</span>
            Добавить товар
          </Button>
          <LogoutButton />
        </div>
      </header>

      <section className={cls.stats} aria-label="Статистика магазина">
        <article className={cls.statCard}>
          <p>Активных товаров</p>
          <strong>{mockSellerStats.activeProducts}</strong>
        </article>
        <article className={cls.statCard}>
          <p>На модерации</p>
          <strong className={cls.moderationValue}>
            {mockSellerStats.productsOnModeration}
          </strong>
        </article>
        <article className={cls.statCard}>
          <p>Заказов в работе</p>
          <strong className={cls.ordersValue}>
            {mockSellerStats.activeOrders}
          </strong>
        </article>
        <article className={cls.statCard}>
          <p>Выручка за месяц</p>
          <strong>
            {moneyFormatter.format(mockSellerStats.monthlyRevenue)} ₽
          </strong>
        </article>
      </section>

      <nav
        className={cls.navContainer}
        aria-label="Разделы кабинета продавца"
      >
        <button
          className={cls.navButton}
          type="button"
          aria-pressed={activeSection === 'products'}
          onClick={() => setActiveSection('products')}
        >
          Товары
        </button>
        <button
          className={cls.navButton}
          type="button"
          aria-pressed={activeSection === 'orders'}
          onClick={() => setActiveSection('orders')}
        >
          Заказы магазина
        </button>
        <button
          className={cls.navButton}
          type="button"
          aria-pressed={activeSection === 'settings'}
          onClick={() => setActiveSection('settings')}
        >
          Настройки магазина
        </button>
      </nav>

      {activeSection === 'products' && <SellerProducts />}
      {activeSection === 'orders' && <SellerOrders />}
      {activeSection === 'settings' && (
        <StoreSettingsForm store={store} onSave={setStore} />
      )}
    </main>
  )
}
