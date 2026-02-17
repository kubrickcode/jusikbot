import { Wallet } from "lucide-react"
import { PageHeader } from "@/components/page-header"
import { PageStub } from "@/components/page-stub"

export default function HoldingsPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="보유 현황"
        description="현재 포지션의 수량과 평균 단가"
      />
      <PageStub icon={Wallet} />
    </div>
  )
}
