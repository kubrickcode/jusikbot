"use client"

import { useCallback, useRef, useState } from "react"
import { daysAgo, today } from "@/lib/format"
import { PriceFilterBar, type PriceFilterValues } from "./price-filter-bar"
import { PriceDataTable } from "./price-data-table"

type PricesViewProps = {
  watchlistSymbols: { name: string; symbol: string }[]
}

export function PricesView({ watchlistSymbols }: PricesViewProps) {
  const defaultFilter: PriceFilterValues = {
    from: daysAgo(30),
    symbols: watchlistSymbols.map((w) => w.symbol),
    to: today(),
  }

  const [filter, setFilter] = useState<PriceFilterValues>(defaultFilter)
  const keyRef = useRef(0)

  const handleFilterChange = useCallback((newFilter: PriceFilterValues) => {
    setFilter(newFilter)
    keyRef.current += 1
  }, [])

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <PriceFilterBar
        watchlistSymbols={watchlistSymbols}
        onFilterChange={handleFilterChange}
      />
      <div className="flex-1 overflow-auto">
        <PriceDataTable key={keyRef.current} defaultFilter={filter} />
      </div>
    </div>
  )
}
