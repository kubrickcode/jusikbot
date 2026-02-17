export type PriceFilter = {
  from?: string
  limit?: number
  offset?: number
  sortDirection?: "asc" | "desc"
  symbols?: string[]
  to?: string
}

export type FxFilter = {
  from?: string
  limit?: number
  offset?: number
  pair?: string
  to?: string
}

export type PriceRow = {
  adjClose: number
  close: number
  date: string
  high: number
  isAnomaly: boolean
  low: number
  open: number
  source: string
  symbol: string
  volume: number
}

export type FxRateRow = {
  date: string
  pair: string
  rate: number
  source: string
}

export type ParameterizedQuery = {
  text: string
  values: unknown[]
}
