"use client"

import { useCallback, useEffect, useState } from "react"
import { ArrowLeft, Clock } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import type { HistoryEntry, ThesisCheckResult } from "@/lib/types/research"
import { EmptyResearchState, ThesisStatusBadge, TransitionIndicator } from "./thesis-status-badge"

type ViewMode = "list" | "detail"

export function HistoryViewer() {
  const [entries, setEntries] = useState<HistoryEntry[]>([])
  const [fetchState, setFetchState] = useState<"loading" | "idle" | "error">("loading")
  const [selectedDate, setSelectedDate] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>("list")

  const fetchEntries = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch("/api/research/history")
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setEntries(json.entries as HistoryEntry[])
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [])

  useEffect(() => {
    fetchEntries()
  }, [fetchEntries])

  const selectEntry = useCallback((date: string) => {
    setSelectedDate(date)
    setViewMode("detail")
  }, [])

  const goBack = useCallback(() => {
    setViewMode("list")
    setSelectedDate(null)
  }, [])

  if (fetchState === "loading") {
    return (
      <div className="p-4 space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    )
  }

  if (fetchState === "error") {
    return (
      <EmptyResearchState
        title="조회 실패"
        description="리서치 히스토리를 읽을 수 없습니다"
      />
    )
  }

  if (entries.length === 0) {
    return (
      <EmptyResearchState
        title="히스토리 없음"
        description="/thesis-research 를 2회 이상 실행하면 히스토리가 누적됩니다"
      />
    )
  }

  if (viewMode === "detail" && selectedDate) {
    return <SnapshotDetail date={selectedDate} onBack={goBack} />
  }

  return (
    <div className="p-4 space-y-2">
      <p className="text-xs text-muted-foreground mb-3">{entries.length}개 스냅샷</p>
      {entries.map((entry) => (
        <button
          key={entry.date}
          onClick={() => selectEntry(entry.date)}
          className="flex items-center gap-3 w-full text-left px-4 py-3 rounded-lg border hover:bg-muted/50 transition-colors"
        >
          <Clock className="size-4 text-muted-foreground shrink-0" />
          <span className="font-mono text-sm">{entry.date}</span>
          <span className="text-xs text-muted-foreground ml-auto">{entry.filename}</span>
        </button>
      ))}
    </div>
  )
}

function SnapshotDetail({ date, onBack }: { date: string; onBack: () => void }) {
  const [data, setData] = useState<ThesisCheckResult | null>(null)
  const [fetchState, setFetchState] = useState<"loading" | "idle" | "error">("loading")

  const fetchSnapshot = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch(`/api/research/history?date=${date}`)
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setData(json as ThesisCheckResult)
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [date])

  useEffect(() => {
    fetchSnapshot()
  }, [fetchSnapshot])

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="size-4 mr-1" />
          목록
        </Button>
        <Badge variant="outline" className="font-mono">{date}</Badge>
      </div>

      {fetchState === "loading" && (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-20 w-full" />
          ))}
        </div>
      )}

      {fetchState === "error" && (
        <EmptyResearchState
          title="조회 실패"
          description={`${date} 스냅샷을 읽을 수 없습니다`}
        />
      )}

      {fetchState === "idle" && data && (
        <div className="space-y-3">
          {data.theses.map((thesis) => (
            <Card key={thesis.name}>
              <CardHeader className="pb-2">
                <div className="flex items-center gap-3 flex-wrap">
                  <CardTitle className="text-sm">{thesis.name}</CardTitle>
                  <ThesisStatusBadge status={thesis.status} />
                  <TransitionIndicator transition={thesis.status_transition} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex gap-4 text-xs text-muted-foreground">
                  <span>
                    유효: {thesis.conditions.filter((c) => c.type === "validity" && (c.status === "met" || c.status === "partially_met")).length}
                    /{thesis.conditions.filter((c) => c.type === "validity").length}
                  </span>
                  <span>
                    무효화: {thesis.conditions.filter((c) => c.type === "invalidation" && (c.status === "met" || c.status === "refuted")).length}
                    /{thesis.conditions.filter((c) => c.type === "invalidation").length}
                  </span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
