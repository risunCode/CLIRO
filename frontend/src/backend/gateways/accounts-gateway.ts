import { ClearCooldown, DeleteAccount, ForceRefreshAllQuotas, GetAccounts, ImportAccounts, RefreshAccount, RefreshAccountWithQuota, RefreshAllQuotas, RefreshQuota, SyncCodexAccountToCodexCLI, SyncCodexAccountToKiloAuth, SyncCodexAccountToOpencodeAuth, ToggleAccount } from '@/backend/client/wails-client'
import { toCodexAuthSyncResult, toKiloAuthSyncResult, toOpencodeAuthSyncResult } from '@/backend/compat/accounts-compat'
import type { Account, CodexAuthSyncResult, KiloAuthSyncResult, OpencodeAuthSyncResult } from '@/features/accounts/types'

export const accountsApi = {
  getAccounts: (): Promise<Account[]> => GetAccounts(),
  importAccounts: (accounts: Account[]): Promise<number> => ImportAccounts(accounts),
  refreshAccount: (accountId: string): Promise<void> => RefreshAccount(accountId),
  refreshAccountWithQuota: (accountId: string): Promise<void> => RefreshAccountWithQuota(accountId),
  refreshQuota: (accountId: string): Promise<void> => RefreshQuota(accountId),
  refreshAllQuotas: (): Promise<void> => RefreshAllQuotas(),
  forceRefreshAllQuotas: (): Promise<void> => ForceRefreshAllQuotas(),
  toggleAccount: (accountId: string, enabled: boolean): Promise<void> => ToggleAccount(accountId, enabled),
  deleteAccount: (accountId: string): Promise<void> => DeleteAccount(accountId),
  clearCooldown: (accountId: string): Promise<void> => ClearCooldown(accountId),
  syncCodexAccountToKiloAuth: async (accountId: string): Promise<KiloAuthSyncResult> => toKiloAuthSyncResult(await SyncCodexAccountToKiloAuth(accountId)),
  syncCodexAccountToOpencodeAuth: async (accountId: string): Promise<OpencodeAuthSyncResult> => toOpencodeAuthSyncResult(await SyncCodexAccountToOpencodeAuth(accountId)),
  syncCodexAccountToCodexCLI: async (accountId: string): Promise<CodexAuthSyncResult> => toCodexAuthSyncResult(await SyncCodexAccountToCodexCLI(accountId))
}
