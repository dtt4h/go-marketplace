import type { ComponentProps } from "react"

type NoImageProps = ComponentProps<'svg'>

export function NoImage ({
  width = 44,
  height = 44,
  ...svgProps
}: NoImageProps) {
  return (
    <svg
      {...svgProps}
      width={width}
      height={height}
      viewBox="0 0 44 44" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M34.8333 5.5H9.16667C7.14162 5.5 5.5 7.14162 5.5 9.16667V34.8333C5.5 36.8584 7.14162 38.5 9.16667 38.5H34.8333C36.8584 38.5 38.5 36.8584 38.5 34.8333V9.16667C38.5 7.14162 36.8584 5.5 34.8333 5.5Z" stroke="#9AA8B8" stroke-width="2.38333"/>
      <path d="M15.5834 18.5166C17.2034 18.5166 18.5167 17.2033 18.5167 15.5832C18.5167 13.9632 17.2034 12.6499 15.5834 12.6499C13.9633 12.6499 12.65 13.9632 12.65 15.5832C12.65 17.2033 13.9633 18.5166 15.5834 18.5166Z" stroke="#9AA8B8" stroke-width="2.38333"/>
      <path d="M38.5 27.5002L29.3333 18.3335L9.16663 38.5002" stroke="#9AA8B8" stroke-width="2.38333"/>
    </svg>
  )
}
