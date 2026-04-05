import { wailsClient } from '@/backend/client/wails-client'
import type { LogEntry } from '@/backend/models/system'

export const logsApi = {
  getLogs: (limit = 500): Promise<LogEntry[]> => wailsClient.logs.getLogs(limit),
  clearLogs: (): Promise<void> => wailsClient.system.runAction({ action: 'clear-logs' }),
}
