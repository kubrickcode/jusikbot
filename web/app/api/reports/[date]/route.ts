import { readFile } from "fs/promises"
import { resolve } from "path"
import { type NextRequest, NextResponse } from "next/server"
import { OUTPUT_DIR } from "@/lib/paths"

const REPORTS_DIR = resolve(OUTPUT_DIR, "reports")

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ date: string }> },
) {
  const { date } = await params
  const safeName = date.replace(/[^0-9-]/g, "")

  if (!/^\d{4}-\d{2}-\d{2}$/.test(safeName)) {
    return NextResponse.json({ error: "잘못된 날짜 형식" }, { status: 400 })
  }

  const filePath = resolve(REPORTS_DIR, `${safeName}.md`)

  try {
    const content = await readFile(filePath, "utf-8")
    return NextResponse.json({ date: safeName, content })
  } catch (err) {
    if (
      err instanceof Error &&
      "code" in err &&
      (err as NodeJS.ErrnoException).code === "ENOENT"
    ) {
      return NextResponse.json({ error: "리포트 없음" }, { status: 404 })
    }
    return NextResponse.json({ error: "리포트 읽기 실패" }, { status: 500 })
  }
}
