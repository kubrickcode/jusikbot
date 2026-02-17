"use client"

import { useCallback, useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { Candidate, CandidatesResult } from "@/lib/types/research"
import { EmptyResearchState } from "./thesis-status-badge"

type FetchState = "idle" | "loading" | "not-found" | "error"

const MARKET_CAP_LABEL: Record<string, string> = {
  large: "대형",
  mid: "중형",
  small: "소형",
}

export function CandidatesViewer() {
  const [data, setData] = useState<CandidatesResult | null>(null)
  const [fetchState, setFetchState] = useState<FetchState>("loading")

  const fetchData = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch("/api/research/candidates")
      if (res.status === 404) {
        setFetchState("not-found")
        return
      }
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setData(json as CandidatesResult)
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  if (fetchState === "loading") return <LoadingSkeleton />

  if (fetchState === "not-found") {
    return (
      <EmptyResearchState
        title="후보 종목 없음"
        description="/thesis-research 스킬을 먼저 실행하세요"
      />
    )
  }

  if (fetchState === "error") {
    return (
      <EmptyResearchState
        title="조회 실패"
        description="candidates.json 파일을 읽을 수 없습니다"
      />
    )
  }

  if (!data || data.candidates.length === 0) {
    return (
      <EmptyResearchState
        title="후보 종목 없음"
        description="최근 리서치에서 발굴된 후보 종목이 없습니다"
      />
    )
  }

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <div className="px-4 py-2 text-xs text-muted-foreground">
        확인일: {data.checked_at} | {data.candidates.length}개 후보
      </div>
      <div className="flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>종목</TableHead>
              <TableHead>시장</TableHead>
              <TableHead>유형</TableHead>
              <TableHead>섹터</TableHead>
              <TableHead>시가총액</TableHead>
              <TableHead>관련 논제</TableHead>
              <TableHead className="min-w-[200px]">근거</TableHead>
              <TableHead className="min-w-[150px]">리스크</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.candidates.map((candidate) => (
              <CandidateRow key={candidate.symbol} candidate={candidate} />
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

function CandidateRow({ candidate }: { candidate: Candidate }) {
  return (
    <TableRow>
      <TableCell>
        <div>
          <span className="font-mono font-medium text-sm">{candidate.symbol}</span>
          <p className="text-xs text-muted-foreground">{candidate.name}</p>
        </div>
      </TableCell>
      <TableCell>
        <Badge
          variant="outline"
          className={candidate.market === "US" ? "text-blue-600 border-blue-200" : "text-red-600 border-red-200"}
        >
          {candidate.market}
        </Badge>
      </TableCell>
      <TableCell className="text-xs">{candidate.type}</TableCell>
      <TableCell className="text-xs">{candidate.sector}</TableCell>
      <TableCell>
        <Badge variant="secondary" className="text-xs">
          {MARKET_CAP_LABEL[candidate.market_cap_category] ?? candidate.market_cap_category}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {candidate.related_theses.map((thesis) => (
            <Badge key={thesis} variant="outline" className="text-[10px]">
              {thesis}
            </Badge>
          ))}
        </div>
      </TableCell>
      <TableCell className="text-xs max-w-[250px]">
        <p className="line-clamp-3">{candidate.rationale}</p>
      </TableCell>
      <TableCell className="text-xs max-w-[200px] text-amber-600">
        <p className="line-clamp-3">{candidate.risks}</p>
      </TableCell>
    </TableRow>
  )
}

function LoadingSkeleton() {
  return (
    <div className="p-4">
      <Table>
        <TableHeader>
          <TableRow>
            {Array.from({ length: 8 }).map((_, i) => (
              <TableHead key={i}>
                <Skeleton className="h-4 w-16" />
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 5 }).map((_, i) => (
            <TableRow key={i}>
              {Array.from({ length: 8 }).map((_, j) => (
                <TableCell key={j}>
                  <Skeleton className="h-4 w-20" />
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
