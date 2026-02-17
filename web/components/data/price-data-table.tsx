"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatPrice, formatVolume } from "@/lib/format"
import type { PriceRow } from "@/lib/types/database"
import { Loader2, TriangleAlert } from "lucide-react"
import { useCallback, useEffect, useRef, useState } from "react"
import type { PriceFilterValues } from "./price-filter-bar"

const PAGE_SIZE = 100

type PriceDataTableProps = {
  defaultFilter: PriceFilterValues
}

export function PriceDataTable({ defaultFilter }: PriceDataTableProps) {
  const [rows, setRows] = useState<PriceRow[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isAppending, setIsAppending] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [error, setError] = useState("")
  const filterRef = useRef(defaultFilter)

  const fetchRows = useCallback(async (filter: PriceFilterValues, offset: number) => {
    const params = new URLSearchParams()
    if (filter.symbols.length > 0) params.set("symbols", filter.symbols.join(","))
    if (filter.from) params.set("from", filter.from)
    if (filter.to) params.set("to", filter.to)
    params.set("limit", String(PAGE_SIZE))
    params.set("offset", String(offset))
    params.set("sort", "desc")

    const res = await fetch(`/api/data/prices?${params}`)
    if (!res.ok) {
      const body = await res.json()
      throw new Error(body.error || "조회 실패")
    }
    const data = await res.json()
    return data.rows as PriceRow[]
  }, [])

  const loadInitial = useCallback(async (filter: PriceFilterValues) => {
    setIsLoading(true)
    setError("")
    try {
      const data = await fetchRows(filter, 0)
      setRows(data)
      setHasMore(data.length === PAGE_SIZE)
    } catch (err) {
      setError(err instanceof Error ? err.message : "조회 실패")
      setRows([])
    } finally {
      setIsLoading(false)
    }
  }, [fetchRows])

  const loadMore = useCallback(async () => {
    setIsAppending(true)
    try {
      const data = await fetchRows(filterRef.current, rows.length)
      setRows((prev) => [...prev, ...data])
      setHasMore(data.length === PAGE_SIZE)
    } catch (err) {
      setError(err instanceof Error ? err.message : "추가 조회 실패")
    } finally {
      setIsAppending(false)
    }
  }, [fetchRows, rows.length])

  useEffect(() => {
    loadInitial(defaultFilter)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps -- mount only

  const applyFilter = useCallback(
    (filter: PriceFilterValues) => {
      filterRef.current = filter
      loadInitial(filter)
    },
    [loadInitial],
  )

  // Expose applyFilter via ref pattern
  const applyFilterRef = useRef(applyFilter)
  applyFilterRef.current = applyFilter

  // Re-export for parent to use
  useEffect(() => {
    applyFilterRef.current = applyFilter
  }, [applyFilter])

  return (
    <div className="flex flex-col">
      {error && (
        <div className="px-4 py-2 text-sm text-destructive">{error}</div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>날짜</TableHead>
            <TableHead>종목</TableHead>
            <TableHead className="text-right">시가</TableHead>
            <TableHead className="text-right">고가</TableHead>
            <TableHead className="text-right">저가</TableHead>
            <TableHead className="text-right">종가</TableHead>
            <TableHead className="text-right">수정종가</TableHead>
            <TableHead className="text-right">거래량</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <SkeletonRows />
          ) : rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8} className="h-32 text-center text-muted-foreground">
                데이터 없음
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow
                key={`${row.symbol}-${row.date}`}
                className={row.isAnomaly ? "bg-yellow-500/8" : undefined}
              >
                <TableCell className="font-mono text-xs">{row.date}</TableCell>
                <TableCell className="font-mono text-xs">
                  {row.isAnomaly && (
                    <TriangleAlert className="size-3 inline mr-1 text-yellow-500" />
                  )}
                  {row.symbol}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatPrice(row.open, row.symbol)}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatPrice(row.high, row.symbol)}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatPrice(row.low, row.symbol)}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatPrice(row.close, row.symbol)}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatPrice(row.adjClose, row.symbol)}
                </TableCell>
                <TableCell className="text-right font-mono text-xs">
                  {formatVolume(row.volume)}
                </TableCell>
              </TableRow>
            ))
          )}
          {isAppending && <SkeletonRows count={3} />}
        </TableBody>
      </Table>

      {hasMore && !isLoading && (
        <div className="flex justify-center py-4">
          <Button variant="outline" size="sm" onClick={loadMore} disabled={isAppending}>
            {isAppending ? (
              <>
                <Loader2 className="size-4 mr-1 animate-spin" />
                불러오는 중...
              </>
            ) : (
              "더 불러오기"
            )}
          </Button>
        </div>
      )}
    </div>
  )
}

// Separate ref accessor for parent component usage
export function usePriceTableRef() {
  const ref = useRef<{ applyFilter: (filter: PriceFilterValues) => void } | null>(null)
  return ref
}

function SkeletonRows({ count = 8 }: { count?: number }) {
  return Array.from({ length: count }).map((_, i) => (
    <TableRow key={`skeleton-${i}`}>
      {Array.from({ length: 8 }).map((_, j) => (
        <TableCell key={`skeleton-${i}-${j}`}>
          <Skeleton className="h-4 w-16" />
        </TableCell>
      ))}
    </TableRow>
  ))
}
