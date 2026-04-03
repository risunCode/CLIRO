import type { AppShellState } from '@/app/services/app-controller'
import type { Account, AuthSession, KiroAuthSession } from '@/features/accounts/types'

export interface AccountsStoreState {
  accounts: Account[]
  busyAccountIds: string[]
  authSession: AuthSession | null
  kiroAuthSession: KiroAuthSession | null
  authWorking: boolean
  refreshingAllQuotas: boolean
}

export const createAccountsStoreState = (shell: AppShellState): AccountsStoreState => {
  return {
    accounts: shell.accounts,
    busyAccountIds: shell.busyAccountIds,
    authSession: shell.authSession,
    kiroAuthSession: shell.kiroAuthSession,
    authWorking: shell.authWorking,
    refreshingAllQuotas: shell.refreshingAllQuotas
  }
}
