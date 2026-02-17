import { PageHeader } from "@/components/page-header"
import { SettingsForm } from "@/components/settings/settings-form"
import { configPaths } from "@/lib/paths"
import type { EtfOption, Settings } from "@/lib/types/settings"
import { readFile } from "fs/promises"

async function loadSettings(): Promise<Settings> {
  const raw = await readFile(configPaths.settings, "utf-8")
  return JSON.parse(raw) as Settings
}

async function loadEtfOptions(): Promise<EtfOption[]> {
  const raw = await readFile(configPaths.watchlist, "utf-8")
  const watchlist = JSON.parse(raw) as Array<{
    symbol: string
    name: string
    type: string
  }>
  return watchlist.filter((item) => item.type === "etf").map(({ symbol, name }) => ({ symbol, name }))
}

export const dynamic = "force-dynamic"

export default async function SettingsPage() {
  const [settings, etfOptions] = await Promise.all([loadSettings(), loadEtfOptions()])

  return (
    <div className="flex flex-col h-full">
      <PageHeader title="예산 및 리스크" description="투자 예산, 리스크 한도, 포지션 사이징 설정" />
      <div className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto">
          <SettingsForm initialValues={settings} etfOptions={etfOptions} />
        </div>
      </div>
    </div>
  )
}
