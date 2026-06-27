import DOMPurify from 'dompurify'
import { useMemo } from 'react'

import { cn } from '@/lib/utils'

interface HtmlContentProps {
  content: string
  className?: string
}

export function HtmlContent(props: HtmlContentProps) {
  const html = useMemo(() => DOMPurify.sanitize(props.content), [props.content])

  return (
    <div
      className={cn(
        'prose prose-neutral dark:prose-invert max-w-none',
        props.className
      )}
      // eslint-disable-next-line react/no-danger -- html is sanitized above
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
