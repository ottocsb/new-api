import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { useAuthStore } from '@/stores/auth-store'

import { ConnectedAccountsCard } from './components/connected-accounts-card'
import { LanguagePreferencesCard } from './components/language-preferences-card'
import { NotificationSettingsCard } from './components/notification-settings-card'
import { ProfileHeader } from './components/profile-header'
import { ProfileSecurityCard } from './components/profile-security-card'
import { SidebarModulesCard } from './components/sidebar-modules-card'
import { useProfile } from './hooks'

export function Profile() {
  const { t } = useTranslation()
  const { profile, loading, refreshProfile } = useProfile()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)

  const canConfigureSidebar = permissions?.sidebar_settings !== false

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Profile')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
          <ProfileHeader profile={profile} loading={loading} />

          <div className='grid gap-4 sm:gap-5 xl:grid-cols-3 xl:items-start'>
            <div className='flex flex-col gap-4 sm:gap-5 xl:col-span-2'>
              <ConnectedAccountsCard
                profile={profile}
                loading={loading}
                onUpdate={refreshProfile}
              />
              <ProfileSecurityCard profile={profile} loading={loading} />
              <NotificationSettingsCard
                profile={profile}
                loading={loading}
                onUpdate={refreshProfile}
              />
            </div>

            <div className='flex flex-col gap-4 sm:gap-5'>
              <LanguagePreferencesCard
                profile={profile}
                onProfileUpdate={refreshProfile}
              />
              {canConfigureSidebar && <SidebarModulesCard />}
            </div>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
