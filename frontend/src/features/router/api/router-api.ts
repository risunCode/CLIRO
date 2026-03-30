import { GetCLISyncFileContent, GetCLISyncStatuses, GetLocalModelCatalog, GetModelAliases, GetProxyStatus, InstallCloudflared, RefreshCloudflaredStatus, RegenerateProxyAPIKey, SaveCLISyncFileContent, SetAllowLAN, SetAuthorizationMode, SetAutoStartProxy, SetCircuitBreaker, SetCircuitSteps, SetCloudflaredConfig, SetModelAliases, SetProxyAPIKey, SetProxyPort, SetSchedulingMode, StartCloudflared, StartProxy, StopCloudflared, StopProxy, SyncCLIConfig } from '@/shared/api/wails/client'
import { toCliSyncResult, toCliSyncStatus, toLocalModelCatalogItem, toProxyStatus } from '@/features/router/api/adapters'
import type { CliSyncAppID, CliSyncResult, CliSyncStatus, LocalModelCatalogItem, ProxyStatus } from '@/features/router/types'

export const routerApi = {
  getProxyStatus: async (): Promise<ProxyStatus> => toProxyStatus(await GetProxyStatus()),
  refreshCloudflaredStatus: async (): Promise<ProxyStatus> => toProxyStatus(await RefreshCloudflaredStatus()),
  getLocalModelCatalog: async (): Promise<LocalModelCatalogItem[]> => (await GetLocalModelCatalog()).map(toLocalModelCatalogItem),
  getCliSyncStatuses: async (): Promise<CliSyncStatus[]> =>
    (await GetCLISyncStatuses())
      .map(toCliSyncStatus)
      .filter((status): status is CliSyncStatus => status !== null),
  getCliSyncFileContent: (appId: CliSyncAppID, path: string): Promise<string> => GetCLISyncFileContent(appId, path),
  saveCliSyncFileContent: (appId: CliSyncAppID, path: string, content: string): Promise<void> => SaveCLISyncFileContent(appId, path, content),
  syncCLIConfig: async (appId: CliSyncAppID, model: string): Promise<CliSyncResult> => toCliSyncResult(await SyncCLIConfig(appId, model)),
  startProxy: (): Promise<void> => StartProxy(),
  stopProxy: (): Promise<void> => StopProxy(),
  setProxyPort: (port: number): Promise<void> => SetProxyPort(port),
  setAllowLAN: (enabled: boolean): Promise<void> => SetAllowLAN(enabled),
  setAutoStartProxy: (enabled: boolean): Promise<void> => SetAutoStartProxy(enabled),
  setProxyAPIKey: (apiKey: string): Promise<void> => SetProxyAPIKey(apiKey),
  regenerateProxyAPIKey: (): Promise<string> => RegenerateProxyAPIKey(),
  setAuthorizationMode: (enabled: boolean): Promise<void> => SetAuthorizationMode(enabled),
  setSchedulingMode: (mode: string): Promise<void> => SetSchedulingMode(mode),
  setCircuitBreaker: (enabled: boolean): Promise<void> => SetCircuitBreaker(enabled),
  setCircuitSteps: (steps: number[]): Promise<void> => SetCircuitSteps(steps),
  setCloudflaredConfig: (mode: string, token: string, useHttp2: boolean): Promise<void> => SetCloudflaredConfig(mode, token, useHttp2),
  installCloudflared: (): Promise<void> => InstallCloudflared(),
  startCloudflared: (): Promise<void> => StartCloudflared(),
  stopCloudflared: (): Promise<void> => StopCloudflared(),
  getModelAliases: (): Promise<Record<string, string>> => GetModelAliases(),
  setModelAliases: (aliases: Record<string, string>): Promise<void> => SetModelAliases(aliases)
}
