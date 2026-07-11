import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/design-system/button'
import { Switch } from '@/components/ui/switch'
import { TitledCard } from '@/components/ui/titled-card'
import { api } from '@/lib/api'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

type SidebarModuleConfig = {
  enabled: boolean
  [key: string]: boolean
}

type SidebarModulesConfig = Record<string, SidebarModuleConfig>

type SectionDef = {
  key: string
  title: string
  description: string
  modules: { key: string; title: string; description: string }[]
}

export function SidebarModulesCard() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [config, setConfig] = useState<SidebarModulesConfig>({})
  const currentUser = useAuthStore((s) => s.auth.user)
  const setUser = useAuthStore((s) => s.auth.setUser)

  const sectionDefs: SectionDef[] = [
    {
      key: 'console',
      title: t('Console Area'),
      description: t('Data management and log viewing'),
      modules: [
        {
          key: 'detail',
          title: t('Dashboard'),
          description: t('System data statistics'),
        },
        {
          key: 'token',
          title: t('Token Management'),
          description: t('API token management'),
        },
        {
          key: 'log',
          title: t('Usage Logs'),
          description: t('API usage records'),
        },
      ],
    },
    {
      key: 'personal',
      title: t('Personal Center Area'),
      description: t('User personal functions'),
      modules: [
        {
          key: 'topup',
          title: t('Wallet Management'),
          description: t('Balance and top-up management'),
        },
        {
          key: 'personal',
          title: t('Personal Settings'),
          description: t('Personal info settings'),
        },
      ],
    },
  ]

  const loadConfig = useCallback(async () => {
    try {
      const res = await api.get('/api/user/self')
      if (res.data.success && res.data.data?.sidebar_modules) {
        const raw = res.data.data.sidebar_modules
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        setConfig(parsed)
      } else {
        const defaults: SidebarModulesConfig = {}
        for (const sec of sectionDefs) {
          defaults[sec.key] = { enabled: true }
          for (const mod of sec.modules) defaults[sec.key][mod.key] = true
        }
        setConfig(defaults)
      }
    } catch {
      /* ignore */
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    loadConfig()
  }, [loadConfig])

  const toggleSection = (sectionKey: string, val: boolean) => {
    setConfig((prev) => ({
      ...prev,
      [sectionKey]: { ...prev[sectionKey], enabled: val },
    }))
  }

  const toggleModule = (
    sectionKey: string,
    moduleKey: string,
    val: boolean
  ) => {
    setConfig((prev) => ({
      ...prev,
      [sectionKey]: { ...prev[sectionKey], [moduleKey]: val },
    }))
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      const serialized = JSON.stringify(config)
      const res = await api.put('/api/user/self', {
        sidebar_modules: serialized,
      })
      if (res.data.success) {
        // Sync to auth-store so useSidebarConfig re-runs and the sidebar
        // updates immediately without needing a page refresh.
        if (currentUser) {
          setUser({ ...currentUser, sidebar_modules: serialized })
        }
        toast.success(t('Saved successfully'))
      } else {
        toast.error(res.data.message || t('Save failed'))
      }
    } catch {
      toast.error(t('Save failed, please retry'))
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    const defaults: SidebarModulesConfig = {}
    for (const sec of sectionDefs) {
      defaults[sec.key] = { enabled: true }
      for (const mod of sec.modules) defaults[sec.key][mod.key] = true
    }
    setConfig(defaults)
    toast.success(t('Reset to default configuration'))
  }

  return (
    <TitledCard
      title={t('Sidebar Personal Settings')}
      description={t('Customize sidebar display content')}
      disableHoverEffect
      contentClassName='space-y-5 sm:space-y-6'
    >
      {sectionDefs.map((section) => {
        const sectionEnabled = config[section.key]?.enabled !== false
        return (
          <div key={section.key} className='space-y-2.5'>
            <div className='flex items-center justify-between gap-3'>
              <div className='min-w-0'>
                <p className='text-sm font-medium'>{section.title}</p>
                <p className='text-muted-foreground text-xs'>
                  {section.description}
                </p>
              </div>
              <Switch
                checked={sectionEnabled}
                onCheckedChange={(v) => toggleSection(section.key, v)}
              />
            </div>
            <div className='divide-border/60 divide-y rounded-lg border'>
              {section.modules.map((mod) => (
                <div
                  key={mod.key}
                  className={cn(
                    'flex items-center justify-between gap-3 px-3 py-2.5',
                    !sectionEnabled && 'opacity-50'
                  )}
                >
                  <div className='min-w-0'>
                    <p className='truncate text-sm'>{mod.title}</p>
                    <p className='text-muted-foreground truncate text-xs'>
                      {mod.description}
                    </p>
                  </div>
                  <Switch
                    checked={config[section.key]?.[mod.key] !== false}
                    onCheckedChange={(v) =>
                      toggleModule(section.key, mod.key, v)
                    }
                    disabled={!sectionEnabled}
                  />
                </div>
              ))}
            </div>
          </div>
        )
      })}

      <div className='flex flex-col-reverse gap-2 border-t pt-4 sm:flex-row sm:justify-end'>
        <Button variant='outline' onClick={handleReset}>
          {t('Reset to Default')}
        </Button>
        <Button onClick={handleSave} disabled={loading}>
          {loading ? t('Saving...') : t('Save Changes')}
        </Button>
      </div>
    </TitledCard>
  )
}
