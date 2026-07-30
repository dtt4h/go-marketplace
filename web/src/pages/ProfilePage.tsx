import { LogoutButton } from '../features/auth/components/LogoutButton'
import { useProfile } from '../features/profile/hooks/useProfile'
import type { UserProfile } from '../features/profile/types'
import { ProfileForm } from '../features/profile/components/ProfileForm'
import { useAuthStore } from '../features/auth/store/authStore'
import { useState } from 'react'

//import { Button } from '../shared/ui/Button/Button'
import cls from './ProfilePage.module.scss'

export function ProfilePage() {
  const [isEditing, setIsEditing] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  
  const updateAuthStore = useAuthStore(
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
    return <p role="status">{error}</p>
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
    updateAuthStore(updatedProfile)
    setIsEditing(false)
    setSuccessMessage('Профиль успешно обновлен')
  }

  return (
    <main className={cls.wrapper}>
      <div className={cls.profileHeader}>
        <div className={cls.avatarContainer}>
          {profile.avatar_url ? (
            <img
              src={profile.avatar_url}
              alt={`Аватар ${profile.username}`}
            />
          ) : (
            <span aria-hidden="true" className={cls.avatarMok}>
              {avatarLetter}
            </span>
          )}
        </div>
        <div className={cls.nameContainer}>
          <p className={cls.profileName}>{profile.username}</p>
          <div className={cls.profileDesc}>
            <p>{profile.email}</p>
            <p> · </p>
            <p>{profile.role}</p>
          </div>
        </div>
        {!isEditing && (
        <button
          className={cls.editButton}
          type="button"
          onClick={() => {
            setSuccessMessage(null)
            setIsEditing(true)
          }}
        >
          Редактировать
        </button>
        )}
        <LogoutButton />
      </div>
      {successMessage && (
        <p role="status">{successMessage}</p>
      )}
      <ProfileForm
        isEditing={isEditing}
        profile={profile}
        onUpdated={handleProfileUpdated}
        onCancel={() => setIsEditing(false)}
      />
    </main>
  )
}
