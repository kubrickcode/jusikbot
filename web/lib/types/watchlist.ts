export type Market = "US" | "KR"
export type SecurityType = "stock" | "etf"

export type WatchlistItem = {
  market: Market
  name: string
  sector: string
  symbol: string
  themes: string[]
  type: SecurityType
}
