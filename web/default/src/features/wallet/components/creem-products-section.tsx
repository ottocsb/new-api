import { useTranslation } from 'react-i18next'

import { Button } from '@/components/design-system/button'
import { Skeleton } from '@/components/ui/skeleton'
import { formatNumber } from '@/lib/format'

import { formatCreemPrice } from '../lib/format'
import type { CreemProduct } from '../types'

interface CreemProductsSectionProps {
  products: CreemProduct[]
  onProductSelect: (product: CreemProduct) => void
  loading?: boolean
}

export function CreemProductsSection({
  products,
  onProductSelect,
  loading,
}: CreemProductsSectionProps) {
  const { t } = useTranslation()

  if (loading) {
    return (
      <div className='grid grid-cols-2 gap-2 md:grid-cols-3'>
        <Skeleton className='h-16 rounded-lg' />
        <Skeleton className='h-16 rounded-lg' />
        <Skeleton className='h-16 rounded-lg' />
      </div>
    )
  }

  if (!Array.isArray(products) || products.length === 0) {
    return null
  }

  return (
    <div className='grid grid-cols-2 gap-2 md:grid-cols-3'>
      {products.map((product) => (
        <Button
          key={product.productId}
          variant='outline'
          className='h-auto min-h-16 flex-col items-start justify-center gap-1 rounded-lg px-3 py-2.5 text-left whitespace-normal sm:h-auto'
          onClick={() => onProductSelect(product)}
        >
          <span className='flex w-full items-center justify-between gap-2'>
            <span className='truncate text-base font-semibold'>
              {product.name}
            </span>
            <span className='shrink-0 text-sm font-medium tabular-nums'>
              {formatCreemPrice(product.price, product.currency)}
            </span>
          </span>
          <span className='text-muted-foreground w-full truncate text-xs font-normal'>
            {t('Quota')}: {formatNumber(product.quota)}
          </span>
        </Button>
      ))}
    </div>
  )
}
