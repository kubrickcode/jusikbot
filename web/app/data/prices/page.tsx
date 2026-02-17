import { readFile } from "fs/promises"
import { PageHeader } from "@/components/page-header"
import { PricesView } from "@/components/data/prices-view"
import { configPaths } from "@/lib/paths"
import type { WatchlistItem } from "@/lib/types/watchlist"

async function loadWatchlistSymbols(): Promise<{ name: string; symbol: string }[]> {
  const raw = await readFile(configPaths.watchlist, "utf-8")
  const watchlist = JSON.parse(raw) as WatchlistItem[]
  return watchlist.map(({ name, symbol }) => ({ name, symbol }))
}

export const dynamic = "force-dynamic"

export default async function PricesPage() {
  const watchlistSymbols = await loadWatchlistSymbols()

  return (
    <div className="flex flex-col h-full">
      <PageHeader title="가격 데이터" description="종목별 OHLCV 가격 이력" />
      <PricesView watchlistSymbols={watchlistSymbols} />
    </div>
  )
}
