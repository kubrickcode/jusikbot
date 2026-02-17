export function isKrSymbol(symbol: string): boolean {
  return /^\d+$/.test(symbol)
}

export function formatPrice(value: number, symbol: string): string {
  if (isKrSymbol(symbol)) {
    return Math.round(value).toLocaleString("ko-KR")
  }
  return value.toFixed(2)
}

export function formatVolume(value: number): string {
  return value.toLocaleString("ko-KR")
}

export function formatRate(value: number): string {
  return value.toLocaleString("ko-KR", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

export function daysAgo(days: number): string {
  const date = new Date()
  date.setDate(date.getDate() - days)
  return date.toISOString().slice(0, 10)
}

export function today(): string {
  return new Date().toISOString().slice(0, 10)
}
