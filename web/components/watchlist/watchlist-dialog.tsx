"use client"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { TagInput } from "@/components/ui/tag-input"
import type { Market, SecurityType, WatchlistItem } from "@/lib/types/watchlist"
import { useCallback, useEffect, useState } from "react"

type WatchlistDialogProps = {
  isOpen: boolean
  onClose: () => void
  onSave: (item: WatchlistItem) => void
  editingItem: WatchlistItem | null
  existingSymbols: string[]
}

const EMPTY_ITEM: WatchlistItem = {
  market: "US",
  name: "",
  sector: "",
  symbol: "",
  themes: [],
  type: "stock",
}

export function WatchlistDialog({
  isOpen,
  onClose,
  onSave,
  editingItem,
  existingSymbols,
}: WatchlistDialogProps) {
  const [item, setItem] = useState<WatchlistItem>(EMPTY_ITEM)
  const [error, setError] = useState("")

  const isEditing = editingItem !== null

  useEffect(() => {
    if (isOpen) {
      setItem(editingItem ?? EMPTY_ITEM)
      setError("")
    }
  }, [isOpen, editingItem])

  const updateField = useCallback(
    <K extends keyof WatchlistItem>(field: K, value: WatchlistItem[K]) => {
      setItem((prev) => ({ ...prev, [field]: value }))
      setError("")
    },
    [],
  )

  const submitItem = useCallback(() => {
    if (!item.symbol.trim()) {
      setError("심볼은 필수입니다")
      return
    }
    if (!item.name.trim()) {
      setError("종목명은 필수입니다")
      return
    }
    if (!item.sector.trim()) {
      setError("섹터는 필수입니다")
      return
    }
    if (!isEditing && existingSymbols.includes(item.symbol.trim())) {
      setError("이미 존재하는 심볼입니다")
      return
    }

    onSave({
      ...item,
      symbol: item.symbol.trim(),
      name: item.name.trim(),
      sector: item.sector.trim(),
    })
    onClose()
  }, [item, isEditing, existingSymbols, onSave, onClose])

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEditing ? "종목 수정" : "종목 추가"}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {error && (
            <p className="text-sm text-destructive">{error}</p>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="watchlist-symbol">심볼</Label>
              <Input
                id="watchlist-symbol"
                value={item.symbol}
                onChange={(e) => updateField("symbol", e.target.value)}
                placeholder="NVDA 또는 360750"
                disabled={isEditing}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="watchlist-name">종목명</Label>
              <Input
                id="watchlist-name"
                value={item.name}
                onChange={(e) => updateField("name", e.target.value)}
                placeholder="NVIDIA"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>거래소</Label>
              <Select
                value={item.market}
                onValueChange={(v) => updateField("market", v as Market)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="US">US (미국)</SelectItem>
                  <SelectItem value="KR">KR (한국)</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>유형</Label>
              <Select
                value={item.type}
                onValueChange={(v) => updateField("type", v as SecurityType)}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="stock">주식</SelectItem>
                  <SelectItem value="etf">ETF</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="watchlist-sector">섹터</Label>
            <Input
              id="watchlist-sector"
              value={item.sector}
              onChange={(e) => updateField("sector", e.target.value)}
              placeholder="semiconductor, big-tech 등"
            />
          </div>

          <div className="space-y-2">
            <Label>테마</Label>
            <TagInput
              value={item.themes}
              onChange={(themes) => updateField("themes", themes)}
              placeholder="테마 입력 후 Enter"
            />
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            취소
          </Button>
          <Button type="button" onClick={submitItem}>
            {isEditing ? "수정" : "추가"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
