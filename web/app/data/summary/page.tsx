import { readFile } from "fs/promises"
import { resolve } from "path"
import { PageHeader } from "@/components/page-header"
import { SummaryViewer } from "@/components/summary/summary-viewer"
import { DATA_DIR } from "@/lib/paths"

const SUMMARY_PATH = resolve(DATA_DIR, "summary.md")

async function loadSummary(): Promise<string> {
  try {
    return await readFile(SUMMARY_PATH, "utf-8")
  } catch {
    return ""
  }
}

export const dynamic = "force-dynamic"

export default async function SummaryPage() {
  const content = await loadSummary()

  return (
    <div className="flex flex-col h-full overflow-auto">
      <PageHeader title="기술 지표 요약" description="14개 컬럼 기술 지표 요약 테이블" />
      <SummaryViewer content={content} />
    </div>
  )
}
