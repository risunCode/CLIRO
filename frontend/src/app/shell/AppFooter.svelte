<script lang="ts">
  import type { AppState } from '@/app/types'
  import type { ProxyStatus } from '@/features/router/types'
  import { openBrowserURL } from '@/backend/client/browser'

  const repoURL = 'https://github.com/risunCode/CLIRO'
  export let proxyStatus: ProxyStatus | null = null
  export let state: AppState | null = null
  export let loading = false
  export let loadingProxyStatus = false
  export let waitingForProxyAutostart = false

  const openURL = (event: MouseEvent, url: string): void => {
    event.preventDefault()
    openBrowserURL(url)
  }

  $: proxyReady = proxyStatus !== null || (!loadingProxyStatus && state !== null)
  $: proxyRunning = proxyStatus?.running ?? (!loadingProxyStatus ? state?.proxyRunning ?? false : false)
  $: proxyBaseURL = proxyStatus?.url || state?.proxyUrl || ''
  $: endpointLabel = proxyBaseURL.replace(/^https?:\/\//, '')
  $: cloudflaredKnown = proxyStatus !== null
  $: cloudflaredRunning = proxyStatus?.cloudflared.running ?? false
  $: proxyToneClass = waitingForProxyAutostart
    ? 'service-pill-starting'
    : !proxyReady && (loading || loadingProxyStatus)
      ? 'service-pill-loading'
      : proxyRunning
        ? 'service-pill-online'
        : 'service-pill-offline'
  $: proxyLabel = waitingForProxyAutostart
    ? 'Starting'
    : !proxyReady && (loading || loadingProxyStatus)
      ? 'Loading'
      : proxyRunning
        ? 'Online'
        : 'Offline'
</script>

<footer class="rounded-t-base border border-border bg-surface px-4 py-3 text-xs text-text-secondary shadow-soft md:px-6">
  <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between md:gap-4">
    <div class="flex flex-wrap items-center gap-2">
      <span class={`service-pill ${proxyToneClass}`}>
        Proxy {proxyLabel}
      </span>
      {#if proxyRunning && endpointLabel}
        <code class="rounded-sm border border-border bg-app px-2 py-0.5 text-[11px] text-text-primary">{endpointLabel}</code>
      {/if}
      <span class={`service-pill ${!cloudflaredKnown && (loading || loadingProxyStatus) ? 'service-pill-loading' : cloudflaredRunning ? 'service-pill-online' : 'service-pill-offline'}`}>
        Cloudflared {!cloudflaredKnown && (loading || loadingProxyStatus) ? 'Loading' : cloudflaredRunning ? 'Online' : 'Offline'}
      </span>
    </div>

    <a
      href={repoURL}
      class="inline-flex w-fit items-center gap-1 rounded-sm border border-transparent px-2 py-1 text-text-secondary transition hover:border-border hover:bg-app hover:text-text-primary"
      on:click={(event) => openURL(event, repoURL)}
    >
      CLIRO GitHub
    </a>
  </div>
</footer>

<style>
  .service-pill {
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    border: 1px solid transparent;
    padding: 0.2rem 0.55rem;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    line-height: 1;
  }

  .service-pill-online {
    color: var(--color-success);
    border-color: color-mix(in srgb, var(--color-success) 46%, var(--color-border));
    background: color-mix(in srgb, var(--color-success) 14%, transparent);
  }

  .service-pill-loading {
    color: var(--color-info);
    border-color: color-mix(in srgb, var(--color-info) 46%, var(--color-border));
    background: color-mix(in srgb, var(--color-info) 14%, transparent);
  }

  .service-pill-starting {
    color: var(--color-warning);
    border-color: color-mix(in srgb, var(--color-warning) 46%, var(--color-border));
    background: color-mix(in srgb, var(--color-warning) 14%, transparent);
  }

  .service-pill-offline {
    color: var(--color-error);
    border-color: color-mix(in srgb, var(--color-error) 46%, var(--color-border));
    background: color-mix(in srgb, var(--color-error) 14%, transparent);
  }
</style>
