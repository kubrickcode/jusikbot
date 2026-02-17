import { NextResponse } from "next/server"
import { readFile } from "fs/promises"
import { validateConfig, writeConfigFile } from "@/lib/services/config-file"
import { configPaths } from "@/lib/paths"

export async function GET() {
  try {
    const raw = await readFile(configPaths.watchlist, "utf-8")
    const watchlist = JSON.parse(raw)
    return NextResponse.json(watchlist)
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to read watchlist"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function PUT(request: Request) {
  try {
    const body = await request.json()

    const errors = await validateConfig(body, configPaths.watchlistSchema)
    if (errors.length > 0) {
      return NextResponse.json({ errors }, { status: 400 })
    }

    const sorted = (body as Array<Record<string, unknown>>).sort((a, b) =>
      String(a.symbol).localeCompare(String(b.symbol)),
    )

    const formatted = await formatJsonArray(sorted)
    const { writeFile } = await import("fs/promises")
    await writeFile(configPaths.watchlist, formatted, "utf-8")

    return NextResponse.json({ success: true })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to save watchlist"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

async function formatJsonArray(data: Array<Record<string, unknown>>): Promise<string> {
  const { format } = await import("prettier")
  return format(JSON.stringify(data), {
    parser: "json",
    tabWidth: 2,
    printWidth: 80,
  })
}
