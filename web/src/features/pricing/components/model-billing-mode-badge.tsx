import { useTranslation } from 'react-i18next'

import { StatusBadge, type StatusVariant } from '@/components/status-badge'

import { getBillingModeLabelKey } from '../lib/billing-mode'
import { isDynamicPricingModel } from '../lib/dynamic-price'
import type { PricingModel } from '../types'

interface ModelBillingModeBadgeProps {
  model: PricingModel
  className?: string
}

export function ModelBillingModeBadge(props: ModelBillingModeBadgeProps) {
  const { t } = useTranslation()
  const labelKey = getBillingModeLabelKey(props.model)
  const label = t(labelKey)
  let variant: StatusVariant = 'purple'

  if (isDynamicPricingModel(props.model)) {
    variant = 'warning'
  } else if (labelKey === 'Token-based') {
    variant = 'info'
  }

  return (
    <StatusBadge
      label={label}
      variant={variant}
      copyable={false}
      size='sm'
      className={props.className}
    />
  )
}
