import { NextResponse } from "next/server"
import { getPool } from "@/lib/db"
import { buildPriceQuery } from "@/lib/services/database-query"
import type { PriceFilter } from "@/lib/types/database"

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)

    const symbolsParam = searchParams.get("symbols")
    const filter: PriceFilter = {
      from: searchParams.get("from") ?? undefined,
      limit: searchParams.has("limit") ? parseInt(searchParams.get("limit")!, 10) : undefined,
      offset: searchParams.has("offset") ? parseInt(searchParams.get("offset")!, 10) : undefined,
      sortDirection: searchParams.get("sort") === "desc" ? "desc" : "asc",
      symbols: symbolsParam ? symbolsParam.split(",").map((s) => s.trim()) : undefined,
      to: searchParams.get("to") ?? undefined,
    }

    const query = buildPriceQuery(filter)
    const pool = getPool()
    const result = await pool.query(query.text, query.values)

    return NextResponse.json({ rows: result.rows })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to query prices"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
