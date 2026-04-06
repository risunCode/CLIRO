import { wailsClient } from '@/backend/client/wails-client'
import type { AppState, SystemAction, UpdateInfo } from '@/backend/models/system'

export const systemApi = {
  getState: (): Promise<AppState> => wailsClient.system.getState() as Promise<AppState>,
  getHostName: wailsClient.system.getHostName,
  runAction: (action: SystemAction): Promise<void> => wailsClient.system.runAction({ action }),
  openExternalURL: wailsClient.system.openExternalURL,
  getUpdateInfo: async (): Promise<UpdateInfo | null> => null,
}
