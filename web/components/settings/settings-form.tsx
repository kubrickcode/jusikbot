"use client"

import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import type { EtfOption, Settings } from "@/lib/types/settings"
import { Loader2 } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import {
  CoreRatioEditor,
  entriesToRatioMap,
  ratioMapToEntries,
} from "./core-ratio-editor"
import { formatKrw, NumberField } from "./number-field"
import { SumIndicator } from "./sum-indicator"

type SettingsFormProps = {
  initialValues: Settings
  etfOptions: EtfOption[]
}

type SaveState = "idle" | "saving" | "success" | "error"

export function SettingsForm({ initialValues, etfOptions }: SettingsFormProps) {
  const [saveState, setSaveState] = useState<SaveState>("idle")
  const [saveError, setSaveError] = useState("")
  const [coreRatioEntries, setCoreRatioEntries] = useState(
    ratioMapToEntries(initialValues.strategy.core_internal_ratio),
  )

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors, isDirty },
  } = useForm<Settings>({
    defaultValues: initialValues,
  })

  const watchedValues = watch()

  const isFormDirty =
    isDirty ||
    JSON.stringify(entriesToRatioMap(coreRatioEntries)) !==
      JSON.stringify(initialValues.strategy.core_internal_ratio)

  useEffect(() => {
    if (!isFormDirty) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isFormDirty])

  const onSubmit = useCallback(
    async (data: Settings) => {
      setSaveState("saving")
      setSaveError("")

      const payload = {
        ...data,
        $schema: "./settings.schema.json",
        strategy: {
          ...data.strategy,
          core_internal_ratio: entriesToRatioMap(coreRatioEntries),
        },
      }

      // Coerce string values from number inputs to actual numbers
      const coerced = coerceNumbers(payload) as Settings & { $schema: string }

      try {
        const res = await fetch("/api/config/settings", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(coerced),
        })

        if (!res.ok) {
          const body = await res.json()
          if (body.errors) {
            const messages = body.errors
              .map((e: { path: string; message: string }) => `${e.path}: ${e.message}`)
              .join("\n")
            throw new Error(messages)
          }
          throw new Error(body.error || "저장 실패")
        }

        setSaveState("success")
        reset(data)
        setCoreRatioEntries(ratioMapToEntries(coerced.strategy.core_internal_ratio))
        setTimeout(() => setSaveState("idle"), 1500)
      } catch (err) {
        setSaveState("error")
        setSaveError(err instanceof Error ? err.message : "알 수 없는 오류")
      }
    },
    [coreRatioEntries, reset],
  )

  const handleReset = useCallback(() => {
    reset(initialValues)
    setCoreRatioEntries(ratioMapToEntries(initialValues.strategy.core_internal_ratio))
  }, [initialValues, reset])

  const registerNumber = (name: Parameters<typeof register>[0]) =>
    register(name, { valueAsNumber: true })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 p-6 pb-24">
      {saveState === "error" && saveError && (
        <Alert variant="destructive">
          <AlertDescription className="whitespace-pre-wrap">{saveError}</AlertDescription>
        </Alert>
      )}

      {/* Section 1: 투자 예산 */}
      <Card>
        <CardHeader>
          <CardTitle>투자 예산</CardTitle>
          <CardDescription>총 예산과 배분 단위</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <NumberField
              label="총 예산"
              suffix="원"
              registration={registerNumber("budget_krw")}
              error={errors.budget_krw}
              helperText={formatKrw(watchedValues.budget_krw)}
            />
            <NumberField
              label="배분 단위"
              suffix="원"
              registration={registerNumber("adjustment_unit_krw")}
              error={errors.adjustment_unit_krw}
              helperText={formatKrw(watchedValues.adjustment_unit_krw)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Section 2: 코어/위성 전략 */}
      <Card>
        <CardHeader>
          <CardTitle>코어/위성 전략</CardTitle>
          <CardDescription>코어-위성 비율과 코어 내부 구성</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <div className="grid grid-cols-2 gap-4">
              <NumberField
                label="코어 비율"
                suffix="%"
                registration={registerNumber("strategy.core_pct")}
                error={errors.strategy?.core_pct}
              />
              <NumberField
                label="위성 비율"
                suffix="%"
                registration={registerNumber("strategy.satellite_pct")}
                error={errors.strategy?.satellite_pct}
              />
            </div>
            <div className="mt-2">
              <SumIndicator
                values={[watchedValues.strategy?.core_pct, watchedValues.strategy?.satellite_pct]}
                target={100}
              />
            </div>
          </div>
          <Separator />
          <CoreRatioEditor
            entries={coreRatioEntries}
            etfOptions={etfOptions}
            onChange={setCoreRatioEntries}
          />
        </CardContent>
      </Card>

      {/* Section 3: 리스크 한도 */}
      <Card>
        <CardHeader>
          <CardTitle>리스크 한도</CardTitle>
          <CardDescription>종목별, 섹터별 집중도 및 낙폭 기준</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <p className="text-sm font-medium mb-3 text-muted-foreground">집중도 한도</p>
            <div className="grid grid-cols-3 gap-4">
              <NumberField
                label="개별주 최대"
                suffix="%"
                registration={registerNumber("risk_tolerance.max_single_stock_pct")}
                error={errors.risk_tolerance?.max_single_stock_pct}
              />
              <NumberField
                label="ETF 최대"
                suffix="%"
                registration={registerNumber("risk_tolerance.max_single_etf_pct")}
                error={errors.risk_tolerance?.max_single_etf_pct}
              />
              <NumberField
                label="섹터 최대"
                suffix="%"
                registration={registerNumber("risk_tolerance.max_sector_concentration_pct")}
                error={errors.risk_tolerance?.max_sector_concentration_pct}
              />
            </div>
          </div>
          <Separator />
          <div>
            <p className="text-sm font-medium mb-3 text-muted-foreground">낙폭 기준</p>
            <div className="grid grid-cols-2 gap-4">
              <NumberField
                label="경고 낙폭"
                suffix="%"
                registration={registerNumber("risk_tolerance.max_drawdown_warning_pct")}
                error={errors.risk_tolerance?.max_drawdown_warning_pct}
              />
              <NumberField
                label="행동 낙폭"
                suffix="%"
                registration={registerNumber("risk_tolerance.max_drawdown_action_pct")}
                error={errors.risk_tolerance?.max_drawdown_action_pct}
              />
            </div>
          </div>
          <Separator />
          <div className="grid grid-cols-2 gap-4">
            <NumberField
              label="최소 포지션"
              suffix="원"
              registration={registerNumber("risk_tolerance.min_position_size_krw")}
              error={errors.risk_tolerance?.min_position_size_krw}
              helperText={formatKrw(watchedValues.risk_tolerance?.min_position_size_krw)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Section 4: 확신도별 배분 */}
      <Card>
        <CardHeader>
          <CardTitle>확신도별 배분</CardTitle>
          <CardDescription>확신도 등급별 풀 비율 (합계 100%)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4">
            <NumberField
              label="High"
              suffix="%"
              registration={registerNumber("sizing.high_confidence_pool_pct")}
              error={errors.sizing?.high_confidence_pool_pct}
            />
            <NumberField
              label="Medium"
              suffix="%"
              registration={registerNumber("sizing.medium_confidence_pool_pct")}
              error={errors.sizing?.medium_confidence_pool_pct}
            />
            <NumberField
              label="Low"
              suffix="%"
              registration={registerNumber("sizing.low_confidence_pool_pct")}
              error={errors.sizing?.low_confidence_pool_pct}
            />
          </div>
          <div className="mt-2">
            <SumIndicator
              values={[
                watchedValues.sizing?.high_confidence_pool_pct,
                watchedValues.sizing?.medium_confidence_pool_pct,
                watchedValues.sizing?.low_confidence_pool_pct,
              ]}
              target={100}
            />
          </div>
        </CardContent>
      </Card>

      {/* Section 5: 변경 속도 제한 */}
      <Card>
        <CardHeader>
          <CardTitle>변경 속도 제한</CardTitle>
          <CardDescription>과잉거래 방지를 위한 월간/분기 변경 한도</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div>
            <p className="text-sm font-medium mb-3 text-muted-foreground">월간 리뷰</p>
            <div className="grid grid-cols-2 gap-4">
              <NumberField
                label="종목당"
                suffix="원"
                registration={registerNumber("anchoring.monthly_max_change_per_position_krw")}
                error={errors.anchoring?.monthly_max_change_per_position_krw}
                helperText={formatKrw(
                  watchedValues.anchoring?.monthly_max_change_per_position_krw,
                )}
              />
              <NumberField
                label="전체 합계"
                suffix="원"
                registration={registerNumber("anchoring.monthly_max_total_change_krw")}
                error={errors.anchoring?.monthly_max_total_change_krw}
                helperText={formatKrw(watchedValues.anchoring?.monthly_max_total_change_krw)}
              />
            </div>
          </div>
          <Separator />
          <div>
            <p className="text-sm font-medium mb-3 text-muted-foreground">분기 리뷰</p>
            <div className="grid grid-cols-2 gap-4">
              <NumberField
                label="종목당"
                suffix="원"
                registration={registerNumber("anchoring.quarterly_max_change_per_position_krw")}
                error={errors.anchoring?.quarterly_max_change_per_position_krw}
                helperText={formatKrw(
                  watchedValues.anchoring?.quarterly_max_change_per_position_krw,
                )}
              />
              <NumberField
                label="전체 합계"
                suffix="원"
                registration={registerNumber("anchoring.quarterly_max_total_change_krw")}
                error={errors.anchoring?.quarterly_max_total_change_krw}
                helperText={formatKrw(
                  watchedValues.anchoring?.quarterly_max_total_change_krw,
                )}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Section 6: 운영 파라미터 */}
      <Card>
        <CardHeader>
          <CardTitle>운영 파라미터</CardTitle>
          <CardDescription>리뷰 주기와 데이터 유효 기간</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <NumberField
              label="최소 리뷰 간격"
              suffix="일"
              registration={registerNumber("min_review_interval_days")}
              error={errors.min_review_interval_days}
            />
            <NumberField
              label="보유현황 유효기간"
              suffix="일"
              registration={registerNumber("holdings_staleness_threshold_days")}
              error={errors.holdings_staleness_threshold_days}
            />
          </div>
        </CardContent>
      </Card>

      {/* Sticky Save Bar */}
      {isFormDirty && (
        <div className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background/95 backdrop-blur-sm">
          <div className="flex items-center justify-end gap-3 px-6 py-3 max-w-4xl mx-auto">
            <span className="text-sm text-muted-foreground mr-auto">변경사항이 있습니다</span>
            <Button type="button" variant="outline" onClick={handleReset}>
              초기화
            </Button>
            <Button type="submit" disabled={saveState === "saving"}>
              {saveState === "saving" ? (
                <>
                  <Loader2 className="size-4 mr-1 animate-spin" />
                  저장 중...
                </>
              ) : saveState === "success" ? (
                "저장 완료"
              ) : (
                "저장"
              )}
            </Button>
          </div>
        </div>
      )}
    </form>
  )
}

function coerceNumbers(obj: unknown): Record<string, unknown> {
  if (obj === null || typeof obj !== "object") return obj as Record<string, unknown>
  if (Array.isArray(obj)) return obj.map(coerceNumbers) as unknown as Record<string, unknown>

  const result: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
    if (typeof value === "string" && value !== "" && !isNaN(Number(value))) {
      result[key] = Number(value)
    } else if (typeof value === "object" && value !== null) {
      result[key] = coerceNumbers(value)
    } else {
      result[key] = value
    }
  }
  return result
}
