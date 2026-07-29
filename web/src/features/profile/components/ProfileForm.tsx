import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { 
  updateProfileSchema, 
  type UpdateProfileFormData 
} from '../schemas/updateProfileSchema'
import type { UserProfile } from '../types'
import { updateCurrentUser } from '../api/users'

type ProfileFormProps = {
  profile: UserProfile
  onUpdated: (profile: UserProfile) => void
}

export function ProfileForm({
  profile,
  onUpdated
}: ProfileFormProps) {
  const {
    register,
    handleSubmit,
    formState: {
      errors,
      isSubmitting,
    },
  } = useForm<UpdateProfileFormData>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      username: profile.username,
      phone: profile.phone ?? '',
      avatar_url: profile.avatar_url ?? '',
    },
  })
  async function onSubmit(
    data: UpdateProfileFormData,
  ) {
    const updatedProfile = 
      await updateCurrentUser(data)

    onUpdated(updatedProfile)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div>
        <label htmlFor="profile-username">
          Имя пользователя
        </label>
        <input
          id="profile-username"
          type="text"
          {...register('username')}
        />
        {errors.username && (
          <p role="alert">
            {errors.username.message}
          </p>
        )}
      </div>
      <div>
        <label htmlFor="profile-phone">
          Телефон
        </label>
        <input
          id="profile-phone"
          type="tel"
          {...register('phone')}
        />
      </div>
      <div>
        <label htmlFor="profile-avatar-url">
          Ссылка на аватар
        </label>
        <input
          id="profile-avatar-url"
          type="url"
          {...register('avatar_url')}
        />
      </div>
      <button
        type="submit"
        disabled={isSubmitting}
      >
        {isSubmitting
          ? 'Сохраняем...'
          : 'Сохранить'}
      </button>
    </form>
  )
}