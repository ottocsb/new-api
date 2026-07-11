'use client'

import {
  CheckmarkCircle02Icon,
  InformationCircleIcon,
  Alert02Icon,
  MultiplicationSignCircleIcon,
  Loading03Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { Toaster as Sonner, type ToasterProps } from 'sonner'

import { useTheme } from '@/context/theme-provider'

const Toaster = ({ ...props }: ToasterProps) => {
  const { resolvedTheme } = useTheme()

  return (
    <Sonner
      theme={resolvedTheme}
      className='toaster group'
      icons={{
        success: (
          <HugeiconsIcon
            icon={CheckmarkCircle02Icon}
            strokeWidth={2}
            className='size-4'
          />
        ),
        info: (
          <HugeiconsIcon
            icon={InformationCircleIcon}
            strokeWidth={2}
            className='size-4'
          />
        ),
        warning: (
          <HugeiconsIcon
            icon={Alert02Icon}
            strokeWidth={2}
            className='size-4'
          />
        ),
        error: (
          <HugeiconsIcon
            icon={MultiplicationSignCircleIcon}
            strokeWidth={2}
            className='size-4'
          />
        ),
        loading: (
          <HugeiconsIcon
            icon={Loading03Icon}
            strokeWidth={2}
            className='size-4 animate-spin'
          />
        ),
      }}
      style={
        {
          '--normal-bg': 'var(--popover)',
          '--normal-text': 'var(--popover-foreground)',
          '--normal-border': 'var(--border)',
          '--success-bg':
            'color-mix(in oklch, var(--success) 10%, var(--popover))',
          '--success-text':
            'color-mix(in oklch, var(--success) 55%, var(--foreground))',
          '--success-border':
            'color-mix(in oklch, var(--success) 32%, var(--border))',
          '--info-bg': 'color-mix(in oklch, var(--info) 10%, var(--popover))',
          '--info-text':
            'color-mix(in oklch, var(--info) 55%, var(--foreground))',
          '--info-border':
            'color-mix(in oklch, var(--info) 32%, var(--border))',
          '--warning-bg':
            'color-mix(in oklch, var(--warning) 10%, var(--popover))',
          '--warning-text':
            'color-mix(in oklch, var(--warning) 55%, var(--foreground))',
          '--warning-border':
            'color-mix(in oklch, var(--warning) 32%, var(--border))',
          '--error-bg':
            'color-mix(in oklch, var(--destructive) 10%, var(--popover))',
          '--error-text':
            'color-mix(in oklch, var(--destructive) 55%, var(--foreground))',
          '--error-border':
            'color-mix(in oklch, var(--destructive) 32%, var(--border))',
          '--border-radius': 'var(--radius)',
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast: 'cn-toast',
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
