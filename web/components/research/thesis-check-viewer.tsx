"use client"

import { useCallback, useEffect, useState } from "react"
import { ChevronDown, ExternalLink } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Skeleton } from "@/components/ui/skeleton"
import type { Thesis, ThesisCheckResult, ThesisCondition } from "@/lib/types/research"
import {
  ConditionStatusBadge,
  ConditionTypeBadge,
  EmptyResearchState,
  SourceTierBadge,
  ThesisStatusBadge,
  TransitionIndicator,
} from "./thesis-status-badge"

type FetchState = "idle" | "loading" | "not-found" | "error"

export function ThesisCheckViewer() {
  const [data, setData] = useState<ThesisCheckResult | null>(null)
  const [fetchState, setFetchState] = useState<FetchState>("loading")

  const fetchData = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch("/api/research/thesis-check")
      if (res.status === 404) {
        setFetchState("not-found")
        return
      }
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setData(json as ThesisCheckResult)
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
        title="리서치 결과 없음"
        description="/thesis-research 스킬을 먼저 실행하세요"
      />
    )
  }

  if (fetchState === "error") {
    return (
      <EmptyResearchState
        title="조회 실패"
        description="thesis-check.json 파일을 읽을 수 없습니다"
      />
    )
  }

  if (!data) return null

  return (
    <div className="p-4 space-y-4">
      <p className="text-xs text-muted-foreground">
        확인일: {data.checked_at}
      </p>
      {data.theses.map((thesis) => (
        <ThesisCard key={thesis.name} thesis={thesis} />
      ))}
    </div>
  )
}

function ThesisCard({ thesis }: { thesis: Thesis }) {
  const validityConditions = thesis.conditions.filter((c) => c.type === "validity")
  const invalidationConditions = thesis.conditions.filter((c) => c.type === "invalidation")

  const validityMet = validityConditions.filter((c) => c.status === "met" || c.status === "partially_met").length
  const invalidationTriggered = invalidationConditions.filter((c) => c.status === "met" || c.status === "refuted").length

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center gap-3 flex-wrap">
          <CardTitle className="text-base">{thesis.name}</CardTitle>
          <ThesisStatusBadge status={thesis.status} />
          <TransitionIndicator transition={thesis.status_transition} />
        </div>
        <div className="flex gap-4 text-xs text-muted-foreground mt-1">
          <span>유효 조건: {validityMet}/{validityConditions.length} 충족</span>
          {invalidationConditions.length > 0 && (
            <span>무효화 조건: {invalidationTriggered}/{invalidationConditions.length} 트리거</span>
          )}
        </div>
        {thesis.upstream_dependency && (
          <p className="text-xs text-amber-600 mt-1">
            상위 의존: {thesis.upstream_dependency}
            {thesis.chain_impact && ` — ${thesis.chain_impact}`}
          </p>
        )}
      </CardHeader>
      <CardContent className="space-y-3">
        {validityConditions.length > 0 && (
          <ConditionGroup title="유효 조건" conditions={validityConditions} borderColor="border-l-emerald-400" />
        )}
        {invalidationConditions.length > 0 && (
          <ConditionGroup title="무효화 조건" conditions={invalidationConditions} borderColor="border-l-red-400" />
        )}
      </CardContent>
    </Card>
  )
}

function ConditionGroup({
  title,
  conditions,
  borderColor,
}: {
  title: string
  conditions: ThesisCondition[]
  borderColor: string
}) {
  return (
    <div className={`border-l-2 ${borderColor} pl-3 space-y-2`}>
      <p className="text-xs font-medium text-muted-foreground">{title}</p>
      {conditions.map((condition, index) => (
        <ConditionRow key={index} condition={condition} />
      ))}
    </div>
  )
}

function ConditionRow({ condition }: { condition: ThesisCondition }) {
  return (
    <Collapsible>
      <CollapsibleTrigger className="flex items-start gap-2 w-full text-left group hover:bg-muted/50 rounded-md p-2 -m-2 transition-colors">
        <ChevronDown className="size-4 mt-0.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-180" />
        <div className="flex-1 min-w-0 space-y-1">
          <div className="flex items-center gap-2 flex-wrap">
            <ConditionStatusBadge status={condition.status} />
            <ConditionTypeBadge type={condition.type} />
            <TransitionIndicator transition={condition.status_transition} />
          </div>
          <p className="text-sm">{condition.text}</p>
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="ml-6 mt-2 space-y-2 text-sm">
          <div className="bg-muted/50 rounded-md p-3">
            <p className="text-xs font-medium text-muted-foreground mb-1">근거</p>
            <p className="text-sm">{condition.evidence}</p>
            {condition.quantitative_distance && (
              <p className="text-xs text-muted-foreground mt-1">
                정량 거리: <code className="bg-muted px-1 rounded">{condition.quantitative_distance}</code>
              </p>
            )}
          </div>
          {condition.sources.length > 0 && (
            <div>
              <p className="text-xs font-medium text-muted-foreground mb-1">출처</p>
              <ul className="space-y-1">
                {condition.sources.map((source, i) => (
                  <li key={i} className="flex items-center gap-2 text-xs">
                    <SourceTierBadge tier={source.tier} />
                    <a
                      href={source.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary underline-offset-2 hover:underline truncate"
                    >
                      {source.title}
                      <ExternalLink className="inline size-3 ml-0.5" />
                    </a>
                    <span className="text-muted-foreground shrink-0">{source.date}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

function LoadingSkeleton() {
  return (
    <div className="p-4 space-y-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <Card key={i}>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-5 w-16" />
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.from({ length: 2 }).map((_, j) => (
              <Skeleton key={j} className="h-12 w-full" />
            ))}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
