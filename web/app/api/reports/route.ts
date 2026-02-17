import { readdir } from "fs/promises"
import { resolve } from "path"
import { NextResponse } from "next/server"
import { OUTPUT_DIR } from "@/lib/paths"

const REPORTS_DIR = resolve(OUTPUT_DIR, "reports")

export type ReportEntry = {
  date: string
  filename: string
}

export async function GET() {
  try {
    const files = await readdir(REPORTS_DIR)
    const entries: ReportEntry[] = files
      .filter((f) => /^\d{4}-\d{2}-\d{2}\.md$/.test(f))
      .map((f) => ({
        filename: f,
        date: f.replace(".md", ""),
      }))
      .sort((a, b) => b.date.localeCompare(a.date))

    return NextResponse.json({ entries })
  } catch (err) {
    if (
      err instanceof Error &&
      "code" in err &&
      (err as NodeJS.ErrnoException).code === "ENOENT"
    ) {
      return NextResponse.json({ entries: [] })
    }
    return NextResponse.json({ error: "리포트 목록 읽기 실패" }, { status: 500 })
  }
}
