import { configPaths } from "@/lib/paths"
import { readFile, writeFile } from "fs/promises"
import { NextResponse } from "next/server"
import { format } from "prettier"

export async function GET() {
  try {
    const content = await readFile(configPaths.theses, "utf-8")
    return NextResponse.json({ content })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to read theses"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function PUT(request: Request) {
  try {
    const body = await request.json()

    if (typeof body.content !== "string") {
      return NextResponse.json({ error: "content must be a string" }, { status: 400 })
    }

    if (body.content.trim().length === 0) {
      return NextResponse.json({ error: "내용이 비어있습니다" }, { status: 400 })
    }

    const formatted = await format(body.content, {
      parser: "markdown",
      tabWidth: 2,
      printWidth: 80,
      proseWrap: "preserve",
    })

    await writeFile(configPaths.theses, formatted, "utf-8")

    return NextResponse.json({ success: true, content: formatted })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to save theses"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
