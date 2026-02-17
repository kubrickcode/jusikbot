import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { ThesisCheckViewer } from "../thesis-check-viewer"

const MOCK_THESIS_CHECK = {
  checked_at: "2026-02-17",
  theses: [
    {
      name: "AI 인프라 확장",
      status: "valid",
      previous_status: null,
      status_transition: "new",
      upstream_dependency: null,
      chain_impact: null,
      conditions: [
        {
          text: "데이터센터 투자 증가세 지속",
          type: "validity",
          status: "met",
          previous_status: null,
          status_transition: "new",
          evidence: "2025년 하이퍼스케일러 CapEx 전년 대비 40% 증가",
          sources: [
            { title: "AWS 실적 발표", url: "https://example.com/aws", tier: 1, date: "2026-01-28" },
          ],
          quantitative_distance: "+40% vs 20% threshold",
        },
        {
          text: "GPU 공급 병목 지속",
          type: "validity",
          status: "partially_met",
          previous_status: null,
          status_transition: "new",
          evidence: "NVIDIA H200 대기 리스트 존재하나 단축 추세",
          sources: [
            { title: "Reuters 보도", url: "https://example.com/reuters", tier: 3, date: "2026-02-10" },
          ],
          quantitative_distance: null,
        },
        {
          text: "GPU 과잉 공급 전환",
          type: "invalidation",
          status: "not_yet",
          previous_status: null,
          status_transition: "new",
          evidence: "현재 과잉 공급 신호 미감지",
          sources: [
            { title: "IDC 리포트", url: "https://example.com/idc", tier: 2, date: "2026-01-15" },
          ],
          quantitative_distance: null,
        },
      ],
    },
    {
      name: "원화 약세",
      status: "weakening",
      previous_status: "valid",
      status_transition: "degrading",
      upstream_dependency: null,
      chain_impact: null,
      conditions: [
        {
          text: "USD/KRW 1400 이상 유지",
          type: "validity",
          status: "met",
          previous_status: "met",
          status_transition: "stable",
          evidence: "2026-02-17 기준 1445원 수준",
          sources: [
            { title: "한국은행", url: "https://example.com/bok", tier: 1, date: "2026-02-17" },
          ],
          quantitative_distance: null,
        },
        {
          text: "한국 금리 인하 기대감 약화",
          type: "invalidation",
          status: "partially_met",
          previous_status: null,
          status_transition: "new",
          evidence: "한국은행 기준금리 동결 결정",
          sources: [
            { title: "매일경제", url: "https://example.com/mk", tier: 3, date: "2026-02-14" },
          ],
          quantitative_distance: null,
        },
      ],
    },
  ],
}

beforeEach(() => {
  vi.restoreAllMocks()
})

describe("ThesisCheckViewer", () => {
  it("논제 카드를 렌더링한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("AI 인프라 확장")).toBeInTheDocument()
    })
    expect(screen.getByText("원화 약세")).toBeInTheDocument()
  })

  it("논제 상태 배지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("유효")).toBeInTheDocument()
    })
    expect(screen.getByText("약화 중")).toBeInTheDocument()
  })

  it("확인일을 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("확인일: 2026-02-17")).toBeInTheDocument()
    })
  })

  it("조건 텍스트를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("데이터센터 투자 증가세 지속")).toBeInTheDocument()
    })
    expect(screen.getByText("GPU 공급 병목 지속")).toBeInTheDocument()
  })

  it("조건 상태 배지를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getAllByText("충족").length).toBeGreaterThan(0)
    })
    expect(screen.getAllByText("부분 충족").length).toBeGreaterThan(0)
    expect(screen.getAllByText("미확인").length).toBeGreaterThan(0)
  })

  it("조건 클릭 시 근거와 출처를 펼친다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("데이터센터 투자 증가세 지속")).toBeInTheDocument()
    })

    await user.click(screen.getByText("데이터센터 투자 증가세 지속"))

    await waitFor(() => {
      expect(screen.getByText("2025년 하이퍼스케일러 CapEx 전년 대비 40% 증가")).toBeInTheDocument()
    })
    expect(screen.getByText("AWS 실적 발표")).toBeInTheDocument()
  })

  it("유효 조건 수 카운트를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("유효 조건: 2/2 충족")).toBeInTheDocument()
    })
  })

  it("전이 표시를 렌더링한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("악화")).toBeInTheDocument()
    })
  })

  it("404 시 빈 상태를 표시한다", async () => {
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ exists: false }), { status: 404 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("리서치 결과 없음")).toBeInTheDocument()
    })
    expect(screen.getByText("/thesis-research 스킬을 먼저 실행하세요")).toBeInTheDocument()
  })

  it("로딩 중 스켈레톤을 표시한다", () => {
    vi.spyOn(global, "fetch").mockReturnValueOnce(new Promise(() => {}))

    render(<ThesisCheckViewer />)

    const skeletons = document.querySelectorAll('[data-slot="skeleton"]')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it("정량 거리를 표시한다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("데이터센터 투자 증가세 지속")).toBeInTheDocument()
    })

    await user.click(screen.getByText("데이터센터 투자 증가세 지속"))

    await waitFor(() => {
      expect(screen.getByText("+40% vs 20% threshold")).toBeInTheDocument()
    })
  })

  it("출처 티어 배지를 표시한다", async () => {
    const user = userEvent.setup()
    vi.spyOn(global, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(MOCK_THESIS_CHECK), { status: 200 }),
    )

    render(<ThesisCheckViewer />)

    await waitFor(() => {
      expect(screen.getByText("데이터센터 투자 증가세 지속")).toBeInTheDocument()
    })

    await user.click(screen.getByText("데이터센터 투자 증가세 지속"))

    await waitFor(() => {
      expect(screen.getByText("T1")).toBeInTheDocument()
    })
  })
})
