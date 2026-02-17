import { render, screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"
import { HoldingsEditor } from "../holdings-editor"
import type { Holdings } from "@/lib/types/holdings"
import type { WatchlistItem } from "@/lib/types/watchlist"

const mockHoldings: Holdings = {
  as_of: "2026-02-16",
  positions: {
    NVDA: { avg_cost: 172.7, currency: "USD", quantity: 1.2656548 },
    "133690": { avg_cost: 120290, currency: "KRW", quantity: 14 },
  },
}

const mockWatchlist: WatchlistItem[] = [
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
    name: "TIGER 미국나스닥100",
    sector: "us-broad-market",
    symbol: "133690",
    themes: ["nasdaq-exposure"],
    type: "etf",
  },
  {
    market: "US",
    name: "Apple",
    sector: "big-tech",
    symbol: "AAPL",
    themes: ["consumer-tech"],
    type: "stock",
  },
]

function renderEditor(overrides?: {
  holdings?: Holdings
  watchlist?: WatchlistItem[]
}) {
  return render(
    <HoldingsEditor
      initialHoldings={overrides?.holdings ?? mockHoldings}
      watchlistItems={overrides?.watchlist ?? mockWatchlist}
    />,
  )
}

describe("HoldingsEditor", () => {
  describe("렌더링", () => {
    it("모든 포지션이 테이블에 표시된다", () => {
      renderEditor()

      expect(screen.getByText("NVDA")).toBeInTheDocument()
      expect(screen.getByText("NVIDIA")).toBeInTheDocument()
      expect(screen.getByText("133690")).toBeInTheDocument()
    })

    it("기준일이 표시된다", () => {
      renderEditor()

      const dateInput = screen.getByDisplayValue("2026-02-16")
      expect(dateInput).toBeInTheDocument()
    })

    it("포지션 수가 카드 제목에 표시된다", () => {
      renderEditor()

      expect(screen.getByText(/포지션 \(2\)/)).toBeInTheDocument()
    })

    it("초기 상태에서 저장 바가 숨겨져 있다", () => {
      renderEditor()

      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })
  })

  describe("watchlist 미등록 경고", () => {
    it("watchlist에 없는 심볼이 있으면 경고를 표시한다", () => {
      const orphanHoldings: Holdings = {
        as_of: "2026-02-16",
        positions: {
          NVDA: { avg_cost: 172.7, currency: "USD", quantity: 1 },
          TSLA: { avg_cost: 250.0, currency: "USD", quantity: 2 },
        },
      }
      renderEditor({ holdings: orphanHoldings })

      const alert = screen.getByText(/watchlist에 등록되지 않았습니다/)
      expect(alert).toBeInTheDocument()
      expect(alert.textContent).toContain("TSLA")
    })

    it("미등록 포지션 행에 경고 아이콘을 표시한다", () => {
      const orphanHoldings: Holdings = {
        as_of: "2026-02-16",
        positions: {
          TSLA: { avg_cost: 250.0, currency: "USD", quantity: 2 },
        },
      }
      renderEditor({ holdings: orphanHoldings })

      expect(screen.getByText("미등록")).toBeInTheDocument()
    })
  })

  describe("포지션 수정", () => {
    it("수량을 변경하면 저장 바가 나타난다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const quantityInputs = screen.getAllByDisplayValue("1.2656548")
      await user.clear(quantityInputs[0])
      await user.type(quantityInputs[0], "2")

      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })

    it("기준일을 변경하면 저장 바가 나타난다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const dateInput = screen.getByDisplayValue("2026-02-16")
      await user.clear(dateInput)
      await user.type(dateInput, "2026-02-17")

      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })
  })

  describe("포지션 삭제", () => {
    it("삭제 버튼으로 포지션을 제거한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const nvdaRow = screen.getByText("NVDA").closest("tr")!
      const deleteButton = within(nvdaRow).getAllByRole("button")[0]
      await user.click(deleteButton)

      expect(screen.queryByText("NVIDIA")).not.toBeInTheDocument()
      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })
  })

  describe("저장", () => {
    it("저장 버튼이 PUT /api/config/holdings를 호출한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderEditor()

      const dateInput = screen.getByDisplayValue("2026-02-16")
      await user.clear(dateInput)
      await user.type(dateInput, "2026-02-17")

      await user.click(screen.getByRole("button", { name: "저장" }))

      await waitFor(() => {
        expect(fetchMock).toHaveBeenCalledWith(
          "/api/config/holdings",
          expect.objectContaining({ method: "PUT" }),
        )
      })

      const body = JSON.parse(fetchMock.mock.calls[0][1].body)
      expect(body.as_of).toBe("2026-02-17")
      expect(body.$schema).toBe("./holdings.schema.json")
      expect(body.positions.NVDA).toBeTruthy()

      vi.unstubAllGlobals()
    })

    it("초기화 버튼이 원래 상태로 복원한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const dateInput = screen.getByDisplayValue("2026-02-16") as HTMLInputElement
      await user.clear(dateInput)
      await user.type(dateInput, "2026-02-17")

      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()

      await user.click(screen.getByRole("button", { name: "초기화" }))

      expect(screen.getByDisplayValue("2026-02-16")).toBeInTheDocument()
      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })
  })
})
