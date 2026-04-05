import type { AppActions, AppShellState, AccountsActions, LogsActions, RouterActions, SettingsActions } from '@/app/services/app-controller'
import type { AppTabId } from '@/app/utils/tabs'
import AccountsTab from '@/tabs/AccountsTab.svelte'
import DashboardTab from '@/tabs/DashboardTab.svelte'

export type RouteComponent = new (...args: any[]) => any

export interface RouteOutletContext {
  shell: AppShellState
  appActions: AppActions
  accountsActions: AccountsActions
  routerActions: RouterActions
  logsActions: LogsActions
  settingsActions: SettingsActions
}

interface BaseRoute {
  id: AppTabId
  loadingLabel: string
  buildProps: (context: RouteOutletContext) => Record<string, unknown>
}

interface StaticRoute extends BaseRoute {
  component: RouteComponent
}

interface LazyRoute extends BaseRoute {
  load: () => Promise<RouteComponent>
}

export type AppRoute = StaticRoute | LazyRoute

interface AppRoutesByTab {
  dashboard: StaticRoute
  accounts: StaticRoute
  'api-router': LazyRoute
  usage: LazyRoute
  'system-logs': LazyRoute
  settings: LazyRoute
}

export const APP_ROUTES: AppRoutesByTab = {
  dashboard: {
    id: 'dashboard',
    loadingLabel: 'Loading dashboard...',
    component: DashboardTab,
    buildProps: ({ shell }) => ({
      state: shell.state,
      accounts: shell.accounts,
      proxyStatus: shell.proxyStatus,
      loading: shell.loadingDashboard,
      loadingProxyStatus: shell.loadingProxyStatus,
      waitingForProxyAutostart: shell.waitingForProxyAutostart
    })
  },
  accounts: {
    id: 'accounts',
    loadingLabel: 'Loading accounts...',
    component: AccountsTab,
    buildProps: ({ shell, appActions, accountsActions }) => ({
      shell,
      appActions,
      accountsActions
    })
  },
  'api-router': {
    id: 'api-router',
    loadingLabel: 'Loading API Router...',
    load: async () => (await import('@/tabs/ApiRouterTab.svelte')).default,
    buildProps: ({ shell, routerActions }) => ({
      proxyStatus: shell.proxyStatus,
      busy: shell.proxyBusy,
      routerActions
    })
  },
  usage: {
    id: 'usage',
    loadingLabel: 'Loading usage view...',
    load: async () => (await import('@/tabs/UsageTab.svelte')).default,
    buildProps: ({ shell }) => ({
      state: shell.state,
      accounts: shell.accounts,
      proxyStatus: shell.proxyStatus,
      logs: shell.logs
    })
  },
  'system-logs': {
    id: 'system-logs',
    loadingLabel: 'Loading system logs...',
    load: async () => (await import('@/tabs/SystemLogsTab.svelte')).default,
    buildProps: ({ shell, logsActions }) => ({
      shell,
      logsActions
    })
  },
  settings: {
    id: 'settings',
    loadingLabel: 'Loading settings...',
    load: async () => (await import('@/features/settings/components/SettingsScreen.svelte')).default,
    buildProps: ({ settingsActions }) => ({
      settingsActions
    })
  }
}

export const isLazyRoute = (route: AppRoute): route is LazyRoute => {
  return 'load' in route
}

export const getAppRoute = (tabId: AppTabId): AppRoute => {
  return APP_ROUTES[tabId]
}
