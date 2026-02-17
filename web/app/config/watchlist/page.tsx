import { List } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function WatchlistPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="추적 종목"
        description="관심 종목 추가, 수정, 삭제"
      />
      <PageStub icon={List} />
    </div>
  )
}
