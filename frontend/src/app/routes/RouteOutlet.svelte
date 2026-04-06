<script lang="ts">
  import type { AppRoute, RouteComponent } from '@/app/routes/app-routes'
  import SurfaceCard from '@/components/common/SurfaceCard.svelte'

  interface MountedRouteView {
    id: AppRoute['id']
    route: AppRoute
    component: RouteComponent
    props: Record<string, unknown>
  }

  export let route: AppRoute
  export let activeRouteId: AppRoute['id']
  export let mountedRoutes: MountedRouteView[] = []

  $: activeView = mountedRoutes.find((entry) => entry.id === activeRouteId) ?? null
</script>

{#if activeView}
  <div class="relative min-h-[20rem]">
    {#each mountedRoutes as entry (entry.id)}
      <div
        aria-hidden={entry.id === activeRouteId ? undefined : 'true'}
        class={`origin-top transition-all duration-150 ease-out ${
          entry.id === activeRouteId
            ? 'visible relative translate-y-0 opacity-100'
            : 'pointer-events-none invisible absolute inset-0 -translate-y-1 opacity-0 overflow-hidden'
        }`}
      >
        <svelte:component this={entry.component} {...entry.props} />
      </div>
    {/each}
  </div>
{:else}
  <SurfaceCard className="p-4 text-sm text-text-secondary">{route.loadingLabel}</SurfaceCard>
{/if}
