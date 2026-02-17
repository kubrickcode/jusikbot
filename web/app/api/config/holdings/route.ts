import { NextResponse } from "next/server"
import { readConfigFile, validateConfig, writeConfigFile } from "@/lib/services/config-file"
import { configPaths } from "@/lib/paths"

export async function GET() {
  try {
    const holdings = await readConfigFile(configPaths.holdings)
    return NextResponse.json(holdings)
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to read holdings"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}

export async function PUT(request: Request) {
  try {
    const body = await request.json()

    const errors = await validateConfig(body, configPaths.holdingsSchema)
    if (errors.length > 0) {
      return NextResponse.json({ errors }, { status: 400 })
    }

    await writeConfigFile(configPaths.holdings, body)

    return NextResponse.json({ success: true })
  } catch (error) {
    const message = error instanceof Error ? error.message : "Failed to save holdings"
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
