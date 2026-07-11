import { Eye, EyeOff } from 'lucide-react'
import * as React from 'react'

import { Button } from '@/components/design-system/button'
import { Input } from '@/components/design-system/input'
import { cn } from '@/lib/utils'

type PasswordInputProps = Omit<
  React.InputHTMLAttributes<HTMLInputElement>,
  'type'
> & {
  ref?: React.Ref<HTMLInputElement>
}

export function PasswordInput({
  className,
  disabled,
  ref,
  ...props
}: PasswordInputProps) {
  const [showPassword, setShowPassword] = React.useState(false)

  return (
    <div className={cn('relative rounded-md', className)}>
      <Input
        type={showPassword ? 'text' : 'password'}
        ref={ref}
        disabled={disabled}
        {...props}
      />
      <Button
        type='button'
        size='icon-xs'
        variant='ghost'
        disabled={disabled}
        className='text-muted-foreground absolute end-1 top-1/2 -translate-y-1/2'
        onClick={() => setShowPassword((prev) => !prev)}
        aria-label='Toggle password visibility'
      >
        {showPassword ? (
          <Eye size={18} aria-hidden='true' />
        ) : (
          <EyeOff size={18} aria-hidden='true' />
        )}
      </Button>
    </div>
  )
}
