import { LogoutButton } from '../features/auth/components/LogoutButton'
import { useProfile } from '../features/profile/hooks/useProfile'
import type { UserProfile } from '../features/profile/types'
import { ProfileForm } from '../features/profile/components/ProfileForm'
import { useAuthStore } from '../features/auth/store/authStore'
import { useState } from 'react'

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
    <main>
      <div>
        {profile.avatar_url ? (
          <img
            src={profile.avatar_url}
            alt={`Аватар ${profile.username}`}
          />
        ) : (
          <span aria-hidden="true">
            {avatarLetter}
          </span>
        )}
      </div>
      {successMessage && (
        <p role="status">{successMessage}</p>
      )}

      {isEditing ? (
        <ProfileForm
        profile={profile}
        onUpdated={handleProfileUpdated}
        onCancel={() => setIsEditing(false)}
      />
      ) : (
        <>
          <p>Имя пользователя: {profile.username}</p>
          <p>Email: {profile.email}</p>
          <p>Телефон: {profile.phone ?? 'Не указан'}</p>
          <p>Роль: {profile.role}</p>

          <button
            type="button"
            onClick={() => {
              setSuccessMessage(null)
              setIsEditing(true)
            }}
          >
            Редактировать
          </button>
        </>
      )}
      <LogoutButton />
    </main>
  )
}
