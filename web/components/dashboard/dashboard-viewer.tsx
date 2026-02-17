"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import {
  ArrowLeftRight,
  BookOpen,
  CandlestickChart,
  FileText,
  List,
  Wallet,
} from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { StalenessDot } from "./staleness-dot"
import { computeStaleness, formatDaysAgo } from "@/lib/staleness"
import { formatRate } from "@/lib/format"
import type { DashboardData } from "@/lib/types/dashboard"
import type { ThesisStatus } from "@/lib/types/research"

type FetchState = "idle" | "loading" | "error"

export function DashboardViewer() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [fetchState, setFetchState] = useState<FetchState>("loading")

  const fetchData = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch("/api/dashboard")
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setData(json as DashboardData)
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  if (fetchState === "loading") return <DashboardSkeleton />

  if (fetchState === "error") {
    return (
      <div className="p-4">
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground text-sm">
            대시보드 데이터를 불러올 수 없습니다.
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!data) return null

  return (
    <div className="p-4 space-y-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <HoldingsCard
          positionCount={data.holdings.positionCount}
          asOf={data.holdings.asOf}
        />
        <WatchlistCard
          count={data.watchlist.count}
          byMarket={data.watchlist.byMarket}
        />
        <FxRateCard
          rate={data.fxRate.rate}
          pair={data.fxRate.pair}
          date={data.fxRate.date}
        />
      </div>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ThesesCard
          counts={data.theses.counts}
          total={data.theses.total}
          checkedAt={data.theses.checkedAt}
        />
        <div className="grid grid-cols-1 gap-4">
          <CollectionCard latestDate={data.collection.latestDate} />
          <ReportCard latestDate={data.report.latestDate} />
        </div>
      </div>
    </div>
  )
}

