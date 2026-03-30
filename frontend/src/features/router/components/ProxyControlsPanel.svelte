<script lang="ts">
  import { onDestroy } from 'svelte'
  import { Copy, KeyRound, Network, Pencil, Power, PowerOff, Save } from 'lucide-svelte'
  import Button from '@/components/common/Button.svelte'
  import StatusBadge from '@/components/common/StatusBadge.svelte'
  import SurfaceCard from '@/components/common/SurfaceCard.svelte'
  import ToggleSwitch from '@/components/common/ToggleSwitch.svelte'
  import type { ProxyStatus } from '@/features/router/types'
  import { copyTextToClipboard, hasClipboardWrite } from '@/shared/lib/browser'

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

  let portInput = '8095'
  let portInputDirty = false
  let allowLanInput = false
  let autoStartProxyInput = true
  let authorizationModeInput = false
  let apiKeyInput = ''
  let apiKeyError = ''
  let apiKeyCopied = false
  let apiKeyCopyTimer: ReturnType<typeof setTimeout> | null = null

  $: if (proxyStatus?.port && !portInputDirty) {
    portInput = String(proxyStatus.port)
  }

  $: if (proxyStatus && !busy) {
    allowLanInput = proxyStatus.allowLan
    autoStartProxyInput = proxyStatus.autoStartProxy
    authorizationModeInput = proxyStatus.authorizationMode
    apiKeyInput = proxyStatus.proxyApiKey || ''
  }

  const applyProxyPort = async (): Promise<void> => {
    const parsedPort = Number.parseInt(portInput.trim(), 10)
    const nextPort = Number.isFinite(parsedPort) && parsedPort >= 1024 && parsedPort <= 65535 ? parsedPort : 8095
    portInput = String(nextPort)
    portInputDirty = false
    await onSetProxyPort(nextPort)
  }

  const updateAllowLan = async (): Promise<void> => {
    await onSetAllowLAN(allowLanInput)
  }

  const updateAutoStartProxy = async (): Promise<void> => {
    await onSetAutoStartProxy(autoStartProxyInput)
  }

  const updateAuthorizationMode = async (): Promise<void> => {
    await onSetAuthorizationMode(authorizationModeInput)
  }

  const editProxyAPIKey = async (): Promise<void> => {
    const currentValue = apiKeyInput || ''
    const updatedValue = window.prompt('Set proxy API key', currentValue)
    if (updatedValue === null) {
      return
    }

    const normalized = updatedValue.trim()
    if (normalized.length === 0) {
      apiKeyError = 'API key cannot be empty.'
      return
    }

    apiKeyError = ''
    await onSetProxyAPIKey(normalized)
    apiKeyInput = normalized
  }

  const regenerateProxyAPIKey = async (): Promise<void> => {
    apiKeyError = ''
    const nextKey = await onRegenerateProxyAPIKey()
    apiKeyInput = nextKey
  }

  const copyProxyAPIKey = async (): Promise<void> => {
    if (!hasClipboardWrite() || apiKeyInput.trim().length === 0) {
      return
    }

    const copied = await copyTextToClipboard(apiKeyInput)
    if (!copied) {
      return
    }

    apiKeyCopied = true
    if (apiKeyCopyTimer) {
      clearTimeout(apiKeyCopyTimer)
    }
    apiKeyCopyTimer = setTimeout(() => {
      apiKeyCopied = false
      apiKeyCopyTimer = null
    }, 1200)
  }

  onDestroy(() => {
    if (apiKeyCopyTimer) {
      clearTimeout(apiKeyCopyTimer)
    }
  })
</script>

