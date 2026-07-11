import * as React from 'react'

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

type TruncatedCellProps = {
  children: React.ReactNode
  cellClassName?: string
  className?: string
  contentClassName?: string
  side?: 'top' | 'bottom' | 'left' | 'right'
  tooltipClassName?: string
  tooltipContent?: React.ReactNode
}

export function TruncatedCell({
  children,
  cellClassName,
  className,
  contentClassName,
  side = 'top',
  tooltipClassName,
  tooltipContent,
}: TruncatedCellProps) {
  const content = tooltipContent ?? getTextContent(children)

  if (!content) {
    return (
      <div
        className={cn(
          'block max-w-full min-w-0 whitespace-normal break-words [overflow-wrap:anywhere]',
          cellClassName,
          className
        )}
      >
        {children}
      </div>
    )
  }

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <span
            tabIndex={0}
            className={cn(
              'block max-w-full min-w-0 truncate',
              cellClassName,
              className
            )}
          />
        }
      >
        <span className={cn('block truncate', contentClassName)}>
          {children}
        </span>
      </TooltipTrigger>
      <TooltipContent
        side={side}
        className={cn('max-w-xs break-all', tooltipClassName)}
      >
        {content}
      </TooltipContent>
    </Tooltip>
  )
}

function getTextContent(node: React.ReactNode): string {
  if (typeof node === 'string' || typeof node === 'number') return String(node)
  if (Array.isArray(node)) return node.map(getTextContent).join('')
  return ''
}
