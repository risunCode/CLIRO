<script lang="ts">
  import { onMount, tick } from 'svelte'
  import type { AppActions, AppShellState, AccountsActions, LogsActions, RouterActions, SettingsActions } from '@/app/services/app-controller'
  import type { AppTabId } from '@/app/utils/tabs'
  import { APP_ROUTES, getAppRoute, isLazyRoute, type AppRoute, type RouteComponent, type RouteOutletContext } from '@/app/routes/app-routes'
  import RouteOutlet from '@/app/routes/RouteOutlet.svelte'
  import AppFooter from '@/app/shell/AppFooter.svelte'
  import AppHeader from '@/app/shell/AppHeader.svelte'
  import type { Theme } from '@/shared/stores/theme'

  export let shell: AppShellState
  export let theme: Theme = 'light'
  export let appActions: AppActions
  export let accountsActions: AccountsActions
  export let routerActions: RouterActions
  export let logsActions: LogsActions
  export let settingsActions: SettingsActions
  export let onToggleTheme: () => void

  interface MountedRouteView {
    id: AppTabId
    route: AppRoute
    component: RouteComponent
    props: Record<string, unknown>
  }

  const LAZY_PRELOAD_ORDER: AppTabId[] = ['api-router', 'usage', 'system-logs', 'settings']

  let activeRoute: AppRoute = getAppRoute(shell.activeTab)
  let routeOutletContext: RouteOutletContext
  let mountedRouteIds: AppTabId[] = [shell.activeTab]
  let mountedRouteViews: MountedRouteView[] = []
  let contentViewport: HTMLElement | null = null

  const tabScrollOffsets: Partial<Record<AppTabId, number>> = {
    [shell.activeTab]: 0
  }

  const routeLoads: Partial<Record<AppTabId, Promise<void>>> = {}
  const loadedRouteComponents: Partial<Record<AppTabId, RouteComponent>> = {
    dashboard: APP_ROUTES.dashboard.component,
    accounts: APP_ROUTES.accounts.component
  }

  const pushMountedRoute = (tabId: AppTabId): void => {
    if (mountedRouteIds.includes(tabId)) {
      return
    }
    mountedRouteIds = [...mountedRouteIds, tabId]
  }

  const handleTabChange = async (tabId: AppTabId): Promise<void> => {
    if (tabId === shell.activeTab) {
      return
    }

    if (contentViewport) {
      tabScrollOffsets[shell.activeTab] = contentViewport.scrollTop
    }

    appActions.setActiveTab(tabId)

    await ensureRouteLoaded(getAppRoute(tabId))
    pushMountedRoute(tabId)
    await tick()

    if (contentViewport) {
      contentViewport.scrollTop = tabScrollOffsets[tabId] ?? 0
    }
  }

  const handlePrefetchTab = (tabId: AppTabId): void => {
    void ensureRouteLoaded(getAppRoute(tabId))
  }

  const ensureRouteLoaded = (route: AppRoute): Promise<void> => {
    if (!isLazyRoute(route)) {
      loadedRouteComponents[route.id] = route.component
      return Promise.resolve()
    }

    if (loadedRouteComponents[route.id]) {
      return Promise.resolve()
    }

    if (!routeLoads[route.id]) {
      routeLoads[route.id] = route.load().then((component) => {
        loadedRouteComponents[route.id] = component
      })
    }

    return routeLoads[route.id] as Promise<void>
  }

  const preloadLazyRoutes = (): void => {
    const run = async (): Promise<void> => {
      for (const tabId of LAZY_PRELOAD_ORDER) {
        await ensureRouteLoaded(getAppRoute(tabId))
      }
    }

    if (typeof window === 'undefined') {
      return
    }

    const idleLoader = (window as Window & {
      requestIdleCallback?: (callback: IdleRequestCallback) => number
    }).requestIdleCallback

    if (typeof idleLoader === 'function') {
      idleLoader(() => {
        void run()
      })
      return
    }

    window.setTimeout(() => {
      void run()
    }, 120)
  }

  onMount(() => {
    pushMountedRoute(shell.activeTab)
    preloadLazyRoutes()
  })

  $: activeRoute = getAppRoute(shell.activeTab)
  $: routeOutletContext = { shell, appActions, accountsActions, routerActions, logsActions, settingsActions }
  $: void ensureRouteLoaded(activeRoute)
  $: if (loadedRouteComponents[activeRoute.id]) {
    pushMountedRoute(activeRoute.id)
  }
  $: mountedRouteViews = mountedRouteIds
    .map((tabId) => {
      const route = getAppRoute(tabId)
      const component = loadedRouteComponents[tabId]
      if (!component) {
        return null
      }
      return {
        id: tabId,
        route,
        component,
        props: route.buildProps(routeOutletContext)
      }
    })
    .filter((entry): entry is MountedRouteView => entry !== null)
</script>

<div class="flex h-full flex-col">
  <AppHeader activeTab={shell.activeTab} onSelectTab={handleTabChange} onPrefetchTab={handlePrefetchTab} {onToggleTheme} {theme} />

  <section bind:this={contentViewport} class="no-scrollbar min-h-0 flex-1 overflow-y-auto px-4 py-4 md:px-6">
    <div class="space-y-4 pb-1">
      <RouteOutlet route={activeRoute} activeRouteId={shell.activeTab} mountedRoutes={mountedRouteViews} />
    </div>
  </section>

  <AppFooter
    proxyStatus={shell.proxyStatus}
    state={shell.state}
    loading={shell.loadingDashboard}
    loadingProxyStatus={shell.loadingProxyStatus}
    waitingForProxyAutostart={shell.waitingForProxyAutostart}
  />
</div>
