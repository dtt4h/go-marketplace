import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

import { Button } from '../../../../shared/ui/Button/Button'
import { Input } from '../../../../shared/ui/Input/Input'
import type { SellerStore } from '../../types'

import cls from './StoreSettingsForm.module.scss'

const storeSettingsSchema = z.object({
  name: z.string().trim().min(1, 'Введите название магазина'),
  description: z
    .string()
    .trim()
    .min(10, 'Описание должно содержать не менее 10 символов'),
})

type StoreSettingsFormData = z.infer<typeof storeSettingsSchema>

type StoreSettingsFormProps = {
  store: SellerStore
  onSave: (store: SellerStore) => void
}

export function StoreSettingsForm({
  store,
  onSave,
}: StoreSettingsFormProps) {
  const [successMessage, setSuccessMessage] =
    useState<string | null>(null)
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty, isSubmitting },
  } = useForm<StoreSettingsFormData>({
    resolver: zodResolver(storeSettingsSchema),
    defaultValues: store,
  })

  useEffect(() => {
    reset(store)
  }, [reset, store])

  async function onSubmit(data: StoreSettingsFormData) {
    onSave(data)
    reset(data)
    setSuccessMessage('Настройки магазина сохранены')
  }

  return (
    <form
      className={cls.form}
      onSubmit={handleSubmit(onSubmit)}
      noValidate
    >
      <h2 className={cls.title}>Профиль магазина</h2>

      <Input
        id="store-name"
        type="text"
        label="Название магазина"
        error={errors.name?.message}
        {...register('name', {
          onChange: () => setSuccessMessage(null),
        })}
      />

      <div className={cls.field}>
        <label className={cls.label} htmlFor="store-description">
          Описание
        </label>
        <textarea
          id="store-description"
          className={cls.textarea}
          rows={3}
          aria-invalid={errors.description ? true : undefined}
          aria-describedby={
            errors.description ? 'store-description-error' : undefined
          }
          {...register('description', {
            onChange: () => setSuccessMessage(null),
          })}
        />
        {errors.description && (
          <p
            className={cls.error}
            id="store-description-error"
            role="alert"
          >
            {errors.description.message}
          </p>
        )}
      </div>

      {successMessage && (
        <p className={cls.success} role="status">
          {successMessage}
        </p>
      )}

      <Button type="submit" disabled={isSubmitting || !isDirty}>
        {isSubmitting ? 'Сохраняем...' : 'Сохранить изменения'}
      </Button>
    </form>
  )
}
