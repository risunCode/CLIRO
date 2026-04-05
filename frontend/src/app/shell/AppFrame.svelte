<script lang="ts">
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

  let activeRoute: AppRoute = getAppRoute(shell.activeTab)
  let activeRouteComponent: RouteComponent | null = null
  let routeOutletContext: RouteOutletContext

  const routeLoads: Partial<Record<AppTabId, Promise<void>>> = {}
  const loadedRouteComponents: Partial<Record<AppTabId, RouteComponent>> = {
    dashboard: APP_ROUTES.dashboard.component,
    accounts: APP_ROUTES.accounts.component
  }

  const handleTabChange = (tabId: AppTabId): void => {
    appActions.setActiveTab(tabId)
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

  $: activeRoute = getAppRoute(shell.activeTab)
  $: routeOutletContext = { shell, appActions, accountsActions, routerActions, logsActions, settingsActions }
  $: void ensureRouteLoaded(activeRoute)
  $: activeRouteComponent = loadedRouteComponents[activeRoute.id] ?? null
</script>

<div class="flex h-full flex-col">
  <AppHeader activeTab={shell.activeTab} onSelectTab={handleTabChange} {onToggleTheme} {theme} />

  <section class="no-scrollbar min-h-0 flex-1 overflow-y-auto px-4 py-4 md:px-6">
    <div class="space-y-4 pb-1">
      <RouteOutlet route={activeRoute} component={activeRouteComponent} context={routeOutletContext} />
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
