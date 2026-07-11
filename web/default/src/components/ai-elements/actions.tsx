'use client'

import type { ComponentProps } from 'react'

import { Button } from '@/components/design-system/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

export type ActionsProps = ComponentProps<'div'>

export const Actions = ({ className, children, ...props }: ActionsProps) => (
  <div className={cn('flex items-center gap-1', className)} {...props}>
    {children}
  </div>
)

export type ActionProps = ComponentProps<typeof Button> & {
  tooltip?: string
  label?: string
}

export const Action = ({
  tooltip,
  children,
  label,
  className,
  variant = 'ghost',
  size = 'icon',
  ...props
}: ActionProps) => {
  const button = (
    <Button
      className={cn(
        'text-muted-foreground hover:text-foreground relative',
        className
      )}
      size={size}
      type='button'
      variant={variant}
      {...props}
    >
      {children}
      <span className='sr-only'>{label || tooltip}</span>
    </Button>
  )

  if (tooltip) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger render={button}></TooltipTrigger>
          <TooltipContent>
            <p>{tooltip}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  return button
}
