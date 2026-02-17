"use client"

import { cn } from "@/lib/utils"
import { Columns2, Eye, Pencil } from "lucide-react"

export type ViewMode = "edit" | "split" | "preview"

type ViewModeToggleProps = {
  mode: ViewMode
  onChange: (mode: ViewMode) => void
}

const modes = [
  { value: "edit" as const, icon: Pencil, label: "편집" },
  { value: "split" as const, icon: Columns2, label: "분할" },
  { value: "preview" as const, icon: Eye, label: "미리보기" },
] as const

export function ViewModeToggle({ mode, onChange }: ViewModeToggleProps) {
  return (
    <div className="flex items-center gap-0.5 rounded-md border border-border p-1">
      {modes.map(({ value, icon: Icon, label }) => (
        <button
          key={value}
          type="button"
          onClick={() => onChange(value)}
          className={cn(
            "flex items-center gap-1.5 px-2.5 py-1 rounded text-xs transition-colors",
            mode === value
              ? "bg-primary text-primary-foreground"
              : "text-muted-foreground hover:text-foreground hover:bg-muted",
          )}
          title={label}
        >
          <Icon className="size-3.5" />
          <span>{label}</span>
        </button>
      ))}
    </div>
  )
}
