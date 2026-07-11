import { useTranslation } from 'react-i18next'

import { Tabs, TabsList, TabsTrigger } from '@/components/design-system/tabs'
import { PublicPageHeader } from '@/components/layout'

import type { RankingPeriod } from '../types'

const PERIODS: { id: RankingPeriod; labelKey: string }[] = [
  { id: 'today', labelKey: 'Today' },
  { id: 'week', labelKey: 'Week' },
  { id: 'month', labelKey: 'Month' },
  { id: 'year', labelKey: 'Year' },
]

type RankingsHeroProps = {
  period: RankingPeriod
  onPeriodChange: (period: RankingPeriod) => void
}

export function RankingsHero(props: RankingsHeroProps) {
  const { t } = useTranslation()

  return (
    <PublicPageHeader
      title={t('Rankings')}
      description={t(
        'Discover the most-used models and rising vendors on the platform, updated from live usage data.'
      )}
    >
      <Tabs
        value={props.period}
        onValueChange={(value) => {
          if (
            value === 'today' ||
            value === 'week' ||
            value === 'month' ||
            value === 'year'
          ) {
            props.onPeriodChange(value)
          }
        }}
      >
        <TabsList
          variant='line'
          aria-label={t('Period')}
          className='w-full justify-start gap-6 overflow-x-auto overflow-y-hidden border-b p-0'
        >
          {PERIODS.map((period) => (
            <TabsTrigger
              key={period.id}
              value={period.id}
              className='flex-none px-0.5 pb-3'
            >
              {t(period.labelKey)}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>
    </PublicPageHeader>
  )
}
