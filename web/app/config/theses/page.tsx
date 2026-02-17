import { PageHeader } from "@/components/page-header"
import { ThesesEditor } from "@/components/theses/theses-editor"
import { configPaths } from "@/lib/paths"
import { readFile } from "fs/promises"

export const dynamic = "force-dynamic"

export default async function ThesesPage() {
  const content = await readFile(configPaths.theses, "utf-8")

  return (
    <div className="flex flex-col h-full">
      <PageHeader title="투자 논제" description="투자 논제와 유효/무효화 조건" />
      <ThesesEditor initialContent={content} />
    </div>
  )
}
