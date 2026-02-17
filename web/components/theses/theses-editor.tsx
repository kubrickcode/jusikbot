"use client"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Loader2 } from "lucide-react"
import { useCallback, useEffect, useRef, useState } from "react"
import { ThesesPreview } from "./theses-preview"
import { type ViewMode, ViewModeToggle } from "./view-mode-toggle"

type ThesesEditorProps = {
  initialContent: string
}

type SaveState = "idle" | "saving" | "success" | "error"

const VIEW_MODE_STORAGE_KEY = "theses-view-mode"

function loadViewMode(): ViewMode {
  if (typeof window === "undefined") return "split"
  const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY)
  if (stored === "edit" || stored === "split" || stored === "preview") return stored
  return "split"
}

export function ThesesEditor({ initialContent }: ThesesEditorProps) {
  const [content, setContent] = useState(initialContent)
  const [savedContent, setSavedContent] = useState(initialContent)
  const [viewMode, setViewMode] = useState<ViewMode>("split")
  const [saveState, setSaveState] = useState<SaveState>("idle")
  const [saveError, setSaveError] = useState("")
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const isDirty = content !== savedContent

  useEffect(() => {
    setViewMode(loadViewMode())
  }, [])

  const changeViewMode = useCallback((mode: ViewMode) => {
    setViewMode(mode)
    localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode)
  }, [])

  useEffect(() => {
    if (!isDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isDirty])

  const saveDraft = useCallback(async () => {
    setSaveState("saving")
    setSaveError("")

    try {
      const res = await fetch("/api/config/theses", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content }),
      })

      if (!res.ok) {
        const body = await res.json()
        throw new Error(body.error || "저장 실패")
      }

      const body = await res.json()
      const formatted = body.content ?? content
      setContent(formatted)
      setSavedContent(formatted)
      setSaveState("success")
      setTimeout(() => setSaveState("idle"), 1500)
    } catch (err) {
      setSaveState("error")
      setSaveError(err instanceof Error ? err.message : "알 수 없는 오류")
    }
  }, [content])

  const resetDraft = useCallback(() => {
    setContent(savedContent)
    setSaveError("")
    setSaveState("idle")
  }, [savedContent])

  const insertTab = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key !== "Tab") return
    e.preventDefault()

    const textarea = e.currentTarget
    const { selectionStart, selectionEnd } = textarea
    const before = textarea.value.slice(0, selectionStart)
    const after = textarea.value.slice(selectionEnd)
    const updated = `${before}  ${after}`

    setContent(updated)

    requestAnimationFrame(() => {
      textarea.selectionStart = selectionStart + 2
      textarea.selectionEnd = selectionStart + 2
    })
  }, [])

  const charCount = content.length

  return (
    <>
      {/* Toolbar: view mode toggle + char count */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-border shrink-0">
        <span className="text-xs text-muted-foreground" data-testid="char-count">
          {charCount.toLocaleString()} 글자
        </span>
        <ViewModeToggle mode={viewMode} onChange={changeViewMode} />
      </div>

      {/* Error alert */}
      {saveState === "error" && saveError && (
        <div className="px-4 pt-3">
          <Alert variant="destructive">
            <AlertDescription>{saveError}</AlertDescription>
          </Alert>
        </div>
      )}

      {/* Editor area */}
      <div className="flex-1 min-h-0 flex">
        {/* Edit pane */}
        {viewMode !== "preview" && (
          <div
            className={cn(
              "flex flex-col min-h-0",
              viewMode === "split" ? "w-1/2 border-r border-border" : "w-full",
            )}
          >
            <textarea
              ref={textareaRef}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              onKeyDown={insertTab}
              className={cn(
                "w-full flex-1 resize-none",
                "bg-transparent text-foreground",
                "font-mono text-sm leading-relaxed",
                "p-4 outline-none",
                "border-0",
              )}
              spellCheck={false}
              data-testid="theses-textarea"
            />
          </div>
        )}

        {/* Preview pane */}
        {viewMode !== "edit" && (
          <div
            className={cn(
              "min-h-0 overflow-auto",
              viewMode === "split" ? "w-1/2" : "w-full",
            )}
            data-testid="preview-pane"
          >
            <ThesesPreview content={content} />
          </div>
        )}
      </div>

      {/* Sticky save bar */}
      {isDirty && (
        <div className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background/95 backdrop-blur-sm">
          <div className="flex items-center justify-end gap-3 px-6 py-3">
            <span className="text-sm text-muted-foreground mr-auto">변경사항이 있습니다</span>
            <Button
              type="button"
              variant="outline"
              onClick={resetDraft}
              disabled={saveState === "saving"}
            >
              초기화
            </Button>
            <Button type="button" onClick={saveDraft} disabled={saveState === "saving"}>
              {saveState === "saving" ? (
                <>
                  <Loader2 className="size-4 mr-1 animate-spin" />
                  저장 중...
                </>
              ) : saveState === "success" ? (
                "저장 완료"
              ) : (
                "저장"
              )}
            </Button>
          </div>
        </div>
      )}
    </>
  )
}
