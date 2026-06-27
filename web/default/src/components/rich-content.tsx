import { HtmlContent } from '@/components/html-content'
import { Markdown } from '@/components/ui/markdown'

type RichContentMode = 'markdown' | 'html'

interface RichContentProps {
  content: string
  mode?: RichContentMode
  breaks?: boolean
  className?: string
}

export function RichContent(props: RichContentProps) {
  if (props.mode === 'html') {
    return <HtmlContent content={props.content} className={props.className} />
  }

  return (
    <Markdown breaks={props.breaks} className={props.className}>
      {props.content}
    </Markdown>
  )
}
