<script lang="ts">
  import { ChevronDown, ChevronUp } from 'lucide-svelte'
  import { slide } from 'svelte/transition'
  import SurfaceCard from '@/shared/components/SurfaceCard.svelte'

  export let open = false
  export let title = ''
  export let subtitle = ''
  export let pill = ''
  export let ariaLabel = 'Toggle section'
  export let icon: typeof ChevronDown | null = null
  export let iconClassName = ''
  export let className = 'api-cli-sync p-0'
  export let bodyClassName = 'api-cli-sync-body'

  const toggleOpen = (): void => {
    open = !open
  }
</script>

<SurfaceCard {className}>
  <button type="button" class="api-cli-sync-header" on:click={toggleOpen} aria-expanded={open} aria-label={ariaLabel}>
    <span class="api-cli-sync-left">
      {#if icon}
        <span class={`api-cli-sync-icon-wrap ${iconClassName}`.trim()}>
          <svelte:component this={icon} size={15} />
        </span>
      {/if}

      <span class="api-cli-sync-copy">
        <span class="api-cli-sync-title-row">
          <span class="api-cli-sync-title">{title}</span>
        </span>
        {#if subtitle}
          <span class="api-cli-sync-subtitle">{subtitle}</span>
        {/if}
      </span>
    </span>

    <span class="api-cli-sync-header-right">
      {#if pill}
        <span class="api-cli-sync-pill">{pill}</span>
      {/if}
      {#if $$slots.headerRight}
        <slot name="headerRight" />
      {/if}
      <span class="api-cli-sync-chevron">
        {#if open}
          <ChevronUp size={15} />
        {:else}
          <ChevronDown size={15} />
        {/if}
      </span>
    </span>
  </button>

  {#if open}
    <div class={bodyClassName} transition:slide={{ duration: 180 }}>
      <slot />
    </div>
  {/if}
</SurfaceCard>
