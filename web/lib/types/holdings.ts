export type Currency = "USD" | "KRW"

export type Position = {
  avg_cost: number
  currency: Currency
  quantity: number
}

export type Holdings = {
  $schema?: string
  as_of: string
  positions: Record<string, Position>
}
