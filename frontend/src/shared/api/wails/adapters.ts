export const asString = (value: unknown): string => {
  return typeof value === 'string' ? value : String(value ?? '')
}

export const asBoolean = (value: unknown): boolean => {
  return Boolean(value)
}

export const asNumber = (value: unknown): number => {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }

  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

export const asStringArray = (value: unknown): string[] => {
  if (!Array.isArray(value)) {
    return []
  }
  return value.map((item) => String(item))
}

export const asNumberArray = (value: unknown): number[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value
    .map((item) => {
      if (typeof item === 'number' && Number.isFinite(item)) {
        return item
      }
      const parsed = Number(item)
      return Number.isFinite(parsed) ? parsed : 0
    })
    .filter((item) => item > 0)
}

export const asRecord = (value: unknown): Record<string, unknown> => {
  if (typeof value === 'object' && value !== null) {
    return value as Record<string, unknown>
  }
  return {}
}

export const pick = (payload: Record<string, unknown>, camelKey: string, snakeKey: string): unknown => {
  return payload[camelKey] ?? payload[snakeKey]
}
