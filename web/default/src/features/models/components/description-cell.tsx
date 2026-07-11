import { Button } from '@/components/design-system/button'

import { useModels } from './models-provider'

type DescriptionCellProps = {
  modelName: string
  description: string
}

export function DescriptionCell({
  modelName,
  description,
}: DescriptionCellProps) {
  const { setOpen, setDescriptionData } = useModels()

  if (!description) {
    return <span className='text-muted-foreground text-xs'>-</span>
  }

  const handleClick = () => {
    setDescriptionData({ modelName, description })
    setOpen('description')
  }

  return (
    <div className='max-w-full sm:max-w-[150px]'>
      <Button
        variant='link'
        onClick={handleClick}
        className='text-muted-foreground hover:text-foreground block h-auto w-full cursor-pointer p-0 text-left text-sm [overflow-wrap:anywhere] whitespace-normal no-underline sm:h-auto sm:overflow-hidden sm:text-ellipsis sm:whitespace-nowrap'
      >
        {description}
      </Button>
    </div>
  )
}
