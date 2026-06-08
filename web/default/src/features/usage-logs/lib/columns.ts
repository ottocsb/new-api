/**
 * Column definitions factory
 */
import type { ColumnDef } from '@tanstack/react-table'
import { useCommonLogsColumns } from '../components/columns/common-logs-columns'
import { useTaskLogsColumns } from '../components/columns/task-logs-columns'
import type { LogCategory } from '../types'

/**
 * Get column definitions based on log category
 * Returns any[] due to different log types (UsageLog, TaskLog)
 */
export function useColumnsByCategory(
  logCategory: LogCategory,
  isAdmin: boolean
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): ColumnDef<any>[] {
  const commonColumns = useCommonLogsColumns(isAdmin)
  const taskColumns = useTaskLogsColumns(isAdmin)

  switch (logCategory) {
    case 'common':
      return commonColumns
    case 'task':
      return taskColumns
    default:
      return commonColumns
  }
}
