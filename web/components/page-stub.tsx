import type { LucideIcon } from "lucide-react"

type PageStubProps = {
  icon: LucideIcon
}

export function PageStub({ icon: Icon }: PageStubProps) {
  return (
    <div className="flex flex-1 items-center justify-center">
      <div className="flex flex-col items-center gap-4 text-muted-foreground">
        <Icon className="size-12 stroke-1" />
        <p className="text-sm">준비 중입니다</p>
      </div>
    </div>
  )
}