<SurfaceCard className="p-4">
  <div class="mb-3 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
    <div>
      <p class="text-sm font-semibold text-text-primary">Proxy Service</p>
      <p class="text-xs text-text-secondary">Grid controls for runtime, bind mode, and startup behavior.</p>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <StatusBadge tone={proxyStatus?.running ? 'success' : 'error'}>
        {proxyStatus?.running ? 'Running' : 'Stopped'}
      </StatusBadge>
    </div>
  </div>

  <div class="grid gap-3 lg:grid-cols-2">
    <div class="rounded-sm border border-border bg-app p-3">
      <p class="mb-2 text-xs font-semibold uppercase tracking-[0.06em] text-text-secondary">Port</p>
      <div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-end">
        <input
          id="router-port"
          class="ui-control-input ui-control-select px-3 text-sm"
          bind:value={portInput}
          on:input={() => {
            portInputDirty = true
          }}
          type="text"
          inputmode="numeric"
          pattern="[0-9]*"
        />
        <Button variant="secondary" size="sm" on:click={applyProxyPort} disabled={busy}>
          <Save size={14} class="mr-1" />
          Apply
        </Button>
      </div>
      <p class="mt-2 truncate text-xs text-text-secondary">Bind Address: {proxyStatus?.bindAddress || '-'}</p>
    </div>

    <div class="rounded-sm border border-border bg-app p-3">
      <p class="mb-2 text-xs font-semibold uppercase tracking-[0.06em] text-text-secondary">Runtime</p>
      <div class="grid gap-2 sm:grid-cols-2">
        <Button variant="primary" size="sm" on:click={onStartProxy} disabled={busy || proxyStatus?.running}>
          <Power size={14} class="mr-1" />
          Start Proxy
        </Button>
        <Button variant="danger" size="sm" on:click={onStopProxy} disabled={busy || !proxyStatus?.running}>
          <PowerOff size={14} class="mr-1" />
          Stop Proxy
        </Button>
      </div>
      <p class="mt-1 text-[11px] text-text-secondary">Use this as the base URL. CLIro-Go now defaults to `/v1`, and `/v1/v1/...` aliases remain safe when a client prepends it again.</p>
    </div>

    <div class="rounded-sm border border-border bg-gradient-to-br from-surface/90 to-app p-3 lg:col-span-2">
      <div class="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.06em] text-text-secondary">Network & Startup</p>
          <p class="mt-1 text-[11px] text-text-secondary">Control how the local proxy binds to the network and how it behaves when the desktop app starts.</p>
        </div>
        <span class="rounded-full border border-border bg-app px-2.5 py-1 text-[10px] uppercase tracking-[0.08em] text-text-secondary">Lifecycle Controls</span>
      </div>

      <div class="grid gap-3 lg:grid-cols-2">
        <div class="grid gap-3 rounded-sm border border-border bg-app/90 p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
          <div>
            <div class="mb-2 flex items-center gap-2">
              <span class="inline-flex h-8 w-8 items-center justify-center rounded-sm border border-border bg-surface text-text-primary">
                <Network size={15} />
              </span>
              <p class="text-sm font-semibold text-text-primary">Allow on LAN</p>
            </div>
            <p class="text-[11px] leading-5 text-text-secondary">Expose the proxy to other devices on the same network. Keep this off when the proxy should stay localhost-only.</p>
          </div>
          <div class="min-w-[240px] rounded-sm border border-border bg-surface/70 p-2">
            <ToggleSwitch label={allowLanInput ? 'LAN access enabled' : 'Localhost only'} bind:checked={allowLanInput} on:change={updateAllowLan} disabled={busy} />
          </div>
        </div>

        <div class="grid gap-3 rounded-sm border border-border bg-app/90 p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
          <div>
            <div class="mb-2 flex items-center gap-2">
              <span class="inline-flex h-8 w-8 items-center justify-center rounded-sm border border-border bg-surface text-text-primary">
                <Power size={15} />
              </span>
              <p class="text-sm font-semibold text-text-primary">Auto Start Proxy</p>
            </div>
            <p class="text-[11px] leading-5 text-text-secondary">Bring the local proxy online automatically when CLIro-Go launches so local clients can reconnect without manual intervention.</p>
          </div>
          <div class="min-w-[240px] rounded-sm border border-border bg-surface/70 p-2">
            <ToggleSwitch
              label={autoStartProxyInput ? 'Start proxy on launch' : 'Manual start only'}
              bind:checked={autoStartProxyInput}
              on:change={updateAutoStartProxy}
              disabled={busy}
            />
          </div>
        </div>
      </div>
    </div>

    <div class="rounded-sm border border-border bg-gradient-to-br from-surface/90 to-app p-3 lg:col-span-2">
      <div class="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.06em] text-text-secondary">Security</p>
          <p class="mt-1 text-[11px] text-text-secondary">Manage the proxy access key and decide whether every inbound request must authenticate.</p>
        </div>
        <StatusBadge tone={authorizationModeInput ? 'warning' : 'neutral'}>
          {authorizationModeInput ? 'API Key Required' : 'Open Access'}
        </StatusBadge>
      </div>

      <div class="grid gap-3 lg:grid-cols-2">
        <div class="rounded-sm border border-border bg-app/90 p-3">
          <div class="mb-3 flex items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <div class="mb-2 flex items-center gap-2">
                <p class="text-[10px] font-semibold uppercase tracking-[0.08em] text-text-secondary">API Key Vault</p>
                <span class="rounded-full border border-border px-2 py-0.5 text-[10px] text-text-secondary">Authorization or X-API-Key</span>
              </div>
              <p class="text-[11px] leading-5 text-text-secondary">Use this key with desktop clients, scripts, or remote access tooling when authorization mode is enabled.</p>
            </div>
            <span class="inline-flex h-10 w-10 items-center justify-center rounded-sm border border-border bg-surface text-text-primary">
              <KeyRound size={16} />
            </span>
          </div>

          <div class="rounded-sm border border-border bg-surface px-3 py-3 font-mono text-xs text-text-primary break-all">
            {apiKeyInput || '-'}
          </div>

          <div class="mt-3 flex flex-wrap items-center gap-2">
            <Button variant="secondary" size="sm" on:click={editProxyAPIKey} disabled={busy}>
              <Pencil size={13} class="mr-1" />
              Edit
            </Button>
            <Button variant="secondary" size="sm" on:click={regenerateProxyAPIKey} disabled={busy}>
              <KeyRound size={13} class="mr-1" />
              Regen
            </Button>
            <Button variant="secondary" size="sm" on:click={copyProxyAPIKey} disabled={!hasClipboardWrite() || apiKeyInput.trim().length === 0}>
              <Copy size={13} class="mr-1" />
              {apiKeyCopied ? 'Copied' : 'Copy'}
            </Button>
          </div>

          {#if apiKeyError}
            <p class="mt-2 text-xs text-error">{apiKeyError}</p>
          {/if}
        </div>

        <div class="grid gap-3 rounded-sm border border-border bg-app/90 p-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
          <div>
            <p class="text-sm font-semibold text-text-primary">Authorization Gate</p>
            <p class="mt-1 text-[11px] leading-5 text-text-secondary">Require the configured key for every proxy route while still keeping the full OpenAI-compatible and Anthropic-compatible surface available.</p>
            <div class="mt-3 rounded-sm border border-border bg-surface/40 px-3 py-2.5 text-[11px] leading-5 text-text-secondary">
              Clients can authenticate with `Authorization: Bearer &lt;key&gt;` or `X-API-Key: &lt;key&gt;`.
            </div>
          </div>

          <div class="min-w-[260px] rounded-sm border border-border bg-surface/70 p-2">
            <ToggleSwitch
              label={authorizationModeInput ? 'API key required for all routes' : 'Requests allowed without API key'}
              bind:checked={authorizationModeInput}
              on:change={updateAuthorizationMode}
              disabled={busy}
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</SurfaceCard>
