import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"
import { SettingsForm } from "../settings-form"
import type { EtfOption, Settings } from "@/lib/types/settings"

const mockSettings: Settings = {
  adjustment_unit_krw: 100000,
  anchoring: {
    monthly_max_change_per_position_krw: 200000,
    monthly_max_total_change_krw: 500000,
    quarterly_max_change_per_position_krw: 400000,
    quarterly_max_total_change_krw: 1500000,
  },
  budget_krw: 5000000,
  holdings_staleness_threshold_days: 30,
  min_review_interval_days: 7,
  risk_tolerance: {
    max_drawdown_action_pct: 25,
    max_drawdown_warning_pct: 15,
    max_sector_concentration_pct: 60,
    max_single_etf_pct: 50,
    max_single_stock_pct: 30,
    min_position_size_krw: 100000,
  },
  sizing: {
    high_confidence_pool_pct: 50,
    low_confidence_pool_pct: 15,
    medium_confidence_pool_pct: 35,
  },
  strategy: {
    core_internal_ratio: { "360750": 1 },
    core_pct: 50,
    satellite_pct: 50,
  },
}

const mockEtfOptions: EtfOption[] = [
  { symbol: "360750", name: "TIGER 미국S&P500" },
  { symbol: "133690", name: "TIGER 미국나스닥100" },
]

function renderForm(overrides?: Partial<Settings>) {
  const settings = { ...mockSettings, ...overrides }
  return render(<SettingsForm initialValues={settings} etfOptions={mockEtfOptions} />)
}

describe("SettingsForm", () => {
  describe("렌더링", () => {
    it("모든 섹션 카드를 렌더링한다", () => {
      renderForm()

      expect(screen.getByText("투자 예산")).toBeInTheDocument()
      expect(screen.getByText("코어/위성 전략")).toBeInTheDocument()
      expect(screen.getByText("리스크 한도")).toBeInTheDocument()
      expect(screen.getByText("확신도별 배분")).toBeInTheDocument()
      expect(screen.getByText("변경 속도 제한")).toBeInTheDocument()
      expect(screen.getByText("운영 파라미터")).toBeInTheDocument()
    })

    it("초기값이 입력 필드에 표시된다", () => {
      renderForm()

      const budgetInput = screen.getByLabelText("총 예산") as HTMLInputElement
      expect(budgetInput.value).toBe("5000000")

      const coreInput = screen.getByLabelText("코어 비율") as HTMLInputElement
      expect(coreInput.value).toBe("50")
    })

    it("초기 상태에서 저장 바가 숨겨져 있다", () => {
      renderForm()

      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })
  })

  describe("값 변경", () => {
    it("값을 변경하면 저장 바가 나타난다", async () => {
      const user = userEvent.setup()
      renderForm()

      const budgetInput = screen.getByLabelText("총 예산")
      await user.clear(budgetInput)
      await user.type(budgetInput, "6000000")

      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })

    it("초기화 버튼이 원래 값으로 복원한다", async () => {
      const user = userEvent.setup()
      renderForm()

      const budgetInput = screen.getByLabelText("총 예산") as HTMLInputElement
      await user.clear(budgetInput)
      await user.type(budgetInput, "6000000")

      const resetButton = screen.getByText("초기화")
      await user.click(resetButton)

      expect(budgetInput.value).toBe("5000000")
    })
  })

  describe("합계 표시기", () => {
    it("전략/사이징 비율 합계가 100%로 표시된다", () => {
      renderForm()

      const indicators = screen.getAllByText(/합계: \d+%/)
      expect(indicators.length).toBeGreaterThanOrEqual(2)
      indicators.forEach((el) => {
        expect(el.textContent).toBe("합계: 100%")
      })
    })
  })

  describe("저장 호출", () => {
    it("저장 버튼이 PUT /api/config/settings를 호출한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderForm()

      const budgetInput = screen.getByLabelText("총 예산")
      await user.clear(budgetInput)
      await user.type(budgetInput, "6000000")

      const saveButtons = screen.getAllByRole("button", { name: /저장/ })
      await user.click(saveButtons[saveButtons.length - 1])

      await waitFor(() => {
        expect(fetchMock).toHaveBeenCalledWith("/api/config/settings", expect.objectContaining({
          method: "PUT",
          headers: { "Content-Type": "application/json" },
        }))
      })

      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body)
      expect(callBody.budget_krw).toBe(6000000)
      expect(callBody.$schema).toBe("./settings.schema.json")

      vi.unstubAllGlobals()
    })

    it("API 에러 시 에러 메시지를 표시한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: false,
        json: () =>
          Promise.resolve({
            errors: [{ path: "budget_krw", message: "must be >= 100000" }],
          }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderForm()

      const budgetInput = screen.getByLabelText("총 예산")
      await user.clear(budgetInput)
      await user.type(budgetInput, "100")

      const saveButtons = screen.getAllByRole("button", { name: /저장/ })
      await user.click(saveButtons[saveButtons.length - 1])

      await waitFor(() => {
        expect(screen.getByText(/budget_krw: must be >= 100000/)).toBeInTheDocument()
      })

      vi.unstubAllGlobals()
    })
  })

  describe("코어 ETF 편집기", () => {
    it("초기 ETF 항목과 추가 버튼을 렌더링한다", () => {
      renderForm()

      const addButtons = screen.getAllByRole("button", { name: /ETF 추가/ })
      expect(addButtons.length).toBeGreaterThanOrEqual(1)
    })
  })
})
