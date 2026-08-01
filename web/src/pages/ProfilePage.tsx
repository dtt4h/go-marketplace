import { useState } from 'react'

import { LogoutButton } from '../features/auth/components/LogoutButton'
import { useAuthStore } from '../features/auth/store/authStore'
import { ProfileForm } from '../features/profile/components/ProfileForm'
import { useProfile } from '../features/profile/hooks/useProfile'
import type { UserProfile } from '../features/profile/types'
import { Button } from '../shared/ui/Button/Button'
import { OrdersPlaceholder } from '../features/orders/components/OrdersPlaceholder'

import cls from './ProfilePage.module.scss'

type ProfileSection = 'profile' | 'orders'

export function ProfilePage() {
  const [isEditing, setIsEditing] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [activeSection, setActiveSection] = useState<ProfileSection>('profile')

  const updateAuthUser = useAuthStore(
    (state) => state.updateUser,
  )
  const {
    profile,
    isLoading,
    error,
    replaceProfile,
  } = useProfile()
  if (isLoading) {
    return <p role="status">Загрузка профиля</p>
  }

  if (error) {
    return <p role="alert">{error}</p>
  }

  if (!profile) {
    return <p role="alert">Профиль не найден</p>
  }

  const avatarLetter =
    profile.username.charAt(0).toUpperCase()

  function handleProfileUpdated(
    updatedProfile: UserProfile,
  ): void {
    replaceProfile(updatedProfile)
    updateAuthUser(updatedProfile)
    setIsEditing(false)
    setSuccessMessage('Профиль успешно обновлен')
  }

  return (
    <main className={cls.wrapper}>
      <div className={cls.profileHeader}>
        <div className={cls.identityContainer}>
          <div className={cls.avatarContainer}>
            {profile.avatar_url ? (
              <img
                src={profile.avatar_url}
                alt={`Аватар ${profile.username}`}
              />
            ) : (
              <span
                aria-hidden="true"
                className={cls.avatarMok}
              >
                {avatarLetter}
              </span>
            )}
          </div>
          <div className={cls.nameContainer}>
            <p className={cls.profileName}>
              {profile.username}
            </p>
            <div className={cls.profileDesc}>
              <p>{profile.email}</p>
              <p aria-hidden="true"> · </p>
              <p>{profile.role}</p>
            </div>
          </div>
        </div>
        <div className={cls.actionContainer}>
          <LogoutButton />
        </div>
      </div>
      <nav className={cls.navContainer} aria-label="Разделы личного кабинета">
          <button
            className={cls.navButton}
            type="button"
            aria-pressed={activeSection === 'orders'}
            disabled={isEditing}
            onClick={() => {
              setSuccessMessage(null)
              setActiveSection('orders')
            }}
          >
            Мои заказы
          </button>
          <button
            className={cls.navButton}
            type="button"
            aria-pressed={activeSection === 'profile'}
            onClick={() => setActiveSection('profile')}
          >
            Профиль
          </button>
        </nav>
      {successMessage && (
        <p role="status">{successMessage}</p>
      )}
      {activeSection === 'profile' ? (
        <ProfileForm
        isEditing={isEditing}
        profile={profile}
        onEdit={() => {
          setSuccessMessage(null)
          setIsEditing(true)
        }}
        onUpdated={handleProfileUpdated}
        onCancel={() => setIsEditing(false)}
      />
      ) : (
        <OrdersPlaceholder />
      )}
    </main>
  )
}
