import { readdir, readFile } from "fs/promises"
import { resolve } from "path"
import { NextRequest, NextResponse } from "next/server"
import { DATA_DIR } from "@/lib/paths"
import type { HistoryEntry } from "@/lib/types/research"

const HISTORY_DIR = resolve(DATA_DIR, "research-history")

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl
  const date = searchParams.get("date")

  if (date) {
    return loadSnapshot(date)
  }

  return listSnapshots()
}

async function listSnapshots(): Promise<NextResponse> {
  try {
    const files = await readdir(HISTORY_DIR)
    const entries: HistoryEntry[] = files
      .filter((f) => f.startsWith("thesis-check-") && f.endsWith(".json"))
      .map((f) => ({
        filename: f,
        date: f.replace("thesis-check-", "").replace(".json", ""),
      }))
      .sort((a, b) => b.date.localeCompare(a.date))

    return NextResponse.json({ entries })
  } catch (err) {
    if (err instanceof Error && "code" in err && (err as NodeJS.ErrnoException).code === "ENOENT") {
      return NextResponse.json({ entries: [] })
    }
    return NextResponse.json({ error: "읽기 실패" }, { status: 500 })
  }
}

async function loadSnapshot(date: string): Promise<NextResponse> {
  const safeName = date.replace(/[^0-9-]/g, "")
  const filePath = resolve(HISTORY_DIR, `thesis-check-${safeName}.json`)

  try {
    const raw = await readFile(filePath, "utf-8")
    const data = JSON.parse(raw)
    return NextResponse.json(data)
  } catch (err) {
    if (err instanceof Error && "code" in err && (err as NodeJS.ErrnoException).code === "ENOENT") {
      return NextResponse.json({ error: "스냅샷 없음" }, { status: 404 })
    }
    return NextResponse.json({ error: "읽기 실패" }, { status: 500 })
  }
}
