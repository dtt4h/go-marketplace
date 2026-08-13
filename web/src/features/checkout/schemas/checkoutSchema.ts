import { z } from 'zod'

export const checkoutSchema = z.object({
  fullName: z
    .string()
    .trim()
    .min(2, 'Введите имя и фамилию')
    .max(100, 'Слишком длинное имя'),
  email: z
    .email('Введите корректный email'),
  phone: z
    .string()
    .trim()
    .min(10, 'Введите номер телефона')
    .max(30, 'Слишком длинный номер телефона'),
  city: z
    .string()
    .trim()
    .min(2, 'Введите город')
    .max(100, 'Слишком длинное название города'),
  address: z
    .string()
    .trim()
    .min(5, 'Введите адрес доставки')
    .max(200, 'Слишком длинный адрес'),
})

export type CheckoutFormData = z.infer<typeof checkoutSchema>
