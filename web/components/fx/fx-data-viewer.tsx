"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { daysAgo, formatRate, today } from "@/lib/format"
import type { FxRateRow } from "@/lib/types/database"
import { Loader2, Search } from "lucide-react"
import { useCallback, useEffect, useState } from "react"

type FetchState = "idle" | "loading" | "error"

const PAGE_SIZE = 30

export function FxDataViewer() {
  const [from, setFrom] = useState(daysAgo(90))
  const [to, setTo] = useState(today())
  const [rows, setRows] = useState<FxRateRow[]>([])
  const [fetchState, setFetchState] = useState<FetchState>("idle")
  const [error, setError] = useState("")
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)

  const fetchRates = useCallback(async (fromDate: string, toDate: string) => {
    setFetchState("loading")
    setError("")
    try {
      const params = new URLSearchParams({
        pair: "USDKRW",
        from: fromDate,
        to: toDate,
        limit: "1000",
      })
      const res = await fetch(`/api/data/fx?${params}`)
      if (!res.ok) {
        const body = await res.json()
        throw new Error(body.error || "조회 실패")
      }
      const data = await res.json()
      setRows(data.rows as FxRateRow[])
      setVisibleCount(PAGE_SIZE)
      setFetchState("idle")
    } catch (err) {
      setFetchState("error")
      setError(err instanceof Error ? err.message : "조회 실패")
      setRows([])
    }
  }, [])

  useEffect(() => {
    fetchRates(from, to)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps -- mount only

  const handleSearch = useCallback(() => {
    fetchRates(from, to)
  }, [fetchRates, from, to])

  // Reverse for display: newest first
  const displayRows = [...rows].reverse().slice(0, visibleCount)
  const hasMore = visibleCount < rows.length

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <div className="flex items-center gap-3 px-4 py-3 border-b">
        <Input
          type="date"
          value={from}
          onChange={(e) => setFrom(e.target.value)}
          className="w-[150px] text-xs"
        />
        <span className="text-xs text-muted-foreground">~</span>
        <Input
          type="date"
          value={to}
          onChange={(e) => setTo(e.target.value)}
          className="w-[150px] text-xs"
        />
        <Button variant="outline" size="sm" onClick={handleSearch}>
          <Search className="size-4 mr-1" />
          조회
        </Button>
      </div>

      {error && (
        <div className="px-4 py-2 text-sm text-destructive">{error}</div>
      )}

      <div className="flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>날짜</TableHead>
              <TableHead className="text-right">환율 (KRW)</TableHead>
              <TableHead className="text-right">전일 대비</TableHead>
              <TableHead>출처</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {fetchState === "loading" ? (
              <SkeletonRows />
            ) : displayRows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="h-32 text-center text-muted-foreground">
                  데이터 없음
                </TableCell>
              </TableRow>
            ) : (
              displayRows.map((row, displayIndex) => {
                // rows is ascending, displayRows is reversed
                const ascIndex = rows.length - 1 - displayIndex
                const prevRow = ascIndex > 0 ? rows[ascIndex - 1] : null
                const delta = prevRow ? row.rate - prevRow.rate : null

                return (
                  <TableRow key={row.date}>
                    <TableCell className="font-mono text-xs">{row.date}</TableCell>
                    <TableCell className="text-right font-mono text-xs">
                      {formatRate(row.rate)}
                    </TableCell>
                    <TableCell className="text-right">
                      {delta !== null && delta !== 0 && (
                        <RateChangeBadge delta={delta} />
                      )}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {row.source}
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>

        {hasMore && fetchState !== "loading" && (
          <div className="flex justify-center py-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setVisibleCount((c) => c + PAGE_SIZE)}
            >
              더 불러오기
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

function RateChangeBadge({ delta }: { delta: number }) {
  // Rate up = KRW weakens = red, Rate down = KRW strengthens = blue
  const isUp = delta > 0
  return (
    <Badge
      variant="outline"
      className={`text-xs font-mono ${
        isUp ? "text-red-400 border-red-400/30" : "text-blue-400 border-blue-400/30"
      }`}
    >
      {isUp ? "+" : ""}
      {delta.toFixed(2)}
    </Badge>
  )
}

function SkeletonRows() {
  return Array.from({ length: 8 }).map((_, i) => (
    <TableRow key={`skeleton-${i}`}>
      {Array.from({ length: 4 }).map((_, j) => (
        <TableCell key={`skeleton-${i}-${j}`}>
          <Skeleton className="h-4 w-20" />
        </TableCell>
      ))}
    </TableRow>
  ))
}
