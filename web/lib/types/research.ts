export type ConditionStatus = "met" | "partially_met" | "not_yet" | "refuted" | "unknown"
export type ConditionType = "validity" | "invalidation"
export type ThesisStatus = "valid" | "weakening" | "invalidated"
export type StatusTransition = "stable" | "improving" | "degrading" | "new"
export type SourceTier = 1 | 2 | 3 | 4

export type ThesisSource = {
  title: string
  url: string
  tier: SourceTier
  date: string
}

export type ThesisCondition = {
  text: string
  type: ConditionType
  status: ConditionStatus
  evidence: string
  sources: ThesisSource[]
  quantitative_distance?: string | null
  previous_status?: ConditionStatus | null
  status_transition?: StatusTransition | null
}

export type Thesis = {
  name: string
  status: ThesisStatus
  previous_status?: ThesisStatus | null
  status_transition?: StatusTransition | null
  upstream_dependency?: string | null
  chain_impact?: string | null
  conditions: ThesisCondition[]
}

export type ThesisCheckResult = {
  checked_at: string
  theses: Thesis[]
}

export type MarketType = "US" | "KR" | "EU"
export type SecurityType = "stock" | "etf"
export type MarketCapCategory = "large" | "mid" | "small"

export type Candidate = {
  symbol: string
  name: string
  market: MarketType
  sector: string
  type: SecurityType
  related_theses: string[]
  rationale: string
  risks: string
  market_cap_category: MarketCapCategory
  already_in_watchlist: boolean
}

export type CandidatesResult = {
  checked_at: string
  candidates: Candidate[]
}

export type HistoryEntry = {
  filename: string
  date: string
}
