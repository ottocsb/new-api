import { Zap, ExternalLink, Gauge } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import { Button } from '@/components/design-system/button'
import { StatusBadge, type StatusVariant } from '@/components/status-badge'
import { openExternalSpeedTest } from '@/features/dashboard/lib/api-info'
import type { ApiInfoItem, PingStatus } from '@/features/dashboard/types'
import { getBgColorClass } from '@/lib/colors'
import { cn } from '@/lib/utils'

interface ApiInfoItemProps {
  item: ApiInfoItem
  status: PingStatus
  onTest: (url: string) => void
}

export function ApiInfoItemComponent(props: ApiInfoItemProps) {
  const { t } = useTranslation()
  const item = props.item
  const status = props.status
  let latencyVariant: StatusVariant = 'success'
  if (status.latency !== null && status.latency >= 500) {
    latencyVariant = 'destructive'
  } else if (status.latency !== null && status.latency >= 200) {
    latencyVariant = 'warning'
  }

  return (
    <div className='group hover:bg-muted/40 flex items-center justify-between gap-2 px-3 py-2.5 transition-colors sm:gap-3 sm:px-5 sm:py-3'>
      <div className='flex min-w-0 flex-1 items-center gap-2 sm:gap-3'>
        <span
          className={cn(
            'inline-block size-2 shrink-0 rounded-full',
            getBgColorClass(item.color)
          )}
        />

        <div className='flex min-w-0 flex-1 flex-col gap-0.5'>
          <div className='flex items-baseline gap-2'>
            <span className='font-mono text-sm font-semibold'>
              {item.route}
            </span>
            <span className='text-muted-foreground/60 hidden truncate text-xs md:inline'>
              {item.description}
            </span>
          </div>
          <span className='text-muted-foreground/40 truncate font-mono text-xs'>
            {item.url}
          </span>
        </div>
      </div>

      <div className='flex shrink-0 items-center gap-2'>
        <div className='flex items-center'>
          {status.testing && (
            <StatusBadge variant='warning'>{t('Testing...')}</StatusBadge>
          )}
          {status.latency !== null && !status.testing && (
            <StatusBadge variant={latencyVariant}>
              {status.latency}
              {t('ms')}
            </StatusBadge>
          )}
          {status.error && (
            <StatusBadge variant='neutral'>{t('N/A')}</StatusBadge>
          )}
        </div>

        <div className='flex items-center gap-0.5'>
          <Button
            variant='ghost'
            size='icon-sm'
            onClick={() => props.onTest(item.url)}
            disabled={status.testing}
            title={t('Test Latency')}
          >
            <Zap
              className={cn('size-3.5', status.testing && 'animate-pulse')}
            />
          </Button>

          <Button
            variant='ghost'
            size='icon-sm'
            onClick={() => openExternalSpeedTest(item.url)}
            className='hidden sm:inline-flex'
            title={t('External Speed Test')}
          >
            <Gauge className='size-3.5' />
          </Button>

          <CopyButton
            value={item.url}
            variant='ghost'
            size='icon-sm'
            iconClassName='size-3.5'
            tooltip={t('Copy URL')}
            aria-label={t('Copy URL')}
          />

          <Button
            variant='ghost'
            size='icon-sm'
            className='hidden sm:inline-flex'
            title={t('Open in New Tab')}
            render={<a href={item.url} target='_blank' rel='noreferrer' />}
          >
            <ExternalLink className='size-3.5' />
          </Button>
        </div>
      </div>
    </div>
  )
}
