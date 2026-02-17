import { PageHeader } from "@/components/page-header"
import { WatchlistEditor } from "@/components/watchlist/watchlist-editor"
import { configPaths } from "@/lib/paths"
import type { Holdings } from "@/lib/types/holdings"
import type { WatchlistItem } from "@/lib/types/watchlist"
import { readFile } from "fs/promises"

async function loadWatchlist(): Promise<WatchlistItem[]> {
  const raw = await readFile(configPaths.watchlist, "utf-8")
  return JSON.parse(raw) as WatchlistItem[]
}

async function loadHoldingsSymbols(): Promise<string[]> {
  const raw = await readFile(configPaths.holdings, "utf-8")
  const holdings = JSON.parse(raw) as Holdings
  return Object.keys(holdings.positions)
}

export const dynamic = "force-dynamic"

export default async function WatchlistPage() {
  const [items, holdingsSymbols] = await Promise.all([
    loadWatchlist(),
    loadHoldingsSymbols(),
  ])

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="추적 종목"
        description="관심 종목 추가, 수정, 삭제"
      />
      <div className="flex-1 overflow-auto">
        <div className="max-w-5xl mx-auto">
          <WatchlistEditor initialItems={items} holdingsSymbols={holdingsSymbols} />
        </div>
      </div>
    </div>
  )
}
