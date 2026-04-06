import { wailsClient } from '@/backend/client/wails-client'
import type { LogEntry } from '@/backend/models/system'

export const logsApi = {
  getLogs: wailsClient.logs.getLogs as (limit?: number) => Promise<LogEntry[]>,
  clearLogs: (): Promise<void> => wailsClient.system.runAction({ action: 'clear-logs' }),
}
