<script lang="ts">
  import AccountsWorkspace from '@/features/accounts/components/AccountsWorkspace.svelte'
  import type {
    Account,
    AuthSession,
    CodexAuthSyncResult,
    KiloAuthSyncResult,
    KiroAuthSession,
    OpencodeAuthSyncResult
  } from '@/features/accounts/types'

  export let accounts: Account[] = []
  export let busyAccountIds: string[] = []
  export let authSession: AuthSession | null = null
  export let kiroAuthSession: KiroAuthSession | null = null
  export let authWorking = false
  export let refreshingAllQuotas = false
  export let onStartConnect: () => Promise<void>
  export let onCancelConnect: () => Promise<void>
  export let onStartKiroConnect: (method: 'device' | 'google' | 'github') => Promise<void>
  export let onCancelKiroConnect: () => Promise<void>
  export let onOpenExternalURL: (url: string) => Promise<void>
  export let onRefreshAllQuotas: () => Promise<void>
  export let onForceRefreshAllQuotas: () => Promise<void>
  export let onToggleAccount: (accountId: string, enabled: boolean) => Promise<void>
  export let onBulkToggleAccounts: (accountIds: string[], enabled: boolean) => Promise<void>
  export let onBulkDeleteAccounts: (accountIds: string[]) => Promise<void>
  export let onImportAccounts: (accounts: Account[]) => Promise<number>
  export let onSyncCodexAccountToKiloAuth: (accountId: string) => Promise<KiloAuthSyncResult>
  export let onSyncCodexAccountToOpencodeAuth: (accountId: string) => Promise<OpencodeAuthSyncResult>
  export let onSyncCodexAccountToCodexCLI: (accountId: string) => Promise<CodexAuthSyncResult>
  export let onRefreshAccountWithQuota: (accountId: string) => Promise<void>
  export let onDeleteAccount: (accountId: string) => Promise<void>
</script>

<AccountsWorkspace
  {accounts}
  {busyAccountIds}
  {authSession}
  {kiroAuthSession}
  {authWorking}
  {refreshingAllQuotas}
  {onStartConnect}
  {onCancelConnect}
  {onStartKiroConnect}
  {onCancelKiroConnect}
  {onOpenExternalURL}
  {onRefreshAllQuotas}
  {onForceRefreshAllQuotas}
  {onToggleAccount}
  {onBulkToggleAccounts}
  {onBulkDeleteAccounts}
  {onImportAccounts}
  {onSyncCodexAccountToKiloAuth}
  {onSyncCodexAccountToOpencodeAuth}
  {onSyncCodexAccountToCodexCLI}
  {onRefreshAccountWithQuota}
  {onDeleteAccount}
/>
