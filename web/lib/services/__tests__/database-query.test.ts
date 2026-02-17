import { describe, expect, it } from "vitest"
import { buildFxQuery, buildPriceQuery } from "../database-query"

describe("buildPriceQuery", () => {
  it("returns base query with defaults when no filters given", () => {
    const query = buildPriceQuery({})

    expect(query.text).toContain("SELECT")
    expect(query.text).toContain("FROM price_history")
    expect(query.text).toContain("ORDER BY date ASC")
    expect(query.text).toContain("LIMIT")
    expect(query.values).toContain(100)
  })

  it("filters by single symbol", () => {
    const query = buildPriceQuery({ symbols: ["NVDA"] })

    expect(query.text).toContain("symbol = ANY($")
    expect(query.values).toContainEqual(["NVDA"])
  })

  it("filters by multiple symbols", () => {
    const query = buildPriceQuery({ symbols: ["NVDA", "AAPL", "META"] })

    expect(query.text).toContain("symbol = ANY($")
    expect(query.values).toContainEqual(["NVDA", "AAPL", "META"])
  })

  it("filters by from date", () => {
    const query = buildPriceQuery({ from: "2026-01-01" })

    expect(query.text).toContain("date >= $")
    expect(query.values).toContain("2026-01-01")
  })

  it("filters by to date", () => {
    const query = buildPriceQuery({ to: "2026-02-01" })

    expect(query.text).toContain("date <= $")
    expect(query.values).toContain("2026-02-01")
  })

  it("filters by date range", () => {
    const query = buildPriceQuery({ from: "2026-01-01", to: "2026-02-01" })

    expect(query.text).toContain("date >= $")
    expect(query.text).toContain("date <= $")
    expect(query.values).toContain("2026-01-01")
    expect(query.values).toContain("2026-02-01")
  })

  it("combines symbol and date filters", () => {
    const query = buildPriceQuery({
      symbols: ["NVDA"],
      from: "2026-01-01",
      to: "2026-02-01",
    })

    expect(query.text).toContain("symbol = ANY($")
    expect(query.text).toContain("date >= $")
    expect(query.text).toContain("date <= $")
    expect(query.values).toContainEqual(["NVDA"])
    expect(query.values).toContain("2026-01-01")
    expect(query.values).toContain("2026-02-01")
  })

  it("applies custom limit", () => {
    const query = buildPriceQuery({ limit: 50 })

    expect(query.values).toContain(50)
  })

  it("applies offset", () => {
    const query = buildPriceQuery({ offset: 20 })

    expect(query.text).toContain("OFFSET $")
    expect(query.values).toContain(20)
  })

  it("sorts descending", () => {
    const query = buildPriceQuery({ sortDirection: "desc" })

    expect(query.text).toContain("ORDER BY date DESC")
  })

  it("uses sequential parameter indices", () => {
    const query = buildPriceQuery({
      symbols: ["NVDA"],
      from: "2026-01-01",
      to: "2026-02-01",
      limit: 50,
      offset: 10,
    })

    const paramCount = (query.text.match(/\$\d+/g) ?? []).length
    expect(paramCount).toBe(query.values.length)

    for (let i = 1; i <= query.values.length; i++) {
      expect(query.text).toContain(`$${i}`)
    }
  })

  it("caps limit at 1000", () => {
    const query = buildPriceQuery({ limit: 5000 })

    expect(query.values).toContain(1000)
    expect(query.values).not.toContain(5000)
  })
})

describe("buildFxQuery", () => {
  it("returns base query with defaults when no filters given", () => {
    const query = buildFxQuery({})

    expect(query.text).toContain("SELECT")
    expect(query.text).toContain("FROM fx_rate")
    expect(query.text).toContain("ORDER BY date ASC")
    expect(query.text).toContain("LIMIT")
    expect(query.values).toContain(100)
  })

  it("filters by pair", () => {
    const query = buildFxQuery({ pair: "USDKRW" })

    expect(query.text).toContain("pair = $")
    expect(query.values).toContain("USDKRW")
  })

  it("filters by date range", () => {
    const query = buildFxQuery({ from: "2026-01-01", to: "2026-02-01" })

    expect(query.text).toContain("date >= $")
    expect(query.text).toContain("date <= $")
    expect(query.values).toContain("2026-01-01")
    expect(query.values).toContain("2026-02-01")
  })

  it("combines pair and date filters", () => {
    const query = buildFxQuery({
      pair: "USDKRW",
      from: "2026-01-01",
      to: "2026-02-01",
    })

    expect(query.text).toContain("pair = $")
    expect(query.text).toContain("date >= $")
    expect(query.text).toContain("date <= $")
    expect(query.values).toContain("USDKRW")
  })

  it("applies offset and limit", () => {
    const query = buildFxQuery({ limit: 30, offset: 10 })

    expect(query.text).toContain("LIMIT $")
    expect(query.text).toContain("OFFSET $")
    expect(query.values).toContain(30)
    expect(query.values).toContain(10)
  })

  it("uses sequential parameter indices", () => {
    const query = buildFxQuery({
      pair: "USDKRW",
      from: "2026-01-01",
      to: "2026-02-01",
      limit: 50,
      offset: 10,
    })

    const paramCount = (query.text.match(/\$\d+/g) ?? []).length
    expect(paramCount).toBe(query.values.length)
  })

  it("caps limit at 1000", () => {
    const query = buildFxQuery({ limit: 5000 })

    expect(query.values).toContain(1000)
  })
})
