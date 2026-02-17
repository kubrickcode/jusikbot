import { Badge } from "@/components/ui/badge"
import type { ConditionStatus, StatusTransition, ThesisStatus } from "@/lib/types/research"
import {
  ArrowDown,
  ArrowUp,
  CheckCircle,
  CircleAlert,
  CircleDashed,
  CircleHelp,
  CircleMinus,
  Clock,
  Minus,
  Sparkles,
} from "lucide-react"

const THESIS_STATUS_CONFIG: Record<ThesisStatus, { label: string; className: string }> = {
  valid: { label: "유효", className: "bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-400 dark:border-emerald-800" },
  weakening: { label: "약화 중", className: "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950 dark:text-amber-400 dark:border-amber-800" },
  invalidated: { label: "무효화", className: "bg-red-50 text-red-700 border-red-200 dark:bg-red-950 dark:text-red-400 dark:border-red-800" },
}

const CONDITION_STATUS_CONFIG: Record<ConditionStatus, { label: string; className: string; icon: typeof CheckCircle }> = {
  met: { label: "충족", className: "bg-emerald-50 text-emerald-700 border-emerald-200", icon: CheckCircle },
  partially_met: { label: "부분 충족", className: "bg-amber-50 text-amber-700 border-amber-200", icon: CircleMinus },
  not_yet: { label: "미확인", className: "bg-blue-50 text-blue-700 border-blue-200", icon: Clock },
  refuted: { label: "반박됨", className: "bg-red-50 text-red-700 border-red-200", icon: CircleAlert },
  unknown: { label: "불명", className: "bg-slate-50 text-slate-600 border-slate-200", icon: CircleHelp },
}

const TRANSITION_CONFIG: Record<StatusTransition, { label: string; icon: typeof ArrowUp; className: string }> = {
  improving: { label: "개선", icon: ArrowUp, className: "text-emerald-600" },
  degrading: { label: "악화", icon: ArrowDown, className: "text-red-600" },
  stable: { label: "유지", icon: Minus, className: "text-slate-400" },
  new: { label: "신규", icon: Sparkles, className: "text-blue-500" },
}

const SOURCE_TIER_CONFIG: Record<number, { label: string; className: string }> = {
  1: { label: "T1", className: "bg-slate-800 text-white" },
  2: { label: "T2", className: "bg-slate-600 text-white" },
  3: { label: "T3", className: "bg-slate-400 text-white" },
  4: { label: "T4", className: "bg-slate-200 text-slate-600" },
}

export function ThesisStatusBadge({ status }: { status: ThesisStatus }) {
  const config = THESIS_STATUS_CONFIG[status]
  return (
    <Badge variant="outline" className={config.className}>
      {config.label}
    </Badge>
  )
}

export function ConditionStatusBadge({ status }: { status: ConditionStatus }) {
  const config = CONDITION_STATUS_CONFIG[status]
  const Icon = config.icon
  return (
    <Badge variant="outline" className={`gap-1 ${config.className}`}>
      <Icon className="size-3" />
      {config.label}
    </Badge>
  )
}

export function TransitionIndicator({ transition }: { transition?: StatusTransition | null }) {
  if (!transition) return null
  const config = TRANSITION_CONFIG[transition]
  const Icon = config.icon
  return (
    <span className={`inline-flex items-center gap-0.5 text-xs ${config.className}`}>
      <Icon className="size-3" />
      {config.label}
    </span>
  )
}

export function SourceTierBadge({ tier }: { tier: number }) {
  const config = SOURCE_TIER_CONFIG[tier] ?? SOURCE_TIER_CONFIG[4]
  return (
    <Badge className={`text-[10px] px-1 py-0 ${config.className}`}>
      {config.label}
    </Badge>
  )
}

export function ConditionTypeBadge({ type }: { type: "validity" | "invalidation" }) {
  const isValidity = type === "validity"
  return (
    <Badge
      variant="outline"
      className={`text-[10px] ${
        isValidity
          ? "bg-emerald-50 text-emerald-600 border-emerald-200"
          : "bg-red-50 text-red-600 border-red-200"
      }`}
    >
      {isValidity ? "유효 조건" : "무효화 조건"}
    </Badge>
  )
}

export function EmptyResearchState({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex flex-1 items-center justify-center py-16">
      <div className="flex flex-col items-center gap-3 text-center text-muted-foreground">
        <CircleDashed className="size-12 stroke-1" />
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs max-w-xs">{description}</p>
      </div>
    </div>
  )
}
