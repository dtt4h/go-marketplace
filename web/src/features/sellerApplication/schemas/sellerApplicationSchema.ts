import { z } from 'zod'

const phoneSchema = z
  .string()
  .trim()
  .min(1, 'Введите номер телефона')
  .regex(
    /^\+?[0-9\s()-]{10,20}$/,
    'Введите корректный номер телефона',
  )

const storeFields = {
  phone: phoneSchema,
  storeName: z
    .string()
    .trim()
    .min(1, 'Введите название магазина')
    .max(200, 'Название не должно превышать 200 символов'),
  description: z
    .string()
    .trim()
    .max(2000, 'Описание не должно превышать 2000 символов'),
  acceptedTerms: z.boolean().refine((value) => value, {
    message: 'Необходимо принять условия',
  }),
}

export const sellerApplicationSchema = z.object({
  ...storeFields,
})

export const sellerRegistrationSchema = z.object({
  username: z
    .string()
    .trim()
    .min(1, 'Введите логин')
    .max(100, 'Логин не должен превышать 100 символов'),
  email: z
    .string()
    .trim()
    .email('Введите корректный email'),
  password: z
    .string()
    .min(8, 'Пароль должен содержать минимум 8 символов'),
  ...storeFields,
})

export type SellerApplicationFormData = z.infer<
  typeof sellerApplicationSchema
>

export type SellerRegistrationFormData = z.infer<
  typeof sellerRegistrationSchema
>
