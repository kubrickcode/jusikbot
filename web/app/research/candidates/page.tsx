import { PageHeader } from "@/components/page-header"
import { CandidatesViewer } from "@/components/research/candidates-viewer"

export const dynamic = "force-dynamic"

export default function CandidatesPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="후보 종목"
        description="리서치에서 발굴된 투자 후보"
      />
      <CandidatesViewer />
    </div>
  )
}
