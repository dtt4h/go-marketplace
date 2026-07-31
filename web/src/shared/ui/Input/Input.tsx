import type { ComponentProps } from 'react'

import cls from './Input.module.scss'

type InputProps = ComponentProps<'input'> & {
  id: string
  label: string
  error?: string
}

export function Input({
  id,
  label,
  error,
  className,
  ...inputProps
}: InputProps) {
  const errorId = `${id}-error`

  return (
    <div className={cls.inputContainer}>
      <label
        className={cls.inputLabel}
        htmlFor={id}
      >
        {label}
      </label>

      <input
        {...inputProps}
        className={`${cls.inputClassName} ${className ?? ''}`}
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
      />

      {error && (
        <p
          className={cls.error}
          id={errorId}
          role="alert"
        >
          {error}
        </p>
      )}
    </div>
  )
}
