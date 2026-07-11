import type { Table } from '@tanstack/react-table'
import type { ReactNode } from 'react'

import {
  DataTableFilterField,
  DataTableFilterInput,
  DataTableFilterPanel,
} from '@/components/data-table'

interface LogsFilterToolbarProps<TData> {
  table: Table<TData>
  primaryFilters: ReactNode
  advancedFilters?: ReactNode
  mobilePinnedFilters?: ReactNode
  mobileFilters?: ReactNode
  mobileFilterCount?: number
  stats?: ReactNode
  actionStart?: ReactNode
  hasActiveFilters: boolean
  hasAdvancedActiveFilters?: boolean
  advancedFilterCount?: number
  searchLoading?: boolean
  onReset: () => void
  onSearch: () => void
  className?: string
}

export const LogsFilterField = DataTableFilterField
export const LogsFilterInput = DataTableFilterInput

export function LogsFilterToolbar<TData>(props: LogsFilterToolbarProps<TData>) {
  return <DataTableFilterPanel {...props} />
}
