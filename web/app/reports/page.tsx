import { PageHeader } from "@/components/page-header"
import { ReportViewer } from "@/components/reports/report-viewer"

export const dynamic = "force-dynamic"

export default function ReportsPage() {
  return (
    <div className="flex flex-col h-full overflow-auto">
      <PageHeader
        title="분석 리포트"
        description="포트폴리오 분석 결과 보고서"
      />
      <ReportViewer />
    </div>
  )
}
