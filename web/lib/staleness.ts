import type { StalenessLevel } from "@/lib/types/dashboard"

export function daysBetween(dateString: string, referenceDate: Date): number {
  const target = new Date(dateString + "T00:00:00Z")
  const reference = new Date(
    referenceDate.toISOString().slice(0, 10) + "T00:00:00Z",
  )
  const diffMs = reference.getTime() - target.getTime()
  return Math.floor(diffMs / (1000 * 60 * 60 * 24))
}

type StalenessThreshold = {
  staleAfterDays: number
  criticalAfterDays: number
}

const STALENESS_THRESHOLDS: Record<string, StalenessThreshold> = {
  collection: { staleAfterDays: 3, criticalAfterDays: 7 },
  fxRate: { staleAfterDays: 3, criticalAfterDays: 7 },
  holdings: { staleAfterDays: 30, criticalAfterDays: 60 },
  report: { staleAfterDays: 7, criticalAfterDays: 30 },
  theses: { staleAfterDays: 14, criticalAfterDays: 30 },
}

export function computeStaleness(
  category: keyof typeof STALENESS_THRESHOLDS,
  dateString: string | null,
  now: Date = new Date(),
): StalenessLevel {
  if (!dateString) return "critical"

  const days = daysBetween(dateString, now)
  const threshold = STALENESS_THRESHOLDS[category]

  if (days >= threshold.criticalAfterDays) return "critical"
  if (days >= threshold.staleAfterDays) return "stale"
  return "fresh"
}

export function formatDaysAgo(dateString: string | null, now: Date = new Date()): string {
  if (!dateString) return "---"

  const days = daysBetween(dateString, now)
  if (days === 0) return "오늘"
  if (days === 1) return "어제"
  return `${days}일 전`
}
