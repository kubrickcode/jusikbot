import { PageHeader } from "@/components/page-header"
import { FxDataViewer } from "@/components/fx/fx-data-viewer"

export default function FxPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader title="환율" description="USD/KRW 환율 데이터" />
      <FxDataViewer />
    </div>
  )
}
