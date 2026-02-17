import type { StalenessLevel } from "@/lib/types/dashboard"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

const STALENESS_STYLES: Record<StalenessLevel, { className: string; label: string }> = {
  critical: {
    className: "bg-red-500",
    label: "데이터가 매우 오래됨",
  },
  fresh: {
    className: "bg-emerald-500",
    label: "최신 상태",
  },
  stale: {
    className: "bg-amber-500",
    label: "갱신 필요",
  },
}

type StalenessDotProps = {
  level: StalenessLevel
  label?: string
}

export function StalenessDot({ level, label }: StalenessDotProps) {
  const config = STALENESS_STYLES[level]
  const tooltipLabel = label ?? config.label

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className={`inline-block size-2 rounded-full shrink-0 ${config.className}`}
            aria-label={tooltipLabel}
          />
        </TooltipTrigger>
        <TooltipContent>{tooltipLabel}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
