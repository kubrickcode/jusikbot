"use client"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { Currency, Holdings, Position } from "@/lib/types/holdings"
import type { WatchlistItem } from "@/lib/types/watchlist"
import { AlertTriangle, Loader2, Plus, Trash2 } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"

type HoldingsEditorProps = {
  initialHoldings: Holdings
  watchlistItems: WatchlistItem[]
}

type SaveState = "idle" | "saving" | "success" | "error"

export function HoldingsEditor({
  initialHoldings,
  watchlistItems,
}: HoldingsEditorProps) {
  const [asOf, setAsOf] = useState(initialHoldings.as_of)
  const [positions, setPositions] = useState<Record<string, Position>>(
    initialHoldings.positions,
  )
  const [saveState, setSaveState] = useState<SaveState>("idle")
  const [saveError, setSaveError] = useState("")
  const [addSymbol, setAddSymbol] = useState("")

  const watchlistMap = useMemo(() => {
    const map = new Map<string, WatchlistItem>()
    for (const item of watchlistItems) {
      map.set(item.symbol, item)
    }
    return map
  }, [watchlistItems])

  const isDirty =
    asOf !== initialHoldings.as_of ||
    JSON.stringify(positions) !== JSON.stringify(initialHoldings.positions)

  const orphanSymbols = useMemo(
    () => Object.keys(positions).filter((s) => !watchlistMap.has(s)),
    [positions, watchlistMap],
  )

  const availableSymbols = useMemo(
    () => watchlistItems.filter((item) => !(item.symbol in positions)),
    [watchlistItems, positions],
  )

  useEffect(() => {
    if (!isDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isDirty])

  const updatePosition = useCallback(
    (symbol: string, field: keyof Position, value: number | string) => {
      setPositions((prev) => ({
        ...prev,
        [symbol]: { ...prev[symbol], [field]: value },
      }))
    },
    [],
  )

  const addPosition = useCallback(() => {
    if (!addSymbol) return
    const watchItem = watchlistMap.get(addSymbol)
    const currency: Currency = watchItem?.market === "KR" ? "KRW" : "USD"
    setPositions((prev) => ({
      ...prev,
      [addSymbol]: { avg_cost: 0, currency, quantity: 0 },
    }))
    setAddSymbol("")
  }, [addSymbol, watchlistMap])

  const deletePosition = useCallback((symbol: string) => {
    setPositions((prev) => {
      const next = { ...prev }
      delete next[symbol]
      return next
    })
  }, [])

  const resetAll = useCallback(() => {
    setAsOf(initialHoldings.as_of)
    setPositions(initialHoldings.positions)
  }, [initialHoldings])

  const submitChanges = useCallback(async () => {
    setSaveState("saving")
    setSaveError("")

    const payload: Holdings = {
      $schema: "./holdings.schema.json",
      as_of: asOf,
      positions,
    }

    try {
      const res = await fetch("/api/config/holdings", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
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
  }, [asOf, positions])

  const sortedSymbols = useMemo(
    () => Object.keys(positions).sort(),
    [positions],
  )

  return (
    <div className="space-y-6 p-6 pb-24">
      {saveState === "error" && saveError && (
        <Alert variant="destructive">
          <AlertDescription className="whitespace-pre-wrap">{saveError}</AlertDescription>
        </Alert>
      )}

      {orphanSymbols.length > 0 && (
        <Alert>
          <AlertTriangle className="size-4" />
          <AlertDescription>
            다음 심볼이 watchlist에 등록되지 않았습니다: {orphanSymbols.join(", ")}
          </AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>기준일</CardTitle>
          <CardDescription>보유 현황의 최종 갱신 일자</CardDescription>
        </CardHeader>
        <CardContent>
          <Input
            type="date"
            value={asOf}
            onChange={(e) => setAsOf(e.target.value)}
            className="w-48"
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <CardTitle>포지션 ({sortedSymbols.length})</CardTitle>
          {availableSymbols.length > 0 && (
            <div className="flex items-center gap-2">
              <Select value={addSymbol} onValueChange={setAddSymbol}>
                <SelectTrigger className="w-[200px]">
                  <SelectValue placeholder="종목 선택" />
                </SelectTrigger>
                <SelectContent>
                  {availableSymbols.map((item) => (
                    <SelectItem key={item.symbol} value={item.symbol}>
                      {item.symbol} ({item.name})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button size="sm" onClick={addPosition} disabled={!addSymbol}>
                <Plus className="size-4 mr-1" />
                추가
              </Button>
            </div>
          )}
        </CardHeader>
        <CardContent>
          {sortedSymbols.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">
              보유 포지션이 없습니다. 종목을 추가해주세요.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>심볼</TableHead>
                  <TableHead>종목명</TableHead>
                  <TableHead>수량</TableHead>
                  <TableHead>평균단가</TableHead>
                  <TableHead>통화</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {sortedSymbols.map((symbol) => {
                  const pos = positions[symbol]
                  const watchItem = watchlistMap.get(symbol)
                  const isOrphan = !watchItem

                  return (
                    <TableRow key={symbol}>
                      <TableCell className="font-mono font-medium">
                        {symbol}
                      </TableCell>
                      <TableCell>
                        {isOrphan ? (
                          <span className="flex items-center gap-1 text-amber-600">
                            <AlertTriangle className="size-3.5" />
                            미등록
                          </span>
                        ) : (
                          watchItem.name
                        )}
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="any"
                          value={pos.quantity}
                          onChange={(e) =>
                            updatePosition(symbol, "quantity", parseFloat(e.target.value) || 0)
                          }
                          className="w-32 h-8"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="any"
                          value={pos.avg_cost}
                          onChange={(e) =>
                            updatePosition(symbol, "avg_cost", parseFloat(e.target.value) || 0)
                          }
                          className="w-32 h-8"
                        />
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{pos.currency}</Badge>
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="size-7 text-destructive hover:text-destructive"
                          onClick={() => deletePosition(symbol)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {isDirty && (
        <div className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background/95 backdrop-blur-sm">
          <div className="flex items-center justify-end gap-3 px-6 py-3 max-w-4xl mx-auto">
            <span className="text-sm text-muted-foreground mr-auto">변경사항이 있습니다</span>
            <Button type="button" variant="outline" onClick={resetAll}>
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
