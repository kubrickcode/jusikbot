import { readFile, readdir } from "fs/promises"
import { resolve } from "path"
import { NextResponse } from "next/server"
import { getPool } from "@/lib/db"
import { configPaths, DATA_DIR, OUTPUT_DIR } from "@/lib/paths"
import type { DashboardData, ThesisSummary } from "@/lib/types/dashboard"
import type { Holdings } from "@/lib/types/holdings"
import type { ThesisCheckResult, ThesisStatus } from "@/lib/types/research"
import type { WatchlistItem } from "@/lib/types/watchlist"

const REPORTS_DIR = resolve(OUTPUT_DIR, "reports")
const THESIS_CHECK_PATH = resolve(DATA_DIR, "thesis-check.json")

export async function GET() {
  const [report, theses, watchlist, holdings, collection, fxRate] =
    await Promise.all([
      loadLatestReport(),
      loadThesisSummary(),
      loadWatchlistSummary(),
      loadHoldingsSummary(),
      loadLatestCollection(),
      loadLatestFxRate(),
    ])

  const data: DashboardData = {
    collection,
    fxRate,
    holdings,
    report,
    theses,
    watchlist,
  }

  return NextResponse.json(data)
}

async function loadLatestReport(): Promise<DashboardData["report"]> {
  try {
    const files = await readdir(REPORTS_DIR)
    const dates = files
      .filter((f) => /^\d{4}-\d{2}-\d{2}\.md$/.test(f))
      .map((f) => f.replace(".md", ""))
      .sort((a, b) => b.localeCompare(a))

    return { latestDate: dates[0] ?? null }
  } catch {
    return { latestDate: null }
  }
}

async function loadThesisSummary(): Promise<ThesisSummary> {
  const empty: ThesisSummary = {
    checkedAt: null,
    counts: { valid: 0, weakening: 0, invalidated: 0 },
    total: 0,
  }

  try {
    const raw = await readFile(THESIS_CHECK_PATH, "utf-8")
    const result = JSON.parse(raw) as ThesisCheckResult

    const counts: Record<ThesisStatus, number> = {
      valid: 0,
      weakening: 0,
      invalidated: 0,
    }

    for (const thesis of result.theses) {
      counts[thesis.status]++
    }

    return {
      checkedAt: result.checked_at,
      counts,
      total: result.theses.length,
    }
  } catch {
    return empty
  }
}

async function loadWatchlistSummary(): Promise<DashboardData["watchlist"]> {
  try {
    const raw = await readFile(configPaths.watchlist, "utf-8")
    const items = JSON.parse(raw) as WatchlistItem[]

    const byMarket: Record<string, number> = {}
    for (const item of items) {
      byMarket[item.market] = (byMarket[item.market] ?? 0) + 1
    }

    return { byMarket, count: items.length }
  } catch {
    return { byMarket: {}, count: 0 }
  }
}

async function loadHoldingsSummary(): Promise<DashboardData["holdings"]> {
  try {
    const raw = await readFile(configPaths.holdings, "utf-8")
    const holdings = JSON.parse(raw) as Holdings

    return {
      asOf: holdings.as_of ?? null,
      positionCount: Object.keys(holdings.positions).length,
    }
  } catch {
    return { asOf: null, positionCount: 0 }
  }
}

async function loadLatestCollection(): Promise<DashboardData["collection"]> {
  try {
    const pool = getPool()
    const result = await pool.query(
      "SELECT MAX(date)::text AS latest_date FROM price_history",
    )
    return { latestDate: result.rows[0]?.latest_date ?? null }
  } catch {
    return { latestDate: null }
  }
}

async function loadLatestFxRate(): Promise<DashboardData["fxRate"]> {
  try {
    const pool = getPool()
    const result = await pool.query(
      "SELECT date::text, pair, rate FROM fx_rate ORDER BY date DESC LIMIT 1",
    )
    const row = result.rows[0]
    if (!row) return { date: null, pair: "USDKRW", rate: null }
    return { date: row.date, pair: row.pair, rate: row.rate }
  } catch {
    return { date: null, pair: "USDKRW", rate: null }
  }
}
