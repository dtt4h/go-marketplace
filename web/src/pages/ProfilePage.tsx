import { LogoutButton } from '../features/auth/components/LogoutButton'
import { useProfile } from '../features/profile/hooks/useProfile'
import { ProfileForm } from '../features/profile/components/ProfileForm'

export function ProfilePage() {
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
  return (
    <main>
      <ProfileForm
        profile={profile}
        onUpdated={replaceProfile}
      />
      <LogoutButton />
    </main>
  )
}
