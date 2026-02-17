"use client"

import { useCallback, useEffect, useState } from "react"
import { ArrowLeft, Calendar, FileText } from "lucide-react"
import type { Components } from "react-markdown"
import Markdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { EmptyResearchState } from "@/components/research/thesis-status-badge"

type ReportEntry = {
  date: string
  filename: string
}

type ViewMode = "list" | "detail"

const markdownComponents: Components = {
  h1: ({ children }) => (
    <h1 className="text-2xl font-bold mt-8 mb-4 pb-2 border-b border-border first:mt-0">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-xl font-semibold mt-8 mb-3 text-foreground">{children}</h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-base font-semibold mt-4 mb-2 text-foreground">{children}</h3>
  ),
  h4: ({ children }) => (
    <h4 className="text-sm font-semibold mt-3 mb-1.5 text-foreground">{children}</h4>
  ),
  p: ({ children }) => (
    <p className="text-sm leading-relaxed mb-3 text-foreground">{children}</p>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold text-foreground">{children}</strong>
  ),
  blockquote: ({ children }) => (
    <blockquote className="border-l-2 border-primary pl-4 my-4 text-muted-foreground text-sm italic">
      {children}
    </blockquote>
  ),
  code: ({ children, className }) => {
    const isBlock = className?.includes("language-")
    if (isBlock) {
      return (
        <code className="block bg-muted rounded-md p-4 text-xs font-mono leading-relaxed overflow-x-auto whitespace-pre">
          {children}
        </code>
      )
    }
    return <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-mono">{children}</code>
  },
  pre: ({ children }) => <pre className="my-4">{children}</pre>,
  ul: ({ children }) => (
    <ul className="list-disc list-outside pl-5 space-y-1 mb-3 text-sm">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="list-decimal list-outside pl-5 space-y-1 mb-3 text-sm">{children}</ol>
  ),
  li: ({ children }) => (
    <li className="text-sm text-foreground leading-relaxed">{children}</li>
  ),
  table: ({ children }) => (
    <div className="overflow-x-auto my-4">
      <table className="w-full text-sm border-collapse">{children}</table>
    </div>
  ),
  thead: ({ children }) => <thead>{children}</thead>,
  tbody: ({ children }) => <tbody>{children}</tbody>,
  tr: ({ children }) => <tr className="border-b border-border">{children}</tr>,
  th: ({ children }) => (
    <th className="border border-border px-3 py-2 text-left font-medium bg-muted text-xs">
      {children}
    </th>
  ),
  td: ({ children }) => (
    <td className="border border-border px-3 py-2 text-xs text-muted-foreground font-mono">
      {children}
    </td>
  ),
  a: ({ children, href }) => (
    <a
      href={href}
      className="text-primary underline underline-offset-2 hover:opacity-80"
      target="_blank"
      rel="noopener noreferrer"
    >
      {children}
    </a>
  ),
  hr: () => <hr className="my-6 border-border" />,
}

export function ReportViewer() {
  const [entries, setEntries] = useState<ReportEntry[]>([])
  const [fetchState, setFetchState] = useState<"loading" | "idle" | "error">("loading")
  const [selectedDate, setSelectedDate] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>("list")

  const fetchEntries = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch("/api/reports")
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setEntries(json.entries as ReportEntry[])
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [])

  useEffect(() => {
    fetchEntries()
  }, [fetchEntries])

  const selectReport = useCallback((date: string) => {
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
        description="리포트 목록을 읽을 수 없습니다"
      />
    )
  }

  if (entries.length === 0) {
    return (
      <EmptyResearchState
        title="리포트 없음"
        description="/portfolio-analyze 스킬을 실행하면 분석 리포트가 생성됩니다"
      />
    )
  }

  if (viewMode === "detail" && selectedDate) {
    return <ReportDetail date={selectedDate} onBack={goBack} />
  }

  return (
    <div className="p-4 space-y-2">
      <p className="text-xs text-muted-foreground mb-3">{entries.length}개 리포트</p>
      {entries.map((entry) => (
        <button
          key={entry.date}
          onClick={() => selectReport(entry.date)}
          className="flex items-center gap-3 w-full text-left px-4 py-3 rounded-lg border hover:bg-muted/50 transition-colors"
        >
          <Calendar className="size-4 text-muted-foreground shrink-0" />
          <span className="font-mono text-sm">{entry.date}</span>
          <FileText className="size-3.5 text-muted-foreground ml-auto" />
        </button>
      ))}
    </div>
  )
}

function ReportDetail({ date, onBack }: { date: string; onBack: () => void }) {
  const [content, setContent] = useState<string | null>(null)
  const [fetchState, setFetchState] = useState<"loading" | "idle" | "error">("loading")

  const fetchReport = useCallback(async () => {
    setFetchState("loading")
    try {
      const res = await fetch(`/api/reports/${date}`)
      if (!res.ok) throw new Error("조회 실패")
      const json = await res.json()
      setContent(json.content as string)
      setFetchState("idle")
    } catch {
      setFetchState("error")
    }
  }, [date])

  useEffect(() => {
    fetchReport()
  }, [fetchReport])

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3 border-b">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="size-4 mr-1" />
          목록
        </Button>
        <Badge variant="outline" className="font-mono">{date}</Badge>
      </div>

      <div className="flex-1 overflow-auto">
        {fetchState === "loading" && (
          <div className="p-4 space-y-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        )}

        {fetchState === "error" && (
          <EmptyResearchState
            title="조회 실패"
            description={`${date} 리포트를 읽을 수 없습니다`}
          />
        )}

        {fetchState === "idle" && content && (
          <div className="p-6">
            <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
              {content}
            </Markdown>
          </div>
        )}
      </div>
    </div>
  )
}
