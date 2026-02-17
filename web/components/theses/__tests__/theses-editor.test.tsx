import { render, screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"
import { ThesesEditor } from "../theses-editor"

const mockContent = `# 투자 논제

## T0. 미국 기업 생태계의 자기 강화

**역할**: core
**시간 지평선**: 장기 (영구 보유)

| 리스크 | 영향 |
| ------ | ---- |
| AI 정체 | 전체 |
`

function renderEditor(content = mockContent) {
  return render(
    <div className="flex flex-col h-full">
      <ThesesEditor initialContent={content} />
    </div>,
  )
}

describe("ThesesEditor", () => {
  describe("렌더링", () => {
    it("초기 콘텐츠가 textarea에 표시된다", () => {
      renderEditor()

      const textarea = screen.getByTestId("theses-textarea") as HTMLTextAreaElement
      expect(textarea.value).toBe(mockContent)
    })

    it("분할 모드에서 편집 영역과 미리보기 영역이 모두 표시된다", () => {
      renderEditor()

      expect(screen.getByTestId("theses-textarea")).toBeInTheDocument()
      expect(screen.getByTestId("preview-pane")).toBeInTheDocument()
    })

    it("글자 수가 표시된다", () => {
      renderEditor()

      const charCount = screen.getByTestId("char-count")
      expect(charCount.textContent).toContain("글자")
    })

    it("초기 상태에서 저장 바가 숨겨져 있다", () => {
      renderEditor()

      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })

    it("미리보기에 마크다운이 렌더링된다", () => {
      renderEditor()

      const preview = within(screen.getByTestId("preview-pane"))
      expect(preview.getByText("투자 논제")).toBeInTheDocument()
      expect(preview.getByText(/미국 기업 생태계/)).toBeInTheDocument()
    })
  })

  describe("뷰 모드 전환", () => {
    it("편집 모드에서 textarea만 표시된다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByTitle("편집"))

      expect(screen.getByTestId("theses-textarea")).toBeInTheDocument()
      expect(screen.queryByTestId("preview-pane")).not.toBeInTheDocument()
    })

    it("미리보기 모드에서 미리보기만 표시된다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByTitle("미리보기"))

      expect(screen.queryByTestId("theses-textarea")).not.toBeInTheDocument()
      expect(screen.getByTestId("preview-pane")).toBeInTheDocument()
    })

    it("분할 모드에서 양쪽 모두 표시된다", async () => {
      const user = userEvent.setup()
      renderEditor()

      await user.click(screen.getByTitle("편집"))
      await user.click(screen.getByTitle("분할"))

      expect(screen.getByTestId("theses-textarea")).toBeInTheDocument()
      expect(screen.getByTestId("preview-pane")).toBeInTheDocument()
    })
  })

  describe("편집과 저장", () => {
    it("텍스트를 변경하면 저장 바가 나타난다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const textarea = screen.getByTestId("theses-textarea")
      await user.click(textarea)
      await user.type(textarea, " 추가 텍스트")

      expect(screen.getByText("변경사항이 있습니다")).toBeInTheDocument()
    })

    it("초기화 버튼이 원래 내용으로 복원한다", async () => {
      const user = userEvent.setup()
      renderEditor()

      const textarea = screen.getByTestId("theses-textarea") as HTMLTextAreaElement
      await user.click(textarea)
      await user.type(textarea, " 추가")

      await user.click(screen.getByText("초기화"))

      expect(textarea.value).toBe(mockContent)
      expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
    })

    it("저장 버튼이 PUT /api/config/theses를 호출한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderEditor()

      const textarea = screen.getByTestId("theses-textarea")
      await user.click(textarea)
      await user.type(textarea, " 수정됨")

      await user.click(screen.getByRole("button", { name: "저장" }))

      await waitFor(() => {
        expect(fetchMock).toHaveBeenCalledWith(
          "/api/config/theses",
          expect.objectContaining({
            method: "PUT",
            headers: { "Content-Type": "application/json" },
          }),
        )
      })

      const callBody = JSON.parse(fetchMock.mock.calls[0][1].body)
      expect(callBody.content).toContain("수정됨")

      vi.unstubAllGlobals()
    })

    it("API 에러 시 에러 메시지를 표시한다", async () => {
      const user = userEvent.setup()
      const fetchMock = vi.fn().mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ error: "내용이 비어있습니다" }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderEditor()

      const textarea = screen.getByTestId("theses-textarea")
      await user.click(textarea)
      await user.type(textarea, " 에러 테스트")

      await user.click(screen.getByRole("button", { name: "저장" }))

      await waitFor(() => {
        expect(screen.getByText("내용이 비어있습니다")).toBeInTheDocument()
      })

      vi.unstubAllGlobals()
    })

    it("저장 성공 후 dirty 상태가 해제된다", async () => {
      const user = userEvent.setup()
      const modifiedContent = mockContent + " 수정됨"
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true, content: modifiedContent }),
      })
      vi.stubGlobal("fetch", fetchMock)

      renderEditor()

      const textarea = screen.getByTestId("theses-textarea")
      await user.click(textarea)
      await user.type(textarea, " 수정됨")

      await user.click(screen.getByRole("button", { name: "저장" }))

      await waitFor(() => {
        expect(screen.queryByText("변경사항이 있습니다")).not.toBeInTheDocument()
      })

      vi.unstubAllGlobals()
    })
  })
})
