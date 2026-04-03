<script lang="ts">
  import type { RouterActions } from '@/app/services/app-controller'
  import CliSyncPanel from '@/features/router/components/cli-sync/CliSyncPanel.svelte'
  import CloudflaredPanel from '@/features/router/components/cloudflared/CloudflaredPanel.svelte'
  import EndpointTesterPanel from '@/features/router/components/endpoint-tester/EndpointTesterPanel.svelte'
  import ModelAliasPanel from '@/features/router/components/model-alias/ModelAliasPanel.svelte'
  import ProxyControlsPanel from '@/features/router/components/proxy/ProxyControlsPanel.svelte'
  import SchedulingPanel from '@/features/router/components/scheduling/SchedulingPanel.svelte'
  import { createRouterPanelActions } from '@/features/router/store/router-actions'
  import { createRouterStoreState } from '@/features/router/store/router-store'
  import type { ProxyStatus } from '@/features/router/types'

  export let proxyStatus: ProxyStatus | null = null
  export let busy = false
  export let routerActions: RouterActions

  $: store = createRouterStoreState({ proxyStatus, busy })
  $: actions = createRouterPanelActions(routerActions)
</script>

<div class="space-y-4">
  <ProxyControlsPanel
    proxyStatus={store.proxyStatus}
    busy={store.busy}
    onStartProxy={actions.startProxy}
    onStopProxy={actions.stopProxy}
    onSetProxyPort={actions.setProxyPort}
    onSetAllowLAN={actions.setAllowLAN}
    onSetAutoStartProxy={actions.setAutoStartProxy}
    onSetProxyAPIKey={actions.setProxyAPIKey}
    onRegenerateProxyAPIKey={actions.regenerateProxyAPIKey}
    onSetAuthorizationMode={actions.setAuthorizationMode}
  />

  <CloudflaredPanel
    proxyStatus={store.proxyStatus}
    busy={store.busy}
    onRefreshCloudflaredStatus={actions.refreshCloudflaredStatus}
    onSetCloudflaredConfig={actions.setCloudflaredConfig}
    onInstallCloudflared={actions.installCloudflared}
    onStartCloudflared={actions.startCloudflared}
    onStopCloudflared={actions.stopCloudflared}
  />

  <CliSyncPanel
    busy={store.busy}
    proxyBaseURL={store.proxyBaseUrl}
    proxyAPIKey={store.proxyApiKey}
    onGetCLISyncStatuses={actions.getCliSyncStatuses}
    onGetEffectiveModelCatalog={actions.getEffectiveModelCatalog}
    onGetCLISyncFileContent={actions.getCliSyncFileContent}
    onSaveCLISyncFileContent={actions.saveCliSyncFileContent}
    onSyncCLIConfig={actions.syncCLIConfig}
  />

  <EndpointTesterPanel proxyStatus={store.proxyStatus} apiKey={store.proxyApiKey} onExecuteEndpointTest={actions.executeEndpointTest} />

  <ModelAliasPanel
    busy={store.busy}
    onGetModelAliases={actions.getModelAliases}
    onSetModelAliases={actions.setModelAliases}
  />

  <SchedulingPanel
    proxyStatus={store.proxyStatus}
    busy={store.busy}
    onSetSchedulingMode={actions.setSchedulingMode}
  />
</div>
