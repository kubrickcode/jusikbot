export type Settings = {
  $schema?: string
  adjustment_unit_krw: number
  anchoring: {
    monthly_max_change_per_position_krw: number
    monthly_max_total_change_krw: number
    quarterly_max_change_per_position_krw: number
    quarterly_max_total_change_krw: number
  }
  budget_krw: number
  holdings_staleness_threshold_days: number
  min_review_interval_days: number
  risk_tolerance: {
    max_drawdown_action_pct: number
    max_drawdown_warning_pct: number
    max_sector_concentration_pct: number
    max_single_etf_pct: number
    max_single_stock_pct: number
    min_position_size_krw: number
  }
  sizing: {
    high_confidence_pool_pct: number
    low_confidence_pool_pct: number
    medium_confidence_pool_pct: number
  }
  strategy: {
    core_internal_ratio: Record<string, number>
    core_pct: number
    satellite_pct: number
  }
}

export type EtfOption = {
  symbol: string
  name: string
}
