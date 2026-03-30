import { GetHostName, GetState, OpenDataDir, OpenExternalURL } from '@/shared/api/wails/client'
import type { AppState, UpdateInfo } from '@/app/types'

export const systemApi = {
  getState: (): Promise<AppState> => GetState() as Promise<AppState>,
  getUpdateInfo: async (): Promise<UpdateInfo> => ({
    currentVersion: '',
    latestVersion: '',
    releaseName: '',
    releaseUrl: '',
    publishedAt: '',
    updateAvailable: false
  }),
  getHostName: async (): Promise<string> => String(await GetHostName()),
  openExternalURL: (url: string): Promise<void> => OpenExternalURL(url),
  openDataDir: (): Promise<void> => OpenDataDir()
}
