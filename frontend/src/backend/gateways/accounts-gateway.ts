import { wailsClient } from '@/backend/client/wails-client'
import type { WailsAuthSyncResult } from '@/backend/models/wails'
import type { Account, AccountAction, AccountSyncResult, QuotaAction, SyncTargetID } from '@/features/accounts/types'

export interface RunAccountActionInput {
  accountId: string
  action: AccountAction
}

export interface RunQuotaActionInput {
  action: QuotaAction
  accountId?: string
}

const toAccountSyncResult = (result: WailsAuthSyncResult): AccountSyncResult => {
  if (result.target !== 'kilo-cli' && result.target !== 'opencode-cli' && result.target !== 'codex-cli') {
    throw new Error(`Unsupported account sync target: ${result.target || 'unknown'}`)
  }

  return result as AccountSyncResult
}

export const accountsApi = {
  getAccounts: (): Promise<Account[]> => wailsClient.accounts.getAccounts(),
  importAccounts: (accounts: Account[]): Promise<number> => wailsClient.accounts.importAccounts(accounts),
  runAccountAction: (input: RunAccountActionInput): Promise<void> =>
    wailsClient.accounts.runAction({
      accountId: input.accountId,
      action: input.action,
    }),
  runQuotaAction: (input: RunQuotaActionInput): Promise<void> =>
    wailsClient.accounts.runQuotaAction({
      action: input.action,
      accountId: input.accountId || '',
    }),
  syncAccountAuth: async (accountId: string, target: SyncTargetID): Promise<AccountSyncResult> =>
    toAccountSyncResult(await wailsClient.accounts.syncAccountAuth(accountId, target)),
}
