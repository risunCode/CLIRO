<script lang="ts">
  import { onMount } from 'svelte'
  import CliSyncPanel from '@/features/router/components/CliSyncPanel.svelte'
  import CloudflaredPanel from '@/features/router/components/CloudflaredPanel.svelte'
  import EndpointTesterPanel from '@/features/router/components/EndpointTesterPanel.svelte'
  import ModelAliasPanel from '@/features/router/components/ModelAliasPanel.svelte'
  import ProxyControlsPanel from '@/features/router/components/ProxyControlsPanel.svelte'
  import SchedulingPanel from '@/features/router/components/SchedulingPanel.svelte'
  import type { CliSyncAppID, CliSyncResult, CliSyncStatus, LocalModelCatalogItem, ProxyStatus } from '@/features/router/types'

  export let proxyStatus: ProxyStatus | null = null
  export let busy = false
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

  onMount(() => {
    void onRefreshProxyStatus().catch(() => {})
  })
</script>

<div class="space-y-4">
  <ProxyControlsPanel
    {proxyStatus}
    {busy}
    {onStartProxy}
    {onStopProxy}
    {onSetProxyPort}
    {onSetAllowLAN}
    {onSetAutoStartProxy}
    {onSetProxyAPIKey}
    {onRegenerateProxyAPIKey}
    {onSetAuthorizationMode}
  />

  <CliSyncPanel
    {busy}
    proxyBaseURL={proxyStatus?.url || ''}
    proxyAPIKey={proxyStatus?.proxyApiKey || ''}
    {onGetCLISyncStatuses}
    {onGetLocalModelCatalog}
    {onGetCLISyncFileContent}
    {onSaveCLISyncFileContent}
    {onSyncCLIConfig}
  />

  <EndpointTesterPanel proxyStatus={proxyStatus} apiKey={proxyStatus?.proxyApiKey || ''} />

  <ModelAliasPanel
    {busy}
    {onGetModelAliases}
    {onSetModelAliases}
  />

  <SchedulingPanel
    {proxyStatus}
    {busy}
    {onSetSchedulingMode}
    {onSetCircuitBreaker}
    {onSetCircuitSteps}
  />

  <CloudflaredPanel
    {proxyStatus}
    {busy}
    {onRefreshCloudflaredStatus}
    {onSetCloudflaredConfig}
    {onInstallCloudflared}
    {onStartCloudflared}
    {onStopCloudflared}
  />
</div>
