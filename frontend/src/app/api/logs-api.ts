import { ClearLogs, GetLogs } from '@/shared/api/wails/client'
import type { ClearLogsResult, LogEntry } from '@/app/types'

export const logsApi = {
  getLogs: (limit = 300): Promise<LogEntry[]> => GetLogs(limit),
  clearLogs: async (): Promise<ClearLogsResult> => {
    await ClearLogs()
    return {
      memoryCleared: true,
      fileCleared: false,
      pendingRetry: false,
      error: ''
    }
  }
}
