import { Sliders } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function SettingsPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="예산 및 리스크"
        description="투자 예산, 리스크 한도, 포지션 사이징 설정"
      />
      <PageStub icon={Sliders} />
    </div>
  )
}
