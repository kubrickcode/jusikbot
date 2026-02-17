import { Clock } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function HistoryPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="리서치 히스토리"
        description="과거 팩트체크 결과 아카이브"
      />
      <PageStub icon={Clock} />
    </div>
  )
}
