import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { ReportViewer } from "../report-viewer"

const MOCK_ENTRIES = {
  entries: [
    { date: "2026-02-17", filename: "2026-02-17.md" },
    { date: "2026-02-10", filename: "2026-02-10.md" },
    { date: "2026-01-20", filename: "2026-01-20.md" },
  ],
}

const MOCK_REPORT = {
  date: "2026-02-17",
  content: "# 포트폴리오 분석\n\n## 요약\n\n현재 포지션은 안정적입니다.\n\n| 종목 | 비중 |\n|------|------|\n| META | 30% |\n| AAPL | 20% |",
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe("ReportViewer", () => {
  it("리포트 목록을 렌더링한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_ENTRIES), { status: 200 }),
    )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-17")).toBeInTheDocument()
    })
    expect(screen.getByText("2026-02-10")).toBeInTheDocument()
    expect(screen.getByText("2026-01-20")).toBeInTheDocument()
  })

  it("리포트 개수를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_ENTRIES), { status: 200 }),
    )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("3개 리포트")).toBeInTheDocument()
    })
  })

  it("리포트 클릭 시 상세 뷰를 표시한다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_ENTRIES), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_REPORT), { status: 200 }),
      )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-17")).toBeInTheDocument()
    })

    await user.click(screen.getByText("2026-02-17"))

    await waitFor(() => {
      expect(screen.getByText("포트폴리오 분석")).toBeInTheDocument()
    })
    expect(screen.getByText("목록")).toBeInTheDocument()
  })

  it("상세 뷰에서 목록으로 돌아간다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_ENTRIES), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_REPORT), { status: 200 }),
      )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-17")).toBeInTheDocument()
    })

    await user.click(screen.getByText("2026-02-17"))

    await waitFor(() => {
      expect(screen.getByText("목록")).toBeInTheDocument()
    })

    await user.click(screen.getByText("목록"))

    await waitFor(() => {
      expect(screen.getByText("3개 리포트")).toBeInTheDocument()
    })
  })

  it("빈 목록일 때 안내 메시지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ entries: [] }), { status: 200 }),
    )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("리포트 없음")).toBeInTheDocument()
    })
    expect(
      screen.getByText("/portfolio-analyze 스킬을 실행하면 분석 리포트가 생성됩니다"),
    ).toBeInTheDocument()
  })

  it("에러 시 에러 메시지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockRejectedValueOnce(new Error("네트워크 오류"))

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("조회 실패")).toBeInTheDocument()
    })
  })

  it("로딩 중 스켈레톤을 표시한다", () => {
    vi.spyOn(global, "fetch").mockReturnValueOnce(new Promise(() => {}))

    render(<ReportViewer />)

    const skeletons = document.querySelectorAll('[data-slot="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it("마크다운 테이블을 렌더링한다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_ENTRIES), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(MOCK_REPORT), { status: 200 }),
      )

    render(<ReportViewer />)

    await waitFor(() => {
      expect(screen.getByText("2026-02-17")).toBeInTheDocument()
    })

    await user.click(screen.getByText("2026-02-17"))

    await waitFor(() => {
      expect(screen.getByText("META")).toBeInTheDocument()
    })
    expect(screen.getByText("30%")).toBeInTheDocument()
  })
})
