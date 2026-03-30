import type { AppState } from '@/app/types'

export interface StartupWarningEntry {
  code: string
  filePath: string
  backupPath?: string
  message: string
}

export const mapStartupWarnings = (nextState: AppState | null): StartupWarningEntry[] => {
  return Array.isArray(nextState?.startupWarnings)
    ? nextState.startupWarnings.map((warning) => ({
        code: String(warning.code || 'startup_warning'),
        filePath: String(warning.filePath || '-'),
        backupPath: warning.backupPath ? String(warning.backupPath) : undefined,
        message: String(warning.message || 'Configuration was recovered during startup.')
      }))
    : []
}
