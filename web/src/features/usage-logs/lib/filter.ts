/**
 * Utility functions for usage logs filters
 */
import type { CommonLogFilters } from '../types'

// ============================================================================
// Filter Building Functions
// ============================================================================

/**
 * Build search params from common log filters
 */
export function buildSearchParams(
  filters: CommonLogFilters
): Record<string, unknown> {
  return {
    ...(filters.startTime && { startTime: filters.startTime.getTime() }),
    ...(filters.endTime && { endTime: filters.endTime.getTime() }),
    ...(filters.channel && { channel: filters.channel }),
    ...(filters.model && { model: filters.model }),
    ...(filters.token && { token: filters.token }),
    ...(filters.group && { group: filters.group }),
    ...(filters.username && { username: filters.username }),
    ...(filters.requestId && { requestId: filters.requestId }),
    ...(filters.upstreamRequestId && {
      upstreamRequestId: filters.upstreamRequestId,
    }),
  }
}
