"use client"

import { cn } from "@/lib/utils"

type SumIndicatorProps = {
  values: (number | undefined)[]
  target: number
  label?: string
}

export function SumIndicator({ values, target, label = "합계" }: SumIndicatorProps) {
  const sum = values.reduce<number>((acc, v) => acc + (Number(v) || 0), 0)
  const isValid = sum === target

  return (
    <p
      aria-live="polite"
      className={cn("text-xs", isValid ? "text-muted-foreground" : "text-destructive")}
    >
      {label}: {sum}%{!isValid && ` (${target}%여야 합니다)`}
    </p>
  )
}
