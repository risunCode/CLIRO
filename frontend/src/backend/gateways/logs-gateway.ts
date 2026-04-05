import { ClearLogs, GetLogs } from '@/backend/client/wails-client'
import type { LogEntry } from '@/backend/models/system'

export const logsApi = {
  getLogs: (limit = 500): Promise<LogEntry[]> => GetLogs(limit),
  clearLogs: (): Promise<void> => ClearLogs()
}
