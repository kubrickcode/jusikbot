import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { PriceDataTable } from "../price-data-table"

const MOCK_ROWS = [
  {
    adjClose: 182.81,
    close: 182.81,
    date: "2026-02-14",
    high: 185.0,
    isAnomaly: false,
    low: 180.5,
    open: 183.0,
    source: "tiingo",
    symbol: "NVDA",
    volume: 45000000,
  },
  {
    adjClose: 639.77,
    close: 639.77,
    date: "2026-02-14",
    high: 645.0,
    isAnomaly: false,
    low: 635.0,
    open: 640.0,
    source: "tiingo",
    symbol: "META",
    volume: 12000000,
  },
  {
    adjClose: 98500,
    close: 98500,
    date: "2026-02-14",
    high: 99000,
    isAnomaly: true,
    low: 97500,
    open: 98000,
    source: "kis",
    symbol: "360750",
    volume: 5000000,
  },
]

const DEFAULT_FILTER = {
  from: "2026-01-01",
  symbols: ["NVDA", "META", "360750"],
  to: "2026-02-14",
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe("PriceDataTable", () => {
  it("표 헤더를 렌더링한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_ROWS }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    expect(screen.getByText("날짜")).toBeInTheDocument()
    expect(screen.getByText("종목")).toBeInTheDocument()
    expect(screen.getByText("시가")).toBeInTheDocument()
    expect(screen.getByText("고가")).toBeInTheDocument()
    expect(screen.getByText("저가")).toBeInTheDocument()
    expect(screen.getByText("종가")).toBeInTheDocument()
    expect(screen.getByText("수정종가")).toBeInTheDocument()
    expect(screen.getByText("거래량")).toBeInTheDocument()
  })

  it("데이터를 조회하여 테이블에 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_ROWS }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("NVDA")).toBeInTheDocument()
    })

    expect(screen.getByText("META")).toBeInTheDocument()
    expect(screen.getByText("360750")).toBeInTheDocument()
  })

  it("US 종목은 소수점 2자리, KR 종목은 정수로 포맷한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_ROWS }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("NVDA")).toBeInTheDocument()
    })

    // US: 2 decimal places (close + adjClose both show 182.81)
    const matches = screen.getAllByText("182.81")
    expect(matches.length).toBeGreaterThanOrEqual(1)
    // KR: integer with comma separator (close + adjClose both show 98,500)
    const krMatches = screen.getAllByText("98,500")
    expect(krMatches.length).toBeGreaterThanOrEqual(1)
  })

  it("이상값 행에 경고 아이콘을 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_ROWS }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("360750")).toBeInTheDocument()
    })

    // TriangleAlert icon renders next to the anomaly symbol
    const row360750 = screen.getByText("360750").closest("tr")!
    expect(row360750.querySelector("svg")).toBeInTheDocument()
  })

  it("데이터가 없으면 빈 상태를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: [] }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("데이터 없음")).toBeInTheDocument()
    })
  })

  it("API 에러 시 에러 메시지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ error: "DB 연결 실패" }), { status: 500 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("DB 연결 실패")).toBeInTheDocument()
    })
  })

  it("로딩 중 스켈레톤을 표시한다", () => {
    vi.spyOn(global, "fetch").mockReturnValueOnce(new Promise(() => {}))

    const { container } = render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    const skeletons = container.querySelectorAll('[data-slot="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it("PAGE_SIZE만큼 결과가 오면 더 불러오기 버튼을 표시한다", async () => {
    const fullPage = Array.from({ length: 100 }, (_, i) => ({
      ...MOCK_ROWS[0],
      date: `2026-01-${String(i + 1).padStart(2, "0")}`,
    }))

    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: fullPage }), { status: 200 }),
    )

    render(<PriceDataTable defaultFilter={DEFAULT_FILTER} />)

    await waitFor(() => {
      expect(screen.getByText("더 불러오기")).toBeInTheDocument()
    })
  })
})
