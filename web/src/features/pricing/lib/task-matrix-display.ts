import type { BillingUsageSchema } from '../types'
import type { ParsedTaskTier } from './billing-expr'
import {
  getTaskEnumFields,
  taskMatrixRowLabel,
  tryParseTaskMatrixConfig,
} from './task-expr'

/**
 * Marketplace display helper: expand a recognized task matrix (flat/uniform
 * or a full enum partition) into one row per combination. Returns null when
 * the schema has no enum fields or the expression is not a recognized matrix,
 * so callers keep the raw parsed-tier display.
 */
export function getTaskMatrixDisplayTiers(
  expression: string | null | undefined,
  schema: BillingUsageSchema | null | undefined
): ParsedTaskTier[] | null {
  if (!schema) return null
  if (getTaskEnumFields(schema).length === 0) return null

  const matrix = tryParseTaskMatrixConfig(expression, schema)
  if (!matrix) return null

  return matrix.rows.map((row) => ({
    label: taskMatrixRowLabel(row.combination),
    conditions: Object.entries(row.combination)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([field, value]) => ({ field, value })),
    constant: row.constant,
    unitPrices: { ...row.unitPrices },
  }))
}
