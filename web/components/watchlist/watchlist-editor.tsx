"use client"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { WatchlistItem } from "@/lib/types/watchlist"
import { Loader2, Pencil, Plus, Trash2 } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { WatchlistDialog } from "./watchlist-dialog"

type WatchlistEditorProps = {
  initialItems: WatchlistItem[]
  holdingsSymbols: string[]
}

type SaveState = "idle" | "saving" | "success" | "error"

export function WatchlistEditor({
  initialItems,
  holdingsSymbols,
}: WatchlistEditorProps) {
  const [items, setItems] = useState<WatchlistItem[]>(initialItems)
  const [saveState, setSaveState] = useState<SaveState>("idle")
  const [saveError, setSaveError] = useState("")
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<WatchlistItem | null>(null)

  const isDirty = JSON.stringify(items) !== JSON.stringify(initialItems)

  useEffect(() => {
    if (!isDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isDirty])

  const openAddDialog = useCallback(() => {
    setEditingItem(null)
    setIsDialogOpen(true)
  }, [])

  const openEditDialog = useCallback((item: WatchlistItem) => {
    setEditingItem(item)
    setIsDialogOpen(true)
  }, [])

  const saveItem = useCallback(
    (saved: WatchlistItem) => {
      if (editingItem) {
        setItems((prev) =>
          prev.map((it) => (it.symbol === saved.symbol ? saved : it)),
        )
      } else {
        setItems((prev) => [...prev, saved])
      }
    },
    [editingItem],
  )

  const deleteItem = useCallback((symbol: string) => {
    setItems((prev) => prev.filter((it) => it.symbol !== symbol))
  }, [])

  const resetItems = useCallback(() => {
    setItems(initialItems)
  }, [initialItems])

  const submitChanges = useCallback(async () => {
    setSaveState("saving")
    setSaveError("")

    try {
      const res = await fetch("/api/config/watchlist", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(items),
      })

      if (!res.ok) {
        const body = await res.json()
        if (body.errors) {
          const messages = body.errors
            .map((e: { path: string; message: string }) => `${e.path}: ${e.message}`)
            .join("\n")
          throw new Error(messages)
        }
        throw new Error(body.error || "저장 실패")
      }

      setSaveState("success")
      setTimeout(() => setSaveState("idle"), 1500)
    } catch (err) {
      setSaveState("error")
      setSaveError(err instanceof Error ? err.message : "알 수 없는 오류")
    }
  }, [items])

  const existingSymbols = items.map((it) => it.symbol)

  return (
    <div className="space-y-6 p-6 pb-24">
      {saveState === "error" && saveError && (
        <Alert variant="destructive">
          <AlertDescription className="whitespace-pre-wrap">{saveError}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <CardTitle>추적 종목 ({items.length})</CardTitle>
          <Button size="sm" onClick={openAddDialog}>
            <Plus className="size-4 mr-1" />
            종목 추가
          </Button>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">
              추적 중인 종목이 없습니다. 종목을 추가해주세요.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>심볼</TableHead>
                  <TableHead>종목명</TableHead>
                  <TableHead>거래소</TableHead>
                  <TableHead>유형</TableHead>
                  <TableHead>섹터</TableHead>
                  <TableHead>테마</TableHead>
                  <TableHead className="w-[80px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.symbol}>
                    <TableCell className="font-mono font-medium">{item.symbol}</TableCell>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{item.market}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">
                        {item.type === "etf" ? "ETF" : "주식"}
                      </Badge>
                    </TableCell>
                    <TableCell>{item.sector}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {item.themes.map((theme) => (
                          <Badge key={theme} variant="outline" className="text-xs">
                            {theme}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7"
                          onClick={() => openEditDialog(item)}
                        >
                          <Pencil className="size-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7 text-destructive hover:text-destructive"
                          onClick={() => deleteItem(item.symbol)}
                          disabled={holdingsSymbols.includes(item.symbol)}
                          title={
                            holdingsSymbols.includes(item.symbol)
                              ? "보유 중인 종목은 삭제할 수 없습니다"
                              : "삭제"
                          }
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <WatchlistDialog
        isOpen={isDialogOpen}
        onClose={() => setIsDialogOpen(false)}
        onSave={saveItem}
        editingItem={editingItem}
        existingSymbols={existingSymbols}
      />

      {isDirty && (
        <div className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background/95 backdrop-blur-sm">
          <div className="flex items-center justify-end gap-3 px-6 py-3 max-w-4xl mx-auto">
            <span className="text-sm text-muted-foreground mr-auto">변경사항이 있습니다</span>
            <Button type="button" variant="outline" onClick={resetItems}>
              초기화
            </Button>
            <Button onClick={submitChanges} disabled={saveState === "saving"}>
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
    </div>
  )
}
