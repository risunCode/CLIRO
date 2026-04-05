import { wailsClient } from '@/backend/client/wails-client'
import type { AppState, SystemAction, UpdateInfo } from '@/backend/models/system'

export const systemApi = {
  getState: (): Promise<AppState> => wailsClient.system.getState() as Promise<AppState>,
  getUpdateInfo: async (): Promise<UpdateInfo | null> => null,
  getHostName: async (): Promise<string> => String(await wailsClient.system.getHostName()),
  runAction: (action: SystemAction): Promise<void> => wailsClient.system.runAction({ action }),
  openExternalURL: (url: string): Promise<void> => wailsClient.system.openExternalURL(url),
  confirmQuit: (): Promise<void> => wailsClient.system.runAction({ action: 'confirm-quit' }),
  hideToTray: (): Promise<void> => wailsClient.system.runAction({ action: 'hide-to-tray' }),
  restoreWindow: (): Promise<void> => wailsClient.system.runAction({ action: 'restore-window' }),
  openDataDir: (): Promise<void> => wailsClient.system.runAction({ action: 'open-data-dir' }),
}
