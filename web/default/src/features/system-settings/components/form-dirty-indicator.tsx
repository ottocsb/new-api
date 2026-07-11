import { useTranslation } from 'react-i18next'

import { StatusBadge } from '@/components/status-badge'

import { SettingsPageTitleStatusPortal } from './settings-page-context'

type FormDirtyIndicatorProps = {
  isDirty: boolean
  message?: string
}

/**
 * Compact page-title status indicator for unsaved form changes.
 *
 * @example
 * ```tsx
 * <FormDirtyIndicator isDirty={form.formState.isDirty} />
 * ```
 */
export function FormDirtyIndicator({
  isDirty,
  message,
}: FormDirtyIndicatorProps) {
  const { t } = useTranslation()
  if (!isDirty) return null

  return (
    <SettingsPageTitleStatusPortal>
      <StatusBadge variant='warning'>
        {message ? t(message) : t('Unsaved changes')}
      </StatusBadge>
    </SettingsPageTitleStatusPortal>
  )
}
