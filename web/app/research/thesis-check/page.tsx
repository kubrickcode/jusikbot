import { CheckCircle } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function ThesisCheckPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="논제 팩트체크"
        description="논제별 유효성 조건 충족 여부"
      />
      <PageStub icon={CheckCircle} />
    </div>
  )
}
