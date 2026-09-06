import {
  normalizeTierLabel,
  type ParsedTaskTier,
  type ParsedTier,
} from './billing-expr'

type BreakdownMatchTier = ParsedTier | ParsedTaskTier

function tierMatchesNormalizedLabel(
  tier: BreakdownMatchTier,
  normalizedMatchedTierLabel: string
): boolean {
  return (
    normalizedMatchedTierLabel !== '' &&
    normalizeTierLabel(tier.label) === normalizedMatchedTierLabel
  )
}

/**
 * Decide whether a price-table row is the settlement hit.
 * Label equality (after normalizeTierLabel) wins; usage-facts matching is
 * only a fallback when no display row still carries the engine label.
 */
export function isBreakdownTierMatched(
  tier: BreakdownMatchTier,
  tiers: readonly BreakdownMatchTier[],
  matchedTierLabel?: string | null,
  usageFacts?: Record<string, string | number>
): boolean {
  const normalizedMatchedTierLabel = normalizeTierLabel(
    matchedTierLabel ?? undefined
  )
  if (tierMatchesNormalizedLabel(tier, normalizedMatchedTierLabel)) {
    return true
  }
  if (
    tiers.some((candidate) =>
      tierMatchesNormalizedLabel(candidate, normalizedMatchedTierLabel)
    )
  ) {
    return false
  }
  if (!usageFacts || tier.conditions.length === 0) {
    return false
  }
  return tier.conditions.every((condition) => {
    if (!('field' in condition)) {
      return false
    }
    const fact = usageFacts[condition.field]
    if (fact === undefined) {
      return false
    }
    return String(fact) === condition.value
  })
}
