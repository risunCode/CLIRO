<script lang="ts">
  import type { AppActions, AppShellState, AccountsActions } from '@/app/services/app-controller'
  import AccountsWorkspace from '@/features/accounts/components/AccountsWorkspace.svelte'
  import { createAccountsScreenActions } from '@/features/accounts/store/accounts-actions'
  import { createAccountsStoreState } from '@/features/accounts/store/accounts-store'

  export let shell: AppShellState
  export let appActions: AppActions
  export let accountsActions: AccountsActions

  $: store = createAccountsStoreState(shell)
  $: actions = createAccountsScreenActions({ appActions, accountsActions })
</script>

<AccountsWorkspace
  accounts={store.accounts}
  busyAccountIds={store.busyAccountIds}
  authSession={store.authSession}
  kiroAuthSession={store.kiroAuthSession}
  authWorking={store.authWorking}
  refreshingAllQuotas={store.refreshingAllQuotas}
  onStartConnect={actions.startConnect}
  onCancelConnect={actions.cancelConnect}
  onStartKiroConnect={actions.startKiroConnect}
  onCancelKiroConnect={actions.cancelKiroConnect}
  onOpenExternalURL={actions.openExternalURL}
  onSubmitCodexAuthCode={actions.submitCodexAuthCode}
  onRefreshAllQuotas={actions.refreshAllQuotas}
  onForceRefreshAllQuotas={actions.forceRefreshAllQuotas}
  onToggleAccount={actions.toggleAccount}
  onBulkToggleAccounts={actions.bulkToggleAccounts}
  onBulkDeleteAccounts={actions.bulkDeleteAccounts}
  onImportAccounts={actions.importAccounts}
  onSyncAccountAuth={actions.syncAccountAuth}
  onRefreshAccountWithQuota={actions.refreshAccountWithQuota}
  onDeleteAccount={actions.deleteAccount}
/>
