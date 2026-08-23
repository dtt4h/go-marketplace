import { useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { updateCurrentUser } from '../../../profile/api/users'
import { Button } from '../../../../shared/ui/Button/Button'
import { Input } from '../../../../shared/ui/Input/Input'
import { createSellerApplication } from '../../api/sellerApplications'
import {
  sellerApplicationSchema,
  type SellerApplicationFormData,
} from '../../schemas/sellerApplicationSchema'
import type { SellerApplication } from '../../types'
import { getApplicationErrorMessage } from '../../utils/getApplicationErrorMessage'

import cls from './SellerApplicationForm.module.scss'

type SellerApplicationFormProps = {
  onCreated: (application: SellerApplication) => void
}

export function SellerApplicationForm({
  onCreated,
}: SellerApplicationFormProps) {
  const [serverError, setServerError] = useState<string | null>(null)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SellerApplicationFormData>({
    resolver: zodResolver(sellerApplicationSchema),
    defaultValues: {
      phone: '',
      storeName: '',
      description: '',
      acceptedTerms: false,
    },
  })

  async function onSubmit(data: SellerApplicationFormData) {
    setServerError(null)

    try {
      await updateCurrentUser({ phone: data.phone })

      const application = await createSellerApplication({
        store_name: data.storeName,
        description: data.description || undefined,
      })

      onCreated(application)
    } catch (error) {
      setServerError(
        getApplicationErrorMessage(
          error,
          'Не удалось отправить заявку продавца',
        ),
      )
    }
  }

  return (
    <form
      className={cls.form}
      onSubmit={handleSubmit(onSubmit)}
      noValidate
    >
      <fieldset className={cls.section}>
        <legend className={cls.sectionTitle}>Контактные данные</legend>
        <div className={cls.fields}>
          <Input
            id="application-phone"
            type="tel"
            label="Телефон"
            autoComplete="tel"
            containerClassName={cls.fullWidthField}
            placeholder="+7 900 000-00-00"
            error={errors.phone?.message}
            {...register('phone')}
          />
        </div>
      </fieldset>

      <fieldset className={cls.section}>
        <legend className={cls.sectionTitle}>Данные магазина</legend>
        <div className={cls.fields}>
          <Input
            id="application-store-name"
            type="text"
            label="Название магазина"
            containerClassName={cls.fullWidthField}
            placeholder="Archive Store"
            error={errors.storeName?.message}
            {...register('storeName')}
          />
          <div className={cls.textareaField}>
            <label
              className={cls.label}
              htmlFor="application-description"
            >
              Описание магазина
            </label>
            <textarea
              id="application-description"
              className={cls.textarea}
              rows={4}
              placeholder="Расскажите о магазине и ассортименте"
              aria-invalid={errors.description ? true : undefined}
              aria-describedby={
                errors.description
                  ? 'application-description-error'
                  : undefined
              }
              {...register('description')}
            />
            {errors.description && (
              <p
                className={cls.error}
                id="application-description-error"
                role="alert"
              >
                {errors.description.message}
              </p>
            )}
          </div>
        </div>
      </fieldset>

      <div>
        <label className={cls.termsContainer}>
          <input
            className={cls.checkbox}
            type="checkbox"
            {...register('acceptedTerms')}
          />
          <span>
            Соглашаюсь с условиями оферты и политикой конфиденциальности
          </span>
        </label>
        {errors.acceptedTerms && (
          <p className={cls.error} role="alert">
            {errors.acceptedTerms.message}
          </p>
        )}
      </div>

      {serverError && (
        <p className={cls.serverAlert} role="alert">
          {serverError}
        </p>
      )}

      <Button
        className={cls.submitButton}
        type="submit"
        disabled={isSubmitting}
      >
        {isSubmitting ? 'Отправляем...' : 'Отправить заявку'}
      </Button>
    </form>
  )
}
