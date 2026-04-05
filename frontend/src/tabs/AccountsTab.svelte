<script lang="ts">
  import type { AppActions, AppShellState, AccountsActions } from '@/app/services/app-controller'
  import AccountsScreen from '@/features/accounts/components/AccountsScreen.svelte'

  export let shell: AppShellState
  export let appActions: AppActions
  export let accountsActions: AccountsActions
</script>

<AccountsScreen {shell} {appActions} {accountsActions} />
