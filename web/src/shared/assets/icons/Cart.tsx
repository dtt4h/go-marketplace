import type { ComponentProps } from "react";

type CartProps = ComponentProps<'svg'>

export function Cart({
  width = 20,
  height = 20,
  ...svgProps
}: CartProps){
  return (
    <svg 
    {...svgProps} 
    width={width} 
    height={height} 
    viewBox="0 0 20 20" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg">
      <g clipPath="url(#clip0_178_481)">
      <path d="M7.50001 17.8333C8.14434 17.8333 8.66668 17.311 8.66668 16.6667C8.66668 16.0223 8.14434 15.5 7.50001 15.5C6.85568 15.5 6.33334 16.0223 6.33334 16.6667C6.33334 17.311 6.85568 17.8333 7.50001 17.8333Z" stroke="white" strokeWidth="1.33333"/>
      <path d="M15 17.8333C15.6443 17.8333 16.1667 17.311 16.1667 16.6667C16.1667 16.0223 15.6443 15.5 15 15.5C14.3557 15.5 13.8333 16.0223 13.8333 16.6667C13.8333 17.311 14.3557 17.8333 15 17.8333Z" stroke="white" strokeWidth="1.33333"/>
      <path d="M1.66666 2.5H4.16666L6.16666 12.8333C6.22555 13.1421 6.39163 13.4201 6.63559 13.6184C6.87955 13.8166 7.1857 13.9222 7.49999 13.9167H14.75C15.0528 13.9262 15.3498 13.8324 15.5922 13.6507C15.8345 13.4689 16.0077 13.21 16.0833 12.9167L19.1667 5.83333H4.99999" stroke="white" strokeWidth="1.33333"/>
      </g>
      <defs>
      <clipPath id="clip0_178_481">
      <rect width="20" height="20" fill="white"/>
      </clipPath>
      </defs>
    </svg>
  )
}
