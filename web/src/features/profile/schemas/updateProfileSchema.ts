import { z } from 'zod'

export const updateProfileSchema = z.object({
  username: z
    .string()
    .trim()
    .min(1, 'Введите имя пользователя'),
  phone: z
    .string()
    .trim(),
})

export type UpdateProfileFormData =
  z.infer<typeof updateProfileSchema>