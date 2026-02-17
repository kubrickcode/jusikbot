import { describe, expect, it } from "vitest"
import { computeStaleness, daysBetween, formatDaysAgo } from "../staleness"

describe("daysBetween", () => {
  const reference = new Date("2026-02-17T12:00:00Z")

  it("returns 0 for same day", () => {
    expect(daysBetween("2026-02-17", reference)).toBe(0)
  })

  it("returns 1 for yesterday", () => {
    expect(daysBetween("2026-02-16", reference)).toBe(1)
  })

  it("returns 7 for one week ago", () => {
    expect(daysBetween("2026-02-10", reference)).toBe(7)
  })

  it("returns 30 for thirty days ago", () => {
    expect(daysBetween("2026-01-18", reference)).toBe(30)
  })
})

describe("computeStaleness", () => {
  const now = new Date("2026-02-17T12:00:00Z")

  it("returns critical when date is null", () => {
    expect(computeStaleness("collection", null, now)).toBe("critical")
  })

  it("returns fresh for collection data from today", () => {
    expect(computeStaleness("collection", "2026-02-17", now)).toBe("fresh")
  })

  it("returns fresh for collection data from 2 days ago", () => {
    expect(computeStaleness("collection", "2026-02-15", now)).toBe("fresh")
  })

  it("returns stale for collection data from 4 days ago", () => {
    expect(computeStaleness("collection", "2026-02-13", now)).toBe("stale")
  })

  it("returns critical for collection data from 8 days ago", () => {
    expect(computeStaleness("collection", "2026-02-09", now)).toBe("critical")
  })

  it("returns fresh for holdings updated 15 days ago", () => {
    expect(computeStaleness("holdings", "2026-02-02", now)).toBe("fresh")
  })

  it("returns stale for holdings updated 35 days ago", () => {
    expect(computeStaleness("holdings", "2026-01-13", now)).toBe("stale")
  })

  it("returns critical for holdings updated 65 days ago", () => {
    expect(computeStaleness("holdings", "2025-12-14", now)).toBe("critical")
  })

  it("returns fresh for theses checked 10 days ago", () => {
    expect(computeStaleness("theses", "2026-02-07", now)).toBe("fresh")
  })

  it("returns stale for theses checked 20 days ago", () => {
    expect(computeStaleness("theses", "2026-01-28", now)).toBe("stale")
  })

  it("returns critical for theses checked 35 days ago", () => {
    expect(computeStaleness("theses", "2026-01-13", now)).toBe("critical")
  })
})

describe("formatDaysAgo", () => {
  const now = new Date("2026-02-17T12:00:00Z")

  it("returns --- for null date", () => {
    expect(formatDaysAgo(null, now)).toBe("---")
  })

  it("returns 오늘 for today", () => {
    expect(formatDaysAgo("2026-02-17", now)).toBe("오늘")
  })

  it("returns 어제 for yesterday", () => {
    expect(formatDaysAgo("2026-02-16", now)).toBe("어제")
  })

  it("returns N일 전 for older dates", () => {
    expect(formatDaysAgo("2026-02-10", now)).toBe("7일 전")
  })
})
