/**
 * Status mappers for different log types
 * Centralized mapper instances for consistent usage across components
 */
import {
  TASK_ACTION_MAPPINGS,
  TASK_STATUS_MAPPINGS,
  TASK_PLATFORM_MAPPINGS,
} from '../constants'
import { createStatusMapper } from './status'

// ============================================================================
// Task Logs Mappers
// ============================================================================

/**
 * Task action type mapper
 */
export const taskActionMapper = createStatusMapper(TASK_ACTION_MAPPINGS)

/**
 * Task status mapper
 */
export const taskStatusMapper = createStatusMapper(TASK_STATUS_MAPPINGS)

/**
 * Task platform mapper
 */
export const taskPlatformMapper = createStatusMapper(TASK_PLATFORM_MAPPINGS)
