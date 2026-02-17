import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { mkdir, readFile, rm, writeFile } from "fs/promises"
import { join } from "path"
import {
  readConfigFile,
  writeConfigFile,
  validateConfig,
  type ConfigValidationError,
} from "../config-file"

const TEST_DIR = join(__dirname, "__fixtures__")
const TEST_FILE = join(TEST_DIR, "test-settings.json")
const TEST_SCHEMA = join(TEST_DIR, "test-settings.schema.json")

const validSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  required: ["budget_krw", "risk_tolerance"],
  properties: {
    budget_krw: {
      type: "integer",
      minimum: 100000,
    },
    adjustment_unit_krw: {
      type: "integer",
      minimum: 10000,
    },
    risk_tolerance: {
      type: "object",
      required: ["max_single_stock_pct"],
      properties: {
        max_single_stock_pct: {
          type: "number",
          minimum: 0,
          maximum: 100,
        },
        max_single_etf_pct: {
          type: "number",
          minimum: 0,
          maximum: 100,
        },
      },
    },
  },
}

const validSettings = {
  budget_krw: 5000000,
  adjustment_unit_krw: 100000,
  risk_tolerance: {
    max_single_stock_pct: 30,
    max_single_etf_pct: 50,
  },
}

beforeEach(async () => {
  await mkdir(TEST_DIR, { recursive: true })
  await writeFile(TEST_SCHEMA, JSON.stringify(validSchema, null, 2))
})

afterEach(async () => {
  await rm(TEST_DIR, { recursive: true, force: true })
})

describe("readConfigFile", () => {
  it("reads and parses JSON file", async () => {
    await writeFile(TEST_FILE, JSON.stringify(validSettings, null, 2))

    const result = await readConfigFile(TEST_FILE)

    expect(result).toEqual(validSettings)
  })

  it("throws for non-existent file", async () => {
    await expect(readConfigFile(join(TEST_DIR, "missing.json"))).rejects.toThrow()
  })

  it("throws for invalid JSON", async () => {
    await writeFile(TEST_FILE, "{ invalid json }")

    await expect(readConfigFile(TEST_FILE)).rejects.toThrow()
  })
})

describe("validateConfig", () => {
  it("returns no errors for valid data", async () => {
    const errors = await validateConfig(validSettings, TEST_SCHEMA)

    expect(errors).toHaveLength(0)
  })

  it("returns errors for missing required field", async () => {
    const invalid = { budget_krw: 5000000 }

    const errors = await validateConfig(invalid, TEST_SCHEMA)

    expect(errors.length).toBeGreaterThan(0)
    expect(errors.some((e: ConfigValidationError) => e.path.includes("risk_tolerance"))).toBe(
      true,
    )
  })

  it("returns errors for type violation", async () => {
    const invalid = {
      budget_krw: "not a number",
      risk_tolerance: { max_single_stock_pct: 30, max_single_etf_pct: 50 },
    }

    const errors = await validateConfig(invalid, TEST_SCHEMA)

    expect(errors.length).toBeGreaterThan(0)
    expect(errors.some((e: ConfigValidationError) => e.path.includes("budget_krw"))).toBe(true)
  })

  it("returns errors for value below minimum", async () => {
    const invalid = {
      budget_krw: 100,
      risk_tolerance: { max_single_stock_pct: 30, max_single_etf_pct: 50 },
    }

    const errors = await validateConfig(invalid, TEST_SCHEMA)

    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0].message).toContain("100000")
  })

  it("returns errors for value above maximum", async () => {
    const invalid = {
      budget_krw: 5000000,
      risk_tolerance: { max_single_stock_pct: 150, max_single_etf_pct: 50 },
    }

    const errors = await validateConfig(invalid, TEST_SCHEMA)

    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0].path).toContain("max_single_stock_pct")
  })

  it("returns errors for nested required field", async () => {
    const invalid = {
      budget_krw: 5000000,
      risk_tolerance: {},
    }

    const errors = await validateConfig(invalid, TEST_SCHEMA)

    expect(errors.length).toBeGreaterThan(0)
  })
})

describe("writeConfigFile", () => {
  it("writes JSON with prettier formatting and preserves data", async () => {
    await writeConfigFile(TEST_FILE, validSettings)

    const written = await readFile(TEST_FILE, "utf-8")
    const parsed = JSON.parse(written)

    expect(parsed).toEqual(validSettings)
  })

  it("maintains consistent formatting on re-write", async () => {
    await writeConfigFile(TEST_FILE, validSettings)
    const firstWrite = await readFile(TEST_FILE, "utf-8")

    await writeConfigFile(TEST_FILE, validSettings)
    const secondWrite = await readFile(TEST_FILE, "utf-8")

    expect(firstWrite).toBe(secondWrite)
  })

  it("outputs multi-line JSON with trailing newline", async () => {
    await writeConfigFile(TEST_FILE, validSettings)

    const written = await readFile(TEST_FILE, "utf-8")

    expect(written).toMatch(/^\{/)
    expect(written).toMatch(/\}\n$/)
    expect(written.split("\n").length).toBeGreaterThan(2)
  })

  it("sorts top-level keys alphabetically", async () => {
    const unordered = {
      risk_tolerance: { max_single_stock_pct: 30, max_single_etf_pct: 50 },
      budget_krw: 5000000,
      adjustment_unit_krw: 100000,
    }

    await writeConfigFile(TEST_FILE, unordered)

    const written = await readFile(TEST_FILE, "utf-8")
    const adjustmentIndex = written.indexOf("adjustment_unit_krw")
    const budgetIndex = written.indexOf("budget_krw")
    const riskIndex = written.indexOf("risk_tolerance")

    expect(adjustmentIndex).toBeLessThan(budgetIndex)
    expect(budgetIndex).toBeLessThan(riskIndex)
  })

  it("sorts nested keys alphabetically", async () => {
    const unordered = {
      budget_krw: 5000000,
      adjustment_unit_krw: 100000,
      risk_tolerance: { max_single_stock_pct: 30, max_single_etf_pct: 50 },
    }

    await writeConfigFile(TEST_FILE, unordered)

    const written = await readFile(TEST_FILE, "utf-8")
    const etfIndex = written.indexOf("max_single_etf_pct")
    const stockIndex = written.indexOf("max_single_stock_pct")

    expect(etfIndex).toBeLessThan(stockIndex)
  })

  it("preserves $schema field at top", async () => {
    const withSchema = {
      $schema: "./settings.schema.json",
      ...validSettings,
    }

    await writeConfigFile(TEST_FILE, withSchema)

    const written = await readFile(TEST_FILE, "utf-8")
    const schemaIndex = written.indexOf('"$schema"')
    const budgetIndex = written.indexOf('"adjustment_unit_krw"')

    expect(schemaIndex).toBeLessThan(budgetIndex)
  })
})
