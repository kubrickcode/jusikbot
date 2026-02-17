"use client"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { daysAgo, today } from "@/lib/format"
import { ChevronDown } from "lucide-react"
import { useCallback, useState } from "react"

type WatchlistSymbol = {
  name: string
  symbol: string
}

export type PriceFilterValues = {
  from: string
  symbols: string[]
  to: string
}

type PriceFilterBarProps = {
  onFilterChange: (filter: PriceFilterValues) => void
  watchlistSymbols: WatchlistSymbol[]
}

export function PriceFilterBar({ watchlistSymbols, onFilterChange }: PriceFilterBarProps) {
  const [selectedSymbols, setSelectedSymbols] = useState<string[]>(
    watchlistSymbols.map((w) => w.symbol),
  )
  const [from, setFrom] = useState(daysAgo(30))
  const [to, setTo] = useState(today())

  const applyFilter = useCallback(
    (symbols: string[], fromDate: string, toDate: string) => {
      onFilterChange({ symbols, from: fromDate, to: toDate })
    },
    [onFilterChange],
  )

  const toggleSymbol = useCallback(
    (symbol: string) => {
      setSelectedSymbols((prev) => {
        const next = prev.includes(symbol)
          ? prev.filter((s) => s !== symbol)
          : [...prev, symbol]
        applyFilter(next, from, to)
        return next
      })
    },
    [applyFilter, from, to],
  )

  const toggleAll = useCallback(() => {
    setSelectedSymbols((prev) => {
      const next =
        prev.length === watchlistSymbols.length
          ? []
          : watchlistSymbols.map((w) => w.symbol)
      applyFilter(next, from, to)
      return next
    })
  }, [applyFilter, from, to, watchlistSymbols])

  const selectedLabel =
    selectedSymbols.length === 0
      ? "종목 선택"
      : selectedSymbols.length === watchlistSymbols.length
        ? "전체 선택됨"
        : `${selectedSymbols.length}개 선택됨`

  return (
    <div className="flex items-center gap-3 px-4 py-3 border-b">
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" className="min-w-[140px] justify-between">
            {selectedLabel}
            <ChevronDown className="size-4 ml-1 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-56 p-2" align="start">
          <div className="flex items-center gap-2 px-2 py-1.5 border-b mb-1">
            <Checkbox
              id="select-all"
              checked={selectedSymbols.length === watchlistSymbols.length}
              onCheckedChange={toggleAll}
            />
            <Label htmlFor="select-all" className="text-xs cursor-pointer">
              전체 선택
            </Label>
          </div>
          {watchlistSymbols.map((item) => (
            <div
              key={item.symbol}
              className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-accent cursor-pointer"
              onClick={() => toggleSymbol(item.symbol)}
            >
              <Checkbox
                checked={selectedSymbols.includes(item.symbol)}
                onCheckedChange={() => toggleSymbol(item.symbol)}
              />
              <span className="text-xs font-mono">{item.symbol}</span>
              <span className="text-xs text-muted-foreground truncate">{item.name}</span>
            </div>
          ))}
        </PopoverContent>
      </Popover>

      <div className="flex items-center gap-2">
        <Input
          type="date"
          value={from}
          onChange={(e) => {
            setFrom(e.target.value)
            applyFilter(selectedSymbols, e.target.value, to)
          }}
          className="w-[150px] text-xs"
        />
        <span className="text-xs text-muted-foreground">~</span>
        <Input
          type="date"
          value={to}
          onChange={(e) => {
            setTo(e.target.value)
            applyFilter(selectedSymbols, from, e.target.value)
          }}
          className="w-[150px] text-xs"
        />
      </div>
    </div>
  )
}
