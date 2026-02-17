import { render, screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"
import { WatchlistEditor } from "../watchlist-editor"
import type { WatchlistItem } from "@/lib/types/watchlist"

const mockItems: WatchlistItem[] = [
  {
    market: "US",
    name: "NVIDIA",
    sector: "semiconductor",
    symbol: "NVDA",
    themes: ["AI-infra"],
    type: "stock",
  },
  {
    market: "KR",
    name: "TIGER 미국S&P500",
    sector: "us-broad-market",
    symbol: "360750",
    themes: ["sp500-core"],
    type: "etf",
  },
]

const holdingsSymbols = ["NVDA"]

function renderEditor(overrides?: Partial<{ items: WatchlistItem[]; holdingsSymbols: string[] }>) {
  return render(
    <WatchlistEditor
      initialItems={overrides?.items ?? mockItems}
      holdingsSymbols={overrides?.holdingsSymbols ?? holdingsSymbols}
    />,
  )
}

describe("WatchlistEditor", () => {
  describe("렌더링", () => {
    it("모든 종목이 테이블에 표시된다", () => {
      renderEditor()

      expect(screen.getByText("NVDA")).toBeInTheDocument()
      expect(screen.getByText("NVIDIA")).toBeInTheDocument()
      expect(screen.getByText("360750")).toBeInTheDocument()
      expect(screen.getByText("TIGER 미국S&P500")).toBeInTheDocument()
    })

    it("종목 수가 카드 제목에 표시된다", () => {
      renderEditor()

      expect(screen.getByText(/추적 종목 \(2\)/)).toBeInTheDocument()
    })

    it("초기 상태에서 저장 바가 숨겨져 있다", () => {
      renderEditor()

      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })

    it("빈 목록이면 안내 메시지를 표시한다", () => {
      renderEditor({ items: [] })

      expect(screen.getByText(/추적 중인 종목이 없습니다/)).toBeInTheDocument()
    })
  })

  describe("종목 추가", () => {
    it("다이얼로그에서 새 종목을 추가하면 테이블에 반영된다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByText("종목 추가"))

      const dialog = screen.getByRole("dialog")
      await user.type(within(dialog).getByLabelText("심볼"), "AAPL")
      await user.type(within(dialog).getByLabelText("종목명"), "Apple")
      await user.type(within(dialog).getByLabelText("섹터"), "big-tech")

      await user.click(within(dialog).getByRole("button", { name: "추가" }))

      expect(screen.getByText("AAPL")).toBeInTheDocument()
      expect(screen.getByText("Apple")).toBeInTheDocument()
      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })

    it("중복 심볼 추가 시 에러를 표시한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByText("종목 추가"))

      const dialog = screen.getByRole("dialog")
      await user.type(within(dialog).getByLabelText("심볼"), "NVDA")
      await user.type(within(dialog).getByLabelText("종목명"), "NVIDIA")
      await user.type(within(dialog).getByLabelText("섹터"), "semiconductor")

      await user.click(within(dialog).getByRole("button", { name: "추가" }))

      expect(within(dialog).getByText("이미 존재하는 심볼입니다")).toBeInTheDocument()
    })

    it("필수 필드 누락 시 에러를 표시한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByText("종목 추가"))

      const dialog = screen.getByRole("dialog")
      await user.click(within(dialog).getByRole("button", { name: "추가" }))

      expect(within(dialog).getByText("심볼은 필수입니다")).toBeInTheDocument()
    })
  })

  describe("종목 삭제", () => {
    it("보유하지 않은 종목을 삭제하면 테이블에서 제거된다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const row = screen.getByText("360750").closest("tr")!
      const deleteButton = within(row).getByRole("button", { name: "" })
      const deleteButtons = within(row).getAllByRole("button")
      const trashButton = deleteButtons.find((btn) =>
        btn.querySelector('[class*="lucide-trash"]') !== null ||
        !btn.querySelector('[class*="lucide-pencil"]'),
      )

      await user.click(deleteButtons[1])

      expect(screen.queryByText("360750")).not.toBeInTheDocument()
      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })

    it("보유 중인 종목의 삭제 버튼은 비활성화된다", () => {
      renderEditor()

      const row = screen.getByText("NVDA").closest("tr")!
      const buttons = within(row).getAllByRole("button")
      const deleteButton = buttons[buttons.length - 1]

      expect(deleteButton).toBeDisabled()
    })
  })

  describe("저장", () => {
    it("저장 버튼이 PUT /api/config/watchlist를 호출한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderEditor()

      await user.click(screen.getByText("종목 추가"))

      const dialog = screen.getByRole("dialog")
      await user.type(within(dialog).getByLabelText("심볼"), "AAPL")
      await user.type(within(dialog).getByLabelText("종목명"), "Apple")
      await user.type(within(dialog).getByLabelText("섹터"), "big-tech")
      await user.click(within(dialog).getByRole("button", { name: "추가" }))

      await user.click(screen.getByRole("button", { name: "저장" }))

      await waitFor(() => {
        expect(fetchMock).toHaveBeenCalledWith(
          "/api/config/watchlist",
          expect.objectContaining({ method: "PUT" }),
        )
      })

      const body = JSON.parse(fetchMock.mock.calls[0][1].body)
      expect(body).toHaveLength(3)
      expect(body.find((i: WatchlistItem) => i.symbol === "AAPL")).toBeTruthy()

      vi.unstubAllGlobals()
    })

    it("초기화 버튼이 원래 상태로 복원한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByText("종목 추가"))

      const dialog = screen.getByRole("dialog")
      await user.type(within(dialog).getByLabelText("심볼"), "AAPL")
      await user.type(within(dialog).getByLabelText("종목명"), "Apple")
      await user.type(within(dialog).getByLabelText("섹터"), "big-tech")
      await user.click(within(dialog).getByRole("button", { name: "추가" }))

      expect(screen.getByText("AAPL")).toBeInTheDocument()

      await user.click(screen.getByRole("button", { name: "초기화" }))

      expect(screen.queryByText("AAPL")).not.toBeInTheDocument()
      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })
  })
})
