import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { FxDataViewer } from "../fx-data-viewer"

const MOCK_FX_ROWS = [
  { date: "2026-02-12", pair: "USDKRW", rate: 1444.50, source: "frankfurter" },
  { date: "2026-02-13", pair: "USDKRW", rate: 1446.45, source: "frankfurter" },
  { date: "2026-02-14", pair: "USDKRW", rate: 1443.20, source: "frankfurter" },
]

beforeEach(() => {
  vi.restoreAllMocks()
})

describe("FxDataViewer", () => {
  it("테이블 헤더를 렌더링한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_FX_ROWS }), { status: 200 }),
    )

    render(<FxDataViewer />)

    expect(screen.getByText("날짜")).toBeInTheDocument()
    expect(screen.getByText("환율 (KRW)")).toBeInTheDocument()
    expect(screen.getByText("전일 대비")).toBeInTheDocument()
    expect(screen.getByText("출처")).toBeInTheDocument()
  })

  it("환율 데이터를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_FX_ROWS }), { status: 200 }),
    )

    render(<FxDataViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-14")).toBeInTheDocument()
    })

    expect(screen.getByText("2026-02-13")).toBeInTheDocument()
  })

  it("전일 대비 변동을 배지로 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_FX_ROWS }), { status: 200 }),
    )

    render(<FxDataViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-14")).toBeInTheDocument()
    })

    // 2026-02-13 to 2026-02-14: 1443.20 - 1446.45 = -3.25
    expect(screen.getByText("-3.25")).toBeInTheDocument()
    // 2026-02-12 to 2026-02-13: 1446.45 - 1444.50 = +1.95
    expect(screen.getByText("+1.95")).toBeInTheDocument()
  })

  it("데이터가 없으면 빈 상태를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: [] }), { status: 200 }),
    )

    render(<FxDataViewer />)

    await waitFor(() => {
      expect(screen.getByText("데이터 없음")).toBeInTheDocument()
    })
  })

  it("API 에러 시 에러 메시지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ error: "DB 연결 실패" }), { status: 500 }),
    )

    render(<FxDataViewer />)

    await waitFor(() => {
      expect(screen.getByText("DB 연결 실패")).toBeInTheDocument()
    })
  })

  it("조회 버튼 클릭 시 새로운 데이터를 가져온다", async () => {
    const user = userEvent.setup()
    const fetchSpy = vi.spyOn(global, "fetch")

    // Initial load
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: MOCK_FX_ROWS }), { status: 200 }),
    )

    render(<FxDataViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-14")).toBeInTheDocument()
    })

    // Click search button
    const newRows = [
      { date: "2026-01-15", pair: "USDKRW", rate: 1450.00, source: "frankfurter" },
    ]
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({ rows: newRows }), { status: 200 }),
    )

    await user.click(screen.getByText("조회"))

    await waitFor(() => {
      expect(screen.getByText("2026-01-15")).toBeInTheDocument()
    })
  })

  it("로딩 중 스켈레톤을 표시한다", () => {
    vi.spyOn(global, "fetch").mockReturnValueOnce(new Promise(() => {}))

    render(<FxDataViewer />)

    const skeletons = document.querySelectorAll('[data-slot="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })
})
