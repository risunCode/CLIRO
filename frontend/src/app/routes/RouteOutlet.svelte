<script lang="ts">
  import type { AppRoute, RouteComponent, RouteOutletContext } from '@/app/routes/app-routes'
  import SurfaceCard from '@/components/common/SurfaceCard.svelte'

  export let route: AppRoute
  export let component: RouteComponent | null = null
  export let context: RouteOutletContext

  $: routeProps = route.buildProps(context)
</script>

{#if component}
  <svelte:component this={component} {...routeProps} />
{:else}
  <SurfaceCard className="p-4 text-sm text-text-secondary">{route.loadingLabel}</SurfaceCard>
{/if}
