export interface AppBootstrapDependencies {
  bindRuntimeEvents: () => () => void
  bindActivityEvents: () => () => void
  startProxyHeartbeat: () => void
  startActiveTabRefreshLoop: () => void
  refreshCore: () => Promise<void>
  maybeWaitForProxyAutostart: () => void
  refreshLogs: () => Promise<void>
  bindLogsSubscription: () => void
  checkForUpdates: () => Promise<void>
  onInitializeError: (error: unknown) => void
}

export interface AppBootstrapHandle {
  dispose: () => void
}

export const initializeAppBootstrap = async (dependencies: AppBootstrapDependencies): Promise<AppBootstrapHandle> => {
  const unbindRuntimeEvents = dependencies.bindRuntimeEvents()
  const unbindActivityEvents = dependencies.bindActivityEvents()

  dependencies.startProxyHeartbeat()
  dependencies.startActiveTabRefreshLoop()

  try {
    await dependencies.refreshCore()
    dependencies.maybeWaitForProxyAutostart()
    await dependencies.refreshLogs()
    dependencies.bindLogsSubscription()
    await dependencies.checkForUpdates()
  } catch (error) {
    dependencies.onInitializeError(error)
  }

  return {
    dispose: () => {
      unbindActivityEvents()
      unbindRuntimeEvents()
    }
  }
}
