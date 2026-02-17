import type { FxFilter, ParameterizedQuery, PriceFilter } from "@/lib/types/database"

const MAX_LIMIT = 1000
const DEFAULT_LIMIT = 100

const PRICE_COLUMNS = [
  "adj_close AS \"adjClose\"",
  "close",
  "date::text",
  "high",
  "is_anomaly AS \"isAnomaly\"",
  "low",
  "open",
  "source",
  "symbol",
  "volume",
].join(", ")

const FX_COLUMNS = [
  "date::text",
  "pair",
  "rate",
  "source",
].join(", ")

export function buildPriceQuery(filter: PriceFilter): ParameterizedQuery {
  const conditions: string[] = []
  const values: unknown[] = []
  let paramIndex = 1

  if (filter.symbols && filter.symbols.length > 0) {
    conditions.push(`symbol = ANY($${paramIndex})`)
    values.push(filter.symbols)
    paramIndex++
  }

  if (filter.from) {
    conditions.push(`date >= $${paramIndex}`)
    values.push(filter.from)
    paramIndex++
  }

  if (filter.to) {
    conditions.push(`date <= $${paramIndex}`)
    values.push(filter.to)
    paramIndex++
  }

  const direction = filter.sortDirection === "desc" ? "DESC" : "ASC"
  const limit = Math.min(filter.limit ?? DEFAULT_LIMIT, MAX_LIMIT)

  let text = `SELECT ${PRICE_COLUMNS} FROM price_history`

  if (conditions.length > 0) {
    text += ` WHERE ${conditions.join(" AND ")}`
  }

  text += ` ORDER BY date ${direction}`
  text += ` LIMIT $${paramIndex}`
  values.push(limit)
  paramIndex++

  if (filter.offset && filter.offset > 0) {
    text += ` OFFSET $${paramIndex}`
    values.push(filter.offset)
  }

  return { text, values }
}

export function buildFxQuery(filter: FxFilter): ParameterizedQuery {
  const conditions: string[] = []
  const values: unknown[] = []
  let paramIndex = 1

  if (filter.pair) {
    conditions.push(`pair = $${paramIndex}`)
    values.push(filter.pair)
    paramIndex++
  }

  if (filter.from) {
    conditions.push(`date >= $${paramIndex}`)
    values.push(filter.from)
    paramIndex++
  }

  if (filter.to) {
    conditions.push(`date <= $${paramIndex}`)
    values.push(filter.to)
    paramIndex++
  }

  const limit = Math.min(filter.limit ?? DEFAULT_LIMIT, MAX_LIMIT)

  let text = `SELECT ${FX_COLUMNS} FROM fx_rate`

  if (conditions.length > 0) {
    text += ` WHERE ${conditions.join(" AND ")}`
  }

  text += ` ORDER BY date ASC`
  text += ` LIMIT $${paramIndex}`
  values.push(limit)
  paramIndex++

  if (filter.offset && filter.offset > 0) {
    text += ` OFFSET $${paramIndex}`
    values.push(filter.offset)
  }

  return { text, values }
}
