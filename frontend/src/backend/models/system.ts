import type { logger, main } from '../../../wailsjs/go/models'

export type AppState = main.State

export type LogEntry = logger.Entry

export type SystemAction = 'confirm-quit' | 'hide-to-tray' | 'restore-window' | 'open-data-dir' | 'clear-logs'

export interface UpdateInfo {
  currentVersion: string
  latestVersion: string
  releaseName: string
  releaseUrl: string
  publishedAt: string
  updateAvailable: boolean
}
