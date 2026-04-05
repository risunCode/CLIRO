import type { LogEntry } from '@/app/types'
import type { Account } from '@/features/accounts/types'

export type UsageProvider = 'codex' | 'kiro'

export interface ParsedRequestLog {
  requestId: string
  timestamp: number
  provider: UsageProvider
  model: string
  account: string
  promptTokens: number
  completionTokens: number
  totalTokens: number
}

const ACTIVE_WINDOW_MS = 5000

const parseLogMessage = (message: string): Record<string, string> => {
  const out: Record<string, string> = {}
  const matcher = /(\w+)=((?:"[^"]*")|\S+)/g
  let match: RegExpExecArray | null = null
  while ((match = matcher.exec(message)) !== null) {
    const key = match[1]
    const rawValue = match[2]
    out[key] = rawValue.startsWith('"') && rawValue.endsWith('"') ? rawValue.slice(1, -1) : rawValue
  }
  return out
}

export const normalizeProvider = (value: string): UsageProvider | '' => {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'kiro') {
    return 'kiro'
  }
  if (normalized === 'codex' || normalized === 'chatgpt') {
    return 'codex'
  }
  return ''
}

const getStringField = (entry: LogEntry, key: string): string => {
  const fields = entry.fields as Record<string, unknown> | undefined
  const fieldValue = fields?.[key]
  if (typeof fieldValue === 'string') {
    return fieldValue.trim()
  }
  if (typeof fieldValue === 'number' || typeof fieldValue === 'boolean' || typeof fieldValue === 'bigint') {
    return String(fieldValue)
  }

  const parsed = parseLogMessage(entry.message || '')
  return (parsed[key] || '').trim()
}

const getNumberField = (entry: LogEntry, key: string): number => {
  const fields = entry.fields as Record<string, unknown> | undefined
  const fieldValue = fields?.[key]
  if (typeof fieldValue === 'number' && Number.isFinite(fieldValue)) {
    return fieldValue
  }
  if (typeof fieldValue === 'string' && fieldValue.trim() !== '') {
    const parsedValue = Number(fieldValue)
    if (Number.isFinite(parsedValue)) {
      return parsedValue
    }
  }

  const parsed = parseLogMessage(entry.message || '')
  const parsedValue = Number(parsed[key] || 0)
  return Number.isFinite(parsedValue) ? parsedValue : 0
}

const isUsageCompletionEvent = (entry: LogEntry): boolean => {
  if (entry.scope !== 'proxy') {
    return false
  }

  const eventName = (entry.event || '').trim()
  if (eventName === 'request.provider_completed') {
    return true
  }
  if (eventName === 'request.success') {
    return true
  }
  if (eventName === 'request.completed' && getStringField(entry, 'provider') !== '') {
    return true
  }

  return getStringField(entry, 'phase') === 'provider_completed'
}

const requestLogQuality = (entry: ParsedRequestLog): number => {
  let score = 0
  if (entry.model !== '-') {
    score += 4
  }
  if (entry.account !== '-') {
    score += 2
  }
  if (entry.totalTokens > 0) {
    score += 1
  }
  return score
}

export const toRequestLog = (entry: LogEntry): ParsedRequestLog | null => {
  if (!isUsageCompletionEvent(entry)) {
    return null
  }

  const provider = normalizeProvider(getStringField(entry, 'provider'))
  if (!provider) {
    return null
  }

  return {
    requestId: (entry.requestId || getStringField(entry, 'request_id')).trim(),
    timestamp: Number(entry.timestamp || 0),
    provider,
    model: getStringField(entry, 'model') || '-',
    account: getStringField(entry, 'account') || '-',
    promptTokens: getNumberField(entry, 'prompt_tokens') || getNumberField(entry, 'input_tokens'),
    completionTokens: getNumberField(entry, 'completion_tokens') || getNumberField(entry, 'output_tokens'),
    totalTokens: getNumberField(entry, 'total_tokens')
  }
}

export const getRequestLogs = (logs: LogEntry[]): ParsedRequestLog[] => {
  const parsedLogs = logs.map(toRequestLog).filter((item): item is ParsedRequestLog => item !== null)
  const deduped: ParsedRequestLog[] = []
  const indexByKey = new Map<string, number>()

  for (const item of parsedLogs) {
    const key = item.requestId || `${item.timestamp}:${item.provider}:${item.account}:${item.totalTokens}`
    const existingIndex = indexByKey.get(key)
    if (existingIndex === undefined) {
      indexByKey.set(key, deduped.length)
      deduped.push(item)
      continue
    }

    const existing = deduped[existingIndex]
    if (requestLogQuality(item) >= requestLogQuality(existing)) {
      deduped[existingIndex] = item
    }
  }

  return deduped
}

export const getRecentRequests = (requestLogs: ParsedRequestLog[], limit = 10): ParsedRequestLog[] => [...requestLogs].reverse().slice(0, limit)

export const getLastActiveAt = (requestLogs: ParsedRequestLog[], provider: UsageProvider): number => requestLogs.find((item) => item.provider === provider)?.timestamp || 0

export const isProviderActive = (proxyOnline: boolean, lastActiveAt: number, now: number): boolean => proxyOnline && now - lastActiveAt < ACTIVE_WINDOW_MS

export const getEnabledAccountCount = (accounts: Account[], provider: UsageProvider): number => accounts.filter((account) => account.enabled && normalizeProvider(account.provider || '') === provider).length

export const getProviderRequestCount = (requestLogs: ParsedRequestLog[], provider: UsageProvider): number => requestLogs.filter((item) => item.provider === provider).length

export const formatRelativeTime = (timestamp: number, now: number): string => {
  if (!timestamp) {
    return '-'
  }
  const deltaMs = now - timestamp
  const seconds = Math.max(Math.floor(deltaMs / 1000), 0)
  if (seconds < 5) {
    return 'now'
  }
  if (seconds < 60) {
    return `${seconds}s ago`
  }
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) {
    return `${minutes}m ago`
  }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) {
    return `${hours}h ago`
  }
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export const providerLabel = (provider: UsageProvider): string => (provider === 'kiro' ? 'Kiro' : 'Codex')
