import { subscribeRuntimeEvent } from '@/backend/client/runtime-events'

export interface AppRuntimeEventHandlers {
  onSecondInstanceNotice: (payload: unknown) => void
  onCloseRequested: () => void
  onWindowRestored: () => void
  onProxyStateChanged: () => void
  onTrayStateChanged: () => void
}

export const bindAppRuntimeEvents = (handlers: AppRuntimeEventHandlers): (() => void) => {
  const unsubscribeSecondInstance = subscribeRuntimeEvent('app:second-instance', handlers.onSecondInstanceNotice)
  const unsubscribeCloseRequested = subscribeRuntimeEvent('app:close-requested', handlers.onCloseRequested)
  const unsubscribeWindowRestored = subscribeRuntimeEvent('app:window-restored', handlers.onWindowRestored)
  const unsubscribeProxyStateChanged = subscribeRuntimeEvent('app:proxy-state-changed', handlers.onProxyStateChanged)
  const unsubscribeTrayStateChanged = subscribeRuntimeEvent('app:tray-state-changed', handlers.onTrayStateChanged)

  return () => {
    unsubscribeSecondInstance()
    unsubscribeCloseRequested()
    unsubscribeWindowRestored()
    unsubscribeProxyStateChanged()
    unsubscribeTrayStateChanged()
  }
}

export interface AppActivityEventHandlers {
  isDocumentVisible: () => boolean
  onVisible: () => void
  onFocus: () => void
}

export const bindAppActivityEvents = (handlers: AppActivityEventHandlers): (() => void) => {
  let removeVisibilityListener: (() => void) | null = null
  let removeFocusListener: (() => void) | null = null

  if (typeof document !== 'undefined') {
    const handleVisibilityChange = (): void => {
      if (!handlers.isDocumentVisible()) {
        return
      }
      handlers.onVisible()
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    removeVisibilityListener = () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }

  if (typeof window !== 'undefined') {
    const handleFocus = (): void => {
      if (!handlers.isDocumentVisible()) {
        return
      }
      handlers.onFocus()
    }

    window.addEventListener('focus', handleFocus)
    removeFocusListener = () => {
      window.removeEventListener('focus', handleFocus)
    }
  }

  return () => {
    removeVisibilityListener?.()
    removeFocusListener?.()
  }
}
