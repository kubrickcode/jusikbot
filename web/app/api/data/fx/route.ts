import { NextResponse } from "next/server"
import { getPool } from "@/lib/db"
import { buildFxQuery } from "@/lib/services/database-query"
import type { FxFilter } from "@/lib/types/database"

export async function GET(request: Request) {
  try {
    const { searchParams } = new URL(request.url)

    const filter: FxFilter = {
      from: searchParams.get("from") ?? undefined,
      limit: searchParams.has("limit") ? parseInt(searchParams.get("limit")!, 10) : undefined,
      offset: searchParams.has("offset") ? parseInt(searchParams.get("offset")!, 10) : undefined,
      pair: searchParams.get("pair") ?? undefined,
      to: searchParams.get("to") ?? undefined,
    }

    const query = buildFxQuery(filter)
    const pool = getPool()
    const result = await pool.query(query.text, query.values)

    return NextResponse.json({ rows: result.rows })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to query fx rates"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
