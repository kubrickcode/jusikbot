import { PageHeader } from "@/components/page-header"
import { HoldingsEditor } from "@/components/holdings/holdings-editor"
import { configPaths } from "@/lib/paths"
import type { Holdings } from "@/lib/types/holdings"
import type { WatchlistItem } from "@/lib/types/watchlist"
import { readFile } from "fs/promises"

async function loadHoldings(): Promise<Holdings> {
  const raw = await readFile(configPaths.holdings, "utf-8")
  return JSON.parse(raw) as Holdings
}

async function loadWatchlist(): Promise<WatchlistItem[]> {
  const raw = await readFile(configPaths.watchlist, "utf-8")
  return JSON.parse(raw) as WatchlistItem[]
}

export const dynamic = "force-dynamic"

export default async function HoldingsPage() {
  const [holdings, watchlistItems] = await Promise.all([
    loadHoldings(),
    loadWatchlist(),
  ])

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="보유 현황"
        description="현재 포지션의 수량과 평균 단가"
      />
      <div className="flex-1 overflow-auto">
        <div className="max-w-5xl mx-auto">
          <HoldingsEditor
            initialHoldings={holdings}
            watchlistItems={watchlistItems}
          />
        </div>
      </div>
    </div>
  )
}
