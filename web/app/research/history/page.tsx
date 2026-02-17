import { PageHeader } from "@/components/page-header"
import { HistoryViewer } from "@/components/research/history-viewer"

export const dynamic = "force-dynamic"

export default function HistoryPage() {
  return (
    <div className="flex flex-col h-full overflow-auto">
      <PageHeader
        title="리서치 히스토리"
        description="과거 팩트체크 결과 아카이브"
      />
      <HistoryViewer />
    </div>
  )
}
