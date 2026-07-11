import { useTranslation } from 'react-i18next'

import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { formatQuota } from '@/lib/format'

import type { UserWalletData } from '../types'

interface WalletStatsCardProps {
  user: UserWalletData | null
  loading?: boolean
}

export function WalletStatsCard(props: WalletStatsCardProps) {
  const { t } = useTranslation()

  if (props.loading) {
    return (
      <Card data-card-hover='false' className='gap-0 py-0'>
        <div className='divide-border/60 grid grid-cols-3 divide-x'>
          {['balance', 'usage', 'requests'].map((key) => (
            <div key={key} className='px-4 py-3 sm:px-5 sm:py-4'>
              <Skeleton className='h-4 w-24' />
              <Skeleton className='mt-2 h-7 w-28' />
            </div>
          ))}
        </div>
      </Card>
    )
  }

  const stats = [
    {
      label: t('Current Balance'),
      value: formatQuota(props.user?.quota ?? 0),
    },
    {
      label: t('Total Usage'),
      value: formatQuota(props.user?.used_quota ?? 0),
    },
    {
      label: t('API Requests'),
      value: (props.user?.request_count ?? 0).toLocaleString(),
    },
  ]

  return (
    <Card data-card-hover='false' className='gap-0 py-0'>
      <div className='divide-border/60 grid grid-cols-3 divide-x'>
        {stats.map((item) => (
          <div key={item.label} className='min-w-0 px-4 py-3 sm:px-5 sm:py-4'>
            <div className='text-muted-foreground truncate text-sm'>
              {item.label}
            </div>
            <div className='text-foreground mt-1 truncate text-lg font-semibold tracking-tight tabular-nums sm:text-2xl'>
              {item.value}
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
