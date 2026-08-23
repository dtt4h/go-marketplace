import type { ComponentProps } from 'react'

import cls from './Button.module.scss'

type ButtonProps = ComponentProps<'button'> & {
  variant?: 'primary' | 'secondary'
}

export function Button({
  variant = 'primary',
  className,
  children,
  ...buttonProps
}: ButtonProps) {
  return (
    <button
      {...buttonProps}
      className={[
        cls.button,
        cls[variant],
        className,
      ].filter(Boolean).join(' ')}
    >
      {children}
    </button>
  )
}
