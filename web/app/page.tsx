import { TrendingUp } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function HomePage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader title="대시보드" description="포트폴리오 현황 개요" />
      <PageStub icon={TrendingUp} />
    </div>
  )
}
