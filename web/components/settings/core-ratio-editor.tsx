"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { EtfOption } from "@/lib/types/settings"
import { Plus, X } from "lucide-react"
import { useCallback } from "react"

type CoreRatioEntry = {
  symbol: string
  ratio: number
}

type CoreRatioEditorProps = {
  entries: CoreRatioEntry[]
  etfOptions: EtfOption[]
  onChange: (entries: CoreRatioEntry[]) => void
  error?: string
}

export function CoreRatioEditor({ entries, etfOptions, onChange, error }: CoreRatioEditorProps) {
  const usedSymbols = new Set(entries.map((e) => e.symbol))
  const availableOptions = etfOptions.filter((o) => !usedSymbols.has(o.symbol))

  const updateEntry = useCallback(
    (index: number, field: "symbol" | "ratio", value: string | number) => {
      const updated = entries.map((entry, i) =>
        i === index ? { ...entry, [field]: value } : entry,
      )
      onChange(updated)
    },
    [entries, onChange],
  )

  const addEntry = useCallback(() => {
    const firstAvailable = availableOptions[0]
    if (!firstAvailable) return
    onChange([...entries, { symbol: firstAvailable.symbol, ratio: 1 }])
  }, [entries, availableOptions, onChange])

  const removeEntry = useCallback(
    (index: number) => {
      onChange(entries.filter((_, i) => i !== index))
    },
    [entries, onChange],
  )

  return (
    <div className="space-y-3">
      <Label>코어 ETF 구성 비율</Label>
      <div className="space-y-2">
        {entries.map((entry, index) => (
          <div key={index} className="flex items-center gap-2">
            <select
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={entry.symbol}
              onChange={(e) => updateEntry(index, "symbol", e.target.value)}
            >
              {etfOptions.map((opt) => (
                <option
                  key={opt.symbol}
                  value={opt.symbol}
                  disabled={usedSymbols.has(opt.symbol) && opt.symbol !== entry.symbol}
                >
                  {opt.symbol} - {opt.name}
                </option>
              ))}
            </select>
            <Input
              type="number"
              step={0.1}
              min={0}
              className="w-24 shrink-0"
              value={entry.ratio}
              onChange={(e) => updateEntry(index, "ratio", parseFloat(e.target.value) || 0)}
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="shrink-0"
              disabled={entries.length <= 1}
              onClick={() => removeEntry(index)}
            >
              <X className="size-4" />
            </Button>
          </div>
        ))}
      </div>
      {availableOptions.length > 0 && (
        <Button type="button" variant="outline" size="sm" onClick={addEntry}>
          <Plus className="size-4 mr-1" />
          ETF 추가
        </Button>
      )}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

export function ratioMapToEntries(map: Record<string, number>): CoreRatioEntry[] {
  return Object.entries(map).map(([symbol, ratio]) => ({ symbol, ratio }))
}

export function entriesToRatioMap(entries: CoreRatioEntry[]): Record<string, number> {
  const map: Record<string, number> = {}
  for (const entry of entries) {
    if (entry.symbol) {
      map[entry.symbol] = entry.ratio
    }
  }
  return map
}
