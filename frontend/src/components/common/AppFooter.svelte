<script lang="ts">
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
  import type { ProxyStatus } from '@/features/router/types'

  const repoURL = 'https://github.com/risunCode/Cliro-Go'
  export let proxyStatus: ProxyStatus | null = null

  const openURL = (event: MouseEvent, url: string): void => {
    event.preventDefault()
    BrowserOpenURL(url)
  }

  $: proxyRunning = proxyStatus?.running ?? false
  $: serviceAddress = proxyStatus?.bindAddress || `${proxyStatus?.allowLan ? '0.0.0.0' : '127.0.0.1'}:${proxyStatus?.port ?? 8095}`
  $: proxyBaseURL = proxyStatus?.url || 'http://127.0.0.1:8095/v1'
</script>

<footer class="rounded-t-base border border-border bg-surface px-4 py-3 text-xs text-text-secondary shadow-soft md:px-6">
  <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between md:gap-4">
    <div class="flex flex-wrap items-center gap-2">
      <span class={`service-pill ${proxyRunning ? 'service-pill-online' : 'service-pill-offline'}`}>
        {proxyRunning ? 'Online' : 'Offline'}
      </span>
      {#if proxyRunning}
        <code class="rounded-sm border border-border bg-app px-2 py-0.5 text-[11px] text-text-primary">{serviceAddress}</code>
        <code class="rounded-sm border border-border bg-app px-2 py-0.5 text-[11px] text-text-primary">{proxyBaseURL}</code>
      {:else}
        <span class="text-text-secondary">Proxy service is stopped.</span>
      {/if}
    </div>

    <a
      href={repoURL}
      class="inline-flex w-fit items-center gap-1 rounded-sm border border-transparent px-2 py-1 text-text-secondary transition hover:border-border hover:bg-app hover:text-text-primary"
      on:click={(event) => openURL(event, repoURL)}
    >
      CLIrouter Github
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

  .service-pill-offline {
    color: var(--color-error);
    border-color: color-mix(in srgb, var(--color-error) 46%, var(--color-border));
    background: color-mix(in srgb, var(--color-error) 14%, transparent);
  }
</style>
