import type { WailsAppState, WailsLogEntry } from '@/backend/models/wails'

export type AppState = WailsAppState & {
  startupWarnings?: Array<{
    code?: string
    filePath?: string
    backupPath?: string
    message?: string
  }>
}

export type LogEntry = WailsLogEntry

export type SystemAction = 'confirm-quit' | 'hide-to-tray' | 'restore-window' | 'open-data-dir' | 'clear-logs'

export interface UpdateInfo {
  currentVersion: string
  latestVersion: string
  releaseName: string
  releaseUrl: string
  publishedAt: string
  updateAvailable: boolean
}
