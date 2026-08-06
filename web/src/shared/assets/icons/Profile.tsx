import type { ComponentProps } from 'react'

type ProfileProps = ComponentProps<'svg'>

export function Profile ({
  width = 20,
  height = 20,
  ...svgProps
}: ProfileProps) {
  return (
    <svg 
    {...svgProps}
    width={width} 
    height={height}
    viewBox="0 0 20 20" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M9.99999 10C11.8409 10 13.3333 8.50763 13.3333 6.66668C13.3333 4.82573 11.8409 3.33334 9.99999 3.33334C8.15904 3.33334 6.66666 4.82573 6.66666 6.66668C6.66666 8.50763 8.15904 10 9.99999 10Z" stroke="#334155" stroke-width="1.33333"/>
      <path d="M3.33334 16.6667C3.33334 13.9167 6.33334 11.6667 10 11.6667C13.6667 11.6667 16.6667 13.9167 16.6667 16.6667" stroke="#334155" stroke-width="1.33333"/>
      </svg>
  )
}