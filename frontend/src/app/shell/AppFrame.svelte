<script lang="ts">
  import AppHeader from '@/components/common/AppHeader.svelte'
  import AppFooter from '@/components/common/AppFooter.svelte'
  import DashboardTab from '@/tabs/DashboardTab.svelte'
  import AccountsTab from '@/tabs/AccountsTab.svelte'
  import ApiRouterTab from '@/tabs/ApiRouterTab.svelte'
  import UsageTab from '@/tabs/UsageTab.svelte'
  import SystemLogsTab from '@/tabs/SystemLogsTab.svelte'
  import SettingsTab from '@/tabs/SettingsTab.svelte'
  import type { Theme } from '@/shared/stores/theme'
  import type { AppState, LogEntry } from '@/app/types'
  import type {
    Account,
    AuthSession,
    CodexAuthSyncResult,
    KiroAuthSession,
    KiloAuthSyncResult,
    OpencodeAuthSyncResult
  } from '@/features/accounts/types'
  import type { CliSyncAppID, CliSyncResult, CliSyncStatus, LocalModelCatalogItem, ProxyStatus } from '@/features/router/types'
  import type { AppTabId } from '@/app/lib/tabs'

  interface BackupPayload {
    version: number
    exportedAt: string
    state: AppState | null
    accounts: Account[]
  }

  interface RestoreProgress {
    step: string
    index: number
    total: number
  }

  export let activeTab: AppTabId = 'dashboard'
  export let theme: Theme = 'light'
  export let state: AppState | null = null
  export let accounts: Account[] = []
  export let proxyStatus: ProxyStatus | null = null
  export let logs: LogEntry[] = []
  export let loadingDashboard = false
  export let loadingLogs = false
  export let clearingLogs = false
  export let proxyBusy = false
  export let authWorking = false
  export let refreshingAllQuotas = false
  export let busyAccountIds: string[] = []
  export let authSession: AuthSession | null = null
  export let kiroAuthSession: KiroAuthSession | null = null

  export let onToggleTheme: () => void
  export let onTabChange: (tabId: AppTabId) => void
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
  export let onStartProxy: () => Promise<void>
  export let onStopProxy: () => Promise<void>
  export let onSetProxyPort: (port: number) => Promise<void>
  export let onSetAllowLAN: (enabled: boolean) => Promise<void>
  export let onSetAutoStartProxy: (enabled: boolean) => Promise<void>
  export let onSetProxyAPIKey: (apiKey: string) => Promise<void>
  export let onRegenerateProxyAPIKey: () => Promise<string>
  export let onSetAuthorizationMode: (enabled: boolean) => Promise<void>
  export let onSetSchedulingMode: (mode: string) => Promise<void>
  export let onSetCircuitBreaker: (enabled: boolean) => Promise<void>
  export let onSetCircuitSteps: (steps: number[]) => Promise<void>
  export let onRefreshProxyStatus: () => Promise<void>
  export let onRefreshCloudflaredStatus: () => Promise<void>
  export let onSetCloudflaredConfig: (mode: string, token: string, useHttp2: boolean) => Promise<void>
  export let onInstallCloudflared: () => Promise<void>
  export let onStartCloudflared: () => Promise<void>
  export let onStopCloudflared: () => Promise<void>
  export let onGetCLISyncStatuses: () => Promise<CliSyncStatus[]>
  export let onGetLocalModelCatalog: () => Promise<LocalModelCatalogItem[]>
  export let onGetCLISyncFileContent: (appId: CliSyncAppID, path: string) => Promise<string>
  export let onSaveCLISyncFileContent: (appId: CliSyncAppID, path: string, content: string) => Promise<void>
  export let onSyncCLIConfig: (appId: CliSyncAppID, model: string) => Promise<CliSyncResult>
  export let onGetModelAliases: () => Promise<Record<string, string>>
  export let onSetModelAliases: (aliases: Record<string, string>) => Promise<void>
  export let onRefreshLogs: (limit?: number) => Promise<void>
  export let onClearLogs: () => Promise<void>
  export let onOpenDataDir: () => Promise<void>
  export let onExportBackup: () => Promise<void>
  export let onRestoreBackup: (payload: BackupPayload, onProgress?: (progress: RestoreProgress) => void) => Promise<void>

  const handleTabChange = (event: CustomEvent<AppTabId>): void => {
    onTabChange(event.detail)
  }
</script>

<div class="flex h-full flex-col">
  <AppHeader activeTab={activeTab} on:tabChange={handleTabChange} on:toggleTheme={onToggleTheme} {theme} />

  <section class="no-scrollbar min-h-0 flex-1 overflow-y-auto px-4 py-4 md:px-6">
    <div class="space-y-4 pb-1">
      {#if activeTab === 'dashboard'}
        <DashboardTab {state} {accounts} {proxyStatus} loading={loadingDashboard} />
      {:else if activeTab === 'accounts'}
        <AccountsTab
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
      {:else if activeTab === 'api-router'}
        <ApiRouterTab
          {proxyStatus}
          busy={proxyBusy}
          {onStartProxy}
          {onStopProxy}
          {onSetProxyPort}
          {onSetAllowLAN}
          {onSetAutoStartProxy}
          {onSetProxyAPIKey}
          {onRegenerateProxyAPIKey}
          {onSetAuthorizationMode}
          {onSetSchedulingMode}
          {onSetCircuitBreaker}
          {onSetCircuitSteps}
          {onRefreshProxyStatus}
          {onRefreshCloudflaredStatus}
          {onSetCloudflaredConfig}
          {onInstallCloudflared}
          {onStartCloudflared}
          {onStopCloudflared}
          {onGetCLISyncStatuses}
          {onGetLocalModelCatalog}
          {onGetCLISyncFileContent}
          {onSaveCLISyncFileContent}
          {onSyncCLIConfig}
          {onGetModelAliases}
          {onSetModelAliases}
        />
      {:else if activeTab === 'usage'}
        <UsageTab {state} {accounts} {proxyStatus} {logs} />
      {:else if activeTab === 'system-logs'}
        <SystemLogsTab logs={logs} loading={loadingLogs} clearing={clearingLogs} onRefreshLogs={onRefreshLogs} onClearLogs={onClearLogs} />
      {:else if activeTab === 'settings'}
        <SettingsTab
          {onOpenDataDir}
          onExportBackup={onExportBackup}
          onRestoreBackup={onRestoreBackup}
        />
      {/if}
    </div>
  </section>

  <AppFooter {proxyStatus} />
</div>
