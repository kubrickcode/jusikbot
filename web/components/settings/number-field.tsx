"use client"

import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { FieldError, UseFormRegisterReturn } from "react-hook-form"

type NumberFieldProps = {
  label: string
  suffix: "원" | "%" | "일"
  registration: UseFormRegisterReturn
  error?: FieldError
  helperText?: string
  step?: number
}

export function NumberField({
  label,
  suffix,
  registration,
  error,
  helperText,
  step = 1,
}: NumberFieldProps) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={registration.name}>{label}</Label>
      <div className="relative">
        <Input
          id={registration.name}
          type="number"
          step={step}
          className="pr-10"
          aria-invalid={!!error}
          aria-describedby={error ? `${registration.name}-error` : undefined}
          {...registration}
        />
        <span className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">
          {suffix}
        </span>
      </div>
      {error ? (
        <p id={`${registration.name}-error`} className="text-xs text-destructive">
          {error.message}
        </p>
      ) : helperText ? (
        <p className="text-xs text-muted-foreground">{helperText}</p>
      ) : null}
    </div>
  )
}

export function formatKrw(value: number | undefined): string {
  if (value === undefined || isNaN(value)) return ""
  return new Intl.NumberFormat("ko-KR").format(value) + "원"
}