function HoldingsCard({
  positionCount,
  asOf,
}: {
  positionCount: number
  asOf: string | null
}) {
  const staleness = computeStaleness("holdings", asOf)
  const hasData = positionCount > 0

  return (
    <Link href="/config/holdings" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Wallet className="size-4" />
              보유 포지션
            </CardTitle>
            {hasData && <StalenessDot level={staleness} label={`기준일: ${asOf}`} />}
          </div>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <p className="text-3xl font-bold tabular-nums">{positionCount}</p>
              <p className="text-xs text-muted-foreground mt-1">
                기준일 {asOf} ({formatDaysAgo(asOf)})
              </p>
            </>
          ) : (
            <EmptyMetric />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function WatchlistCard({
  count,
  byMarket,
}: {
  count: number
  byMarket: Record<string, number>
}) {
  const hasData = count > 0
  const marketBreakdown = Object.entries(byMarket)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([market, n]) => `${market} ${n}`)
    .join(" / ")

  return (
    <Link href="/config/watchlist" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
            <List className="size-4" />
            추적 종목
          </CardTitle>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <p className="text-3xl font-bold tabular-nums">{count}</p>
              <p className="text-xs text-muted-foreground mt-1">
                {marketBreakdown}
              </p>
            </>
          ) : (
            <EmptyMetric />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function FxRateCard({
  rate,
  pair,
  date,
}: {
  rate: number | null
  pair: string
  date: string | null
}) {
  const staleness = computeStaleness("fxRate", date)
  const hasData = rate !== null

  return (
    <Link href="/data/fx" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <ArrowLeftRight className="size-4" />
              환율 ({pair})
            </CardTitle>
            {hasData && <StalenessDot level={staleness} />}
          </div>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <p className="text-3xl font-bold tabular-nums">
                {formatRate(rate)}
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                {date} ({formatDaysAgo(date)})
              </p>
            </>
          ) : (
            <EmptyMetric label="환율 데이터 없음" />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

const THESIS_STATUS_CONFIG: Record<
  ThesisStatus,
  { label: string; dotColor: string; badgeClassName: string }
> = {
  valid: {
    label: "유효",
    dotColor: "bg-emerald-500",
    badgeClassName:
      "bg-emerald-950 text-emerald-400 border-emerald-800",
  },
  weakening: {
    label: "약화",
    dotColor: "bg-amber-500",
    badgeClassName:
      "bg-amber-950 text-amber-400 border-amber-800",
  },
  invalidated: {
    label: "무효화",
    dotColor: "bg-red-500",
    badgeClassName:
      "bg-red-950 text-red-400 border-red-800",
  },
}

function ThesesCard({
  counts,
  total,
  checkedAt,
}: {
  counts: Record<ThesisStatus, number>
  total: number
  checkedAt: string | null
}) {
  const staleness = computeStaleness("theses", checkedAt)
  const hasData = total > 0

  return (
    <Link href="/research/thesis-check" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <BookOpen className="size-4" />
              논제 현황
            </CardTitle>
            {hasData && <StalenessDot level={staleness} />}
          </div>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <div className="flex items-baseline gap-3">
                <p className="text-3xl font-bold tabular-nums">{total}</p>
                <span className="text-sm text-muted-foreground">논제</span>
              </div>
              <div className="flex flex-wrap gap-2 mt-3">
                {(Object.keys(THESIS_STATUS_CONFIG) as ThesisStatus[]).map(
                  (status) => {
                    const config = THESIS_STATUS_CONFIG[status]
                    const count = counts[status]
                    if (count === 0) return null
                    return (
                      <Badge
                        key={status}
                        variant="outline"
                        className={`gap-1.5 ${config.badgeClassName}`}
                      >
                        <span
                          className={`inline-block size-2 rounded-full ${config.dotColor}`}
                        />
                        {config.label} {count}
                      </Badge>
                    )
                  },
                )}
              </div>
              <p className="text-xs text-muted-foreground mt-3">
                확인일 {checkedAt} ({formatDaysAgo(checkedAt)})
              </p>
            </>
          ) : (
            <EmptyMetric label="/thesis-research 스킬을 실행하세요" />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function CollectionCard({ latestDate }: { latestDate: string | null }) {
  const staleness = computeStaleness("collection", latestDate)
  const hasData = latestDate !== null

  return (
    <Link href="/data/prices" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <CandlestickChart className="size-4" />
              데이터 수집
            </CardTitle>
            {hasData && <StalenessDot level={staleness} />}
          </div>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <p className="text-lg font-semibold tabular-nums">{latestDate}</p>
              <p className="text-xs text-muted-foreground mt-1">
                최신 가격 데이터 ({formatDaysAgo(latestDate)})
              </p>
            </>
          ) : (
            <EmptyMetric label="just collect 명령을 실행하세요" />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function ReportCard({ latestDate }: { latestDate: string | null }) {
  const staleness = computeStaleness("report", latestDate)
  const hasData = latestDate !== null

  return (
    <Link href="/reports" className="block">
      <Card className="transition-colors hover:bg-card/80">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <FileText className="size-4" />
              분석 리포트
            </CardTitle>
            {hasData && <StalenessDot level={staleness} />}
          </div>
        </CardHeader>
        <CardContent>
          {hasData ? (
            <>
              <p className="text-lg font-semibold tabular-nums">{latestDate}</p>
              <p className="text-xs text-muted-foreground mt-1">
                최신 분석 보고서 ({formatDaysAgo(latestDate)})
              </p>
            </>
          ) : (
            <EmptyMetric label="/portfolio-analyze 스킬을 실행하세요" />
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function EmptyMetric({ label }: { label?: string }) {
  return (
    <div className="py-1">
      <p className="text-lg text-muted-foreground">---</p>
      {label && (
        <p className="text-xs text-muted-foreground mt-1">{label}</p>
      )}
    </div>
  )
}

function DashboardSkeleton() {
  return (
    <div className="p-4 space-y-4">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <Skeleton className="h-4 w-24" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-16 mb-2" />
              <Skeleton className="h-3 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <Skeleton className="h-4 w-20" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-8 w-12 mb-3" />
            <div className="flex gap-2">
              <Skeleton className="h-5 w-16" />
              <Skeleton className="h-5 w-16" />
            </div>
          </CardContent>
        </Card>
        <div className="grid grid-cols-1 gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-2">
                <Skeleton className="h-4 w-24" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-5 w-28 mb-2" />
                <Skeleton className="h-3 w-36" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  )
}
