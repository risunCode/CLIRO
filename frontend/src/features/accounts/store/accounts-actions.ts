import type { AppActions, AccountsActions } from '@/app/services/app-controller'
import { accountsAuthApi } from '@/backend/gateways/auth-gateway'

export interface AccountsScreenActions extends AccountsActions {
  openExternalURL: (url: string) => Promise<void>
  submitCodexAuthCode: (sessionId: string, code: string) => Promise<void>
}

export const createAccountsScreenActions = (actions: {
  appActions: AppActions
  accountsActions: AccountsActions
}): AccountsScreenActions => {
  return {
    ...actions.accountsActions,
    openExternalURL: actions.appActions.openExternalURL,
    submitCodexAuthCode: (sessionId: string, code: string) => accountsAuthApi.submitCodexAuthCode(sessionId, code)
  }
}
