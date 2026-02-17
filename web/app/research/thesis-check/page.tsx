import { PageHeader } from "@/components/page-header"
import { ThesisCheckViewer } from "@/components/research/thesis-check-viewer"

export const dynamic = "force-dynamic"

export default function ThesisCheckPage() {
  return (
    <div className="flex flex-col h-full overflow-auto">
      <PageHeader
        title="논제 팩트체크"
        description="논제별 유효성 조건 충족 여부"
      />
      <ThesisCheckViewer />
    </div>
  )
}
