import { BookOpen } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function ThesesPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="투자 논제"
        description="투자 논제와 유효/무효화 조건"
      />
      <PageStub icon={BookOpen} />
    </div>
  )
}
