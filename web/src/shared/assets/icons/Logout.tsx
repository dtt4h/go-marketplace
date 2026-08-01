import type { ComponentProps } from 'react'

type LogoutProps = ComponentProps<'svg'>

export function Logout({
  width = 16, 
  height = 17,
  ...svgProps
}: LogoutProps){
  return (
    <svg
      {...svgProps}
      width={width}
      height={height}
      viewBox="0 0 16 17" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
    >
      <path 
        d="M2.5 13.3333H4.16667V15H14.1667V1.66667H4.16667V3.33333H2.5V0.833333C2.5 0.3731 2.8731 0 3.33333 0H15C15.4602 0 15.8333 0.3731 15.8333 0.833333V15.8333C15.8333 16.2936 15.4602 16.6667 15 16.6667H3.33333C2.8731 16.6667 2.5 16.2936 2.5 15.8333V13.3333ZM4.16667 7.5H10V9.16667H4.16667V11.6667L0 8.33333L4.16667 5V7.5Z" 
        fill="currentColor"/>
    </svg>
  )
}