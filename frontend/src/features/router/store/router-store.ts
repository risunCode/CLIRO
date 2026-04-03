import type { ProxyStatus } from '@/features/router/types'

export interface RouterStoreState {
  proxyStatus: ProxyStatus | null
  busy: boolean
  proxyBaseUrl: string
  proxyApiKey: string
  proxyBaseURL: string
  proxyAPIKey: string
}

export const createRouterStoreState = (state: { proxyStatus: ProxyStatus | null; busy: boolean }): RouterStoreState => {
  const proxyBaseUrl = state.proxyStatus?.url || ''
  const proxyApiKey = state.proxyStatus?.proxyApiKey || ''

  return {
    proxyStatus: state.proxyStatus,
    busy: state.busy,
    proxyBaseUrl,
    proxyApiKey,
    proxyBaseURL: proxyBaseUrl,
    proxyAPIKey: proxyApiKey
  }
}
