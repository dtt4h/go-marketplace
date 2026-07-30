import type { ComponentProps } from 'react'
import cls from './Input.module.scss'

type InputProps = ComponentProps<'input'> & {
  id: string
  label: string
  error?: string
  placeholder?: string
}

export function Input({
  id,
  label,
  error,
  placeholder,
  ...inputProps
}: InputProps) {
  return (
    <div className={cls.inputContainer}>
      <label className={cls.inputLabel} htmlFor={id}>{label}</label>
      <input
      {...inputProps}
      placeholder={placeholder}
      className={cls.inputClassName}
      id={id}
      />

      {error && (
        <p className={cls.error} role="alert">
          {error}
        </p>
      )}
    </div>
  )
}