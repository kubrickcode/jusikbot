"use client"

import * as React from "react"
import { X } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

interface TagInputProps {
  value: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

function TagInput({
  value,
  onChange,
  placeholder = "태그 입력 후 Enter",
  className,
  disabled = false,
}: TagInputProps) {
  const [inputValue, setInputValue] = React.useState("")
  const inputRef = React.useRef<HTMLInputElement>(null)

  function addTags(raw: string) {
    const candidates = raw
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0)

    const unique = candidates.filter(
      (tag) => !value.includes(tag)
    )

    if (unique.length > 0) {
      onChange([...value, ...unique])
    }
  }

  function removeTag(target: string) {
    onChange(value.filter((tag) => tag !== target))
  }

  function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      event.preventDefault()
      if (inputValue.trim()) {
        addTags(inputValue)
        setInputValue("")
      }
    }

    if (
      event.key === "Backspace" &&
      inputValue === "" &&
      value.length > 0
    ) {
      removeTag(value[value.length - 1])
    }
  }

  function handlePaste(event: React.ClipboardEvent<HTMLInputElement>) {
    event.preventDefault()
    const pasted = event.clipboardData.getData("text/plain")
    addTags(pasted)
    setInputValue("")
  }

  return (
    <div
      className={cn(
        "border-input focus-within:border-ring focus-within:ring-ring/50 flex min-h-9 w-full flex-wrap items-center gap-1.5 rounded-md border bg-transparent px-3 py-1.5 shadow-xs transition-[color,box-shadow] focus-within:ring-[3px]",
        disabled && "pointer-events-none cursor-not-allowed opacity-50",
        className
      )}
      onClick={() => inputRef.current?.focus()}
    >
      {value.map((tag) => (
        <Badge
          key={tag}
          variant="secondary"
          className="gap-1 pr-1"
        >
          {tag}
          <button
            type="button"
            aria-label={`${tag} 삭제`}
            className="rounded-xs hover:bg-muted-foreground/20 inline-flex size-4 items-center justify-center"
            onClick={(event) => {
              event.stopPropagation()
              removeTag(tag)
            }}
            disabled={disabled}
          >
            <X className="size-3" />
          </button>
        </Badge>
      ))}
      <Input
        ref={inputRef}
        value={inputValue}
        onChange={(event) => setInputValue(event.target.value)}
        onKeyDown={handleKeyDown}
        onPaste={handlePaste}
        placeholder={value.length === 0 ? placeholder : undefined}
        disabled={disabled}
        className="h-auto min-w-[80px] flex-1 border-0 p-0 shadow-none focus-visible:ring-0"
      />
    </div>
  )
}

export { TagInput }
export type { TagInputProps }
