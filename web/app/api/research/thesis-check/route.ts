import { readFile } from "fs/promises"
import { resolve } from "path"
import { NextResponse } from "next/server"
import { DATA_DIR } from "@/lib/paths"

const THESIS_CHECK_PATH = resolve(DATA_DIR, "thesis-check.json")

export async function GET() {
  try {
    const raw = await readFile(THESIS_CHECK_PATH, "utf-8")
    const data = JSON.parse(raw)
    return NextResponse.json(data)
  } catch (err) {
    if (err instanceof Error && "code" in err && (err as NodeJS.ErrnoException).code === "ENOENT") {
      return NextResponse.json({ exists: false }, { status: 404 })
    }
    return NextResponse.json({ error: "읽기 실패" }, { status: 500 })
  }
}
