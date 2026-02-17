import Ajv from "ajv"
import { readFile, writeFile } from "fs/promises"
import { format } from "prettier"

export type ConfigValidationError = {
  path: string
  message: string
}

const ajv = new Ajv({ allErrors: true })

export async function readConfigFile(filePath: string): Promise<Record<string, unknown>> {
  const raw = await readFile(filePath, "utf-8")
  return JSON.parse(raw) as Record<string, unknown>
}

export async function validateConfig(
  data: unknown,
  schemaPath: string,
): Promise<ConfigValidationError[]> {
  const schemaRaw = await readFile(schemaPath, "utf-8")
  const schema = JSON.parse(schemaRaw)

  const validate = ajv.compile(schema)
  const isValid = validate(data)

  if (isValid || !validate.errors) {
    return []
  }

  return validate.errors.map((err) => ({
    path: err.instancePath
      ? err.instancePath.slice(1).replace(/\//g, ".")
      : (err.params?.missingProperty as string) ?? "",
    message: err.message ?? "Validation failed",
  }))
}

function sortKeysRecursively(obj: unknown): unknown {
  if (obj === null || typeof obj !== "object" || Array.isArray(obj)) {
    return obj
  }

  const record = obj as Record<string, unknown>
  const sorted: Record<string, unknown> = {}

  const keys = Object.keys(record).sort((a, b) => {
    if (a === "$schema") return -1
    if (b === "$schema") return 1
    return a.localeCompare(b)
  })

  for (const key of keys) {
    sorted[key] = sortKeysRecursively(record[key])
  }

  return sorted
}

export async function writeConfigFile(
  filePath: string,
  data: Record<string, unknown>,
): Promise<void> {
  const sorted = sortKeysRecursively(data)
  const formatted = await format(JSON.stringify(sorted), {
    parser: "json",
    tabWidth: 2,
    printWidth: 80,
  })
  await writeFile(filePath, formatted, "utf-8")
}
