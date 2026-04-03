import type { LogEntry } from '@/app/types'
import { subscribeRuntimeEvent } from '@/shared/api/runtime/events'

const appendLogEntryWithLimit = (entries: LogEntry[], entry: LogEntry, limit = 500): LogEntry[] => {
  return [...entries, entry].slice(-limit)
}

export const subscribeToRingLogs = (
  getLogs: () => LogEntry[],
  setLogs: (entries: LogEntry[]) => void,
  limit = 500
): (() => void) => {
  return subscribeRuntimeEvent('log:entry', (payload: unknown) => {
    if (payload && typeof payload === 'object') {
      setLogs(appendLogEntryWithLimit(getLogs(), payload as LogEntry, limit))
    }
  })
}
