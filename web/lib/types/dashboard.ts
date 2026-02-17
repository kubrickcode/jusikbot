import type { ThesisStatus } from "./research"

export type ThesisSummary = {
  checkedAt: string | null
  counts: Record<ThesisStatus, number>
  total: number
}

export type DashboardData = {
  collection: {
    latestDate: string | null
  }
  fxRate: {
    date: string | null
    pair: string
    rate: number | null
  }
  holdings: {
    asOf: string | null
    positionCount: number
  }
  report: {
    latestDate: string | null
  }
  theses: ThesisSummary
  watchlist: {
    byMarket: Record<string, number>
    count: number
  }
}

export type StalenessLevel = "fresh" | "stale" | "critical"
