import { ConfirmQuit, GetHostName, GetState, HideToTray, OpenDataDir, OpenExternalURL, RestoreWindow } from '@/backend/client/wails-client'
import type { AppState, UpdateInfo } from '@/backend/models/system'

export const systemApi = {
  getState: (): Promise<AppState> => GetState() as Promise<AppState>,
  getUpdateInfo: async (): Promise<UpdateInfo | null> => null,
  getHostName: async (): Promise<string> => String(await GetHostName()),
  openExternalURL: (url: string): Promise<void> => OpenExternalURL(url),
  openDataDir: (): Promise<void> => OpenDataDir(),
  confirmQuit: (): Promise<void> => ConfirmQuit(),
  hideToTray: (): Promise<void> => HideToTray(),
  restoreWindow: (): Promise<void> => RestoreWindow()
}
