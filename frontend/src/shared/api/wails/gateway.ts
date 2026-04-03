import * as client from '@/shared/api/wails/client'
import type { WailsAccount, WailsAppState, WailsCodexAuthSessionView, WailsCodexAuthStart, WailsKiroAuthStart, WailsLogEntry } from '@/shared/api/wails/models'

export const wailsGateway = {
  system: {
    getState: (): Promise<WailsAppState> => client.GetState() as Promise<WailsAppState>,
    getHostName: (): Promise<string> => client.GetHostName(),
    openExternalURL: (url: string): Promise<void> => client.OpenExternalURL(url),
    openDataDir: (): Promise<void> => client.OpenDataDir(),
    confirmQuit: (): Promise<void> => client.ConfirmQuit(),
    hideToTray: (): Promise<void> => client.HideToTray(),
    restoreWindow: (): Promise<void> => client.RestoreWindow()
  },
  logs: {
    getLogs: (limit = 500): Promise<WailsLogEntry[]> => client.GetLogs(limit),
    clearLogs: (): Promise<void> => client.ClearLogs()
  },
  accounts: {
    getAccounts: (): Promise<WailsAccount[]> => client.GetAccounts(),
    importAccounts: (accounts: WailsAccount[]): Promise<number> => client.ImportAccounts(accounts),
    refreshAccount: (accountId: string): Promise<void> => client.RefreshAccount(accountId),
    refreshAccountWithQuota: (accountId: string): Promise<void> => client.RefreshAccountWithQuota(accountId),
    refreshQuota: (accountId: string): Promise<void> => client.RefreshQuota(accountId),
    refreshAllQuotas: (): Promise<void> => client.RefreshAllQuotas(),
    forceRefreshAllQuotas: (): Promise<void> => client.ForceRefreshAllQuotas(),
    toggleAccount: (accountId: string, enabled: boolean): Promise<void> => client.ToggleAccount(accountId, enabled),
    deleteAccount: (accountId: string): Promise<void> => client.DeleteAccount(accountId),
    clearCooldown: (accountId: string): Promise<void> => client.ClearCooldown(accountId)
  },
  auth: {
    startCodexAuth: (): Promise<WailsCodexAuthStart> => client.StartCodexAuth(),
    getCodexAuthSession: (sessionId: string): Promise<WailsCodexAuthSessionView> => client.GetCodexAuthSession(sessionId),
    cancelCodexAuth: (sessionId: string): Promise<void> => client.CancelCodexAuth(sessionId),
    submitCodexAuthCode: (sessionId: string, code: string): Promise<void> => client.SubmitCodexAuthCode(sessionId, code),
    startKiroAuth: (): Promise<WailsKiroAuthStart> => client.StartKiroAuth(),
    startKiroSocialAuth: (provider: string): Promise<WailsKiroAuthStart> => client.StartKiroSocialAuth(provider),
    getKiroAuthSession: (sessionId: string): Promise<unknown> => client.GetKiroAuthSession(sessionId),
    cancelKiroAuth: (sessionId: string): Promise<void> => client.CancelKiroAuth(sessionId),
    submitKiroAuthCode: (sessionId: string, code: string): Promise<void> => client.SubmitKiroAuthCode(sessionId, code)
  }
}
