import { PageHeader } from "@/components/page-header"
import { DashboardViewer } from "@/components/dashboard/dashboard-viewer"

export const dynamic = "force-dynamic"

export default function HomePage() {
  return (
    <div className="flex flex-col h-full overflow-auto">
      <PageHeader title="대시보드" description="포트폴리오 현황 개요" />
      <DashboardViewer />
    </div>
  )
}
