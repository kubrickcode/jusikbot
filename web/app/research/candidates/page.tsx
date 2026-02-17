import { Search } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function CandidatesPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="후보 종목"
        description="리서치에서 발굴된 투자 후보"
      />
      <PageStub icon={Search} />
    </div>
  )
}
