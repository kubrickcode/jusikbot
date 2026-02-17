import { FileText } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function ReportsPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="분석 리포트"
        description="포트폴리오 분석 결과 보고서"
      />
      <PageStub icon={FileText} />
    </div>
  )
}
