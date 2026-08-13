import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { Button } from '../../../../shared/ui/Button/Button'
import { Input } from '../../../../shared/ui/Input/Input'
import {
  checkoutSchema,
  type CheckoutFormData,
} from '../../schemas/checkoutSchema'

import cls from './CheckoutForm.module.scss'

type CheckoutFormProps = {
  onPrepared: (data: CheckoutFormData) => void
}

export function CheckoutForm({
  onPrepared,
}: CheckoutFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CheckoutFormData>({
    resolver: zodResolver(checkoutSchema),
    defaultValues: {
      fullName: '',
      email: '',
      phone: '',
      city: '',
      address: '',
    },
  })

  return (
    <form
      className={cls.form}
      onSubmit={handleSubmit(onPrepared)}
      noValidate
    >
      <h2 className={cls.title}>Адрес и получатель</h2>

      <div className={cls.fields}>
        <Input
          id="checkout-full-name"
          type="text"
          label="Имя и фамилия"
          placeholder="Иван Петров"
          autoComplete="name"
          error={errors.fullName?.message}
          {...register('fullName')}
        />

        <Input
          id="checkout-phone"
          type="tel"
          label="Телефон"
          placeholder="+7 900 000-00-00"
          autoComplete="tel"
          error={errors.phone?.message}
          {...register('phone')}
        />

        <Input
          id="checkout-email"
          type="email"
          label="Email"
          placeholder="you@example.com"
          autoComplete="email"
          containerClassName={cls.fullWidthField}
          error={errors.email?.message}
          {...register('email')}
        />

        <Input
          id="checkout-city"
          type="text"
          label="Город"
          placeholder="Москва"
          autoComplete="address-level2"
          containerClassName={cls.fullWidthField}
          error={errors.city?.message}
          {...register('city')}
        />

        <Input
          id="checkout-address"
          type="text"
          label="Адрес доставки"
          placeholder="Улица, дом, квартира"
          autoComplete="street-address"
          containerClassName={cls.fullWidthField}
          error={errors.address?.message}
          {...register('address')}
        />
      </div>

      <fieldset className={cls.delivery}>
        <legend className={cls.deliveryTitle}>
          Способ доставки
        </legend>

        <label className={cls.deliveryOption}>
          <input
            className={cls.deliveryRadio}
            type="radio"
            name="delivery"
            checked
            readOnly
          />
          <span className={cls.deliveryText}>
            <span className={cls.deliveryName}>
              Доставка со страховкой
            </span>
            <span className={cls.deliveryDescription}>
              Бережная упаковка · трек-номер после отправки
            </span>
          </span>
          <span className={cls.deliveryPrice}>900 ₽</span>
        </label>
      </fieldset>

      <Button
        className={cls.submitButton}
        type="submit"
        disabled={isSubmitting}
      >
        {isSubmitting
          ? 'Проверяем данные...'
          : 'Продолжить к оплате'}
      </Button>
    </form>
  )
}
