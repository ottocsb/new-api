import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'

interface FooterProps {
  logo?: string
  name?: string
  copyright?: string
  className?: string
  variant?: 'default' | 'compact'
}

export function Footer(props: FooterProps) {
  const { t } = useTranslation()
  const {
    systemName,
    footerHtml,
  } = useSystemConfig()
  const { status } = useStatus()

  const displayName = systemName || props.name || 'New API'
  const currentYear = new Date().getFullYear()

  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)

  const isCompact = props.variant === 'compact'

  if (isCompact) {
    return (
      <footer
        className={cn(
          'border-border/40 bg-background relative z-10 border-t',
          props.className
        )}
      >
        <div className='mx-auto flex w-full max-w-6xl items-center justify-center gap-2 px-4 py-2.5 sm:px-6'>
          <div className='text-muted-foreground/70 flex flex-wrap items-center justify-center gap-x-3 gap-y-1 text-xs'>
            {hasPrivacyPolicy && (
              <Link
                to='/privacy-policy'
                className='hover:text-foreground transition-colors'
              >
                {t('Privacy Policy')}
              </Link>
            )}
            {hasUserAgreement && (
              <Link
                to='/user-agreement'
                className='hover:text-foreground transition-colors'
              >
                {t('User Agreement')}
              </Link>
            )}
            <a
              href='mailto:support@apiwarrior.xyz'
              className='hover:text-foreground transition-colors'
            >
              support@apiwarrior.xyz
            </a>
            <span className='text-muted-foreground/50'>
              &copy; {currentYear} {displayName}
            </span>
          </div>
        </div>
      </footer>
    )
  }

  if (footerHtml) {
    return (
      <footer
        className={cn(
          'border-border/40 relative z-10 border-t',
          props.className
        )}
      >
        <div className='mx-auto w-full max-w-6xl px-6 py-5'>
          <div className='bg-muted/20 border-border/50 flex flex-col items-center justify-between gap-4 rounded-2xl border px-4 py-4 backdrop-blur-sm sm:flex-row sm:px-5'>
            <div
              className='custom-footer text-muted-foreground min-w-0 text-center text-sm sm:text-left'
              dangerouslySetInnerHTML={{ __html: footerHtml }}
            />
          </div>
        </div>
      </footer>
    )
  }

  return (
    <footer
      className={cn('border-border/40 relative z-10 border-t', props.className)}
    >
      <div className='mx-auto max-w-6xl px-6 py-12 md:py-16'>
        {/* Bottom section */}
        <div>
          <div className='text-muted-foreground/40 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 text-xs'>
            <span>
              &copy; {currentYear} {displayName}.{' '}
              {props.copyright ?? t('footer.defaultCopyright')}
            </span>
            {hasPrivacyPolicy && (
              <Link
                to='/privacy-policy'
                className='hover:text-foreground transition-colors'
              >
                {t('Privacy Policy')}
              </Link>
            )}
            {hasUserAgreement && (
              <Link
                to='/user-agreement'
                className='hover:text-foreground transition-colors'
              >
                {t('User Agreement')}
              </Link>
            )}
            <a
              href='mailto:support@apiwarrior.xyz'
              className='hover:text-foreground transition-colors'
            >
              support@apiwarrior.xyz
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
