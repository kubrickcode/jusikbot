import { NextResponse } from "next/server"
import { readFile } from "fs/promises"
import { resolve } from "path"
import { DATA_DIR } from "@/lib/paths"

const SUMMARY_PATH = resolve(DATA_DIR, "summary.md")

export async function GET() {
  try {
    const content = await readFile(SUMMARY_PATH, "utf-8")
    return NextResponse.json({ content })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to read summary"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
