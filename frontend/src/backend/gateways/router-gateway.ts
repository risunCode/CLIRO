import { wailsClient } from '@/backend/client/wails-client'
import { toCliSyncResult, toCliSyncStatus, toLocalModelCatalogItem, toProxyStatus } from '@/backend/compat/router-compat'
import { buildEndpointTarget, getEndpointPreset } from '@/features/router/utils/endpoint-tester'
import { getErrorMessage } from '@/shared/utils/error'
import type {
  CliSyncFileInput,
  CliSyncResult,
  CliSyncStatus,
  CloudflaredAction,
  EndpointTestRequest,
  EndpointTestResult,
  LocalModelCatalogItem,
  ProxyRuntimeAction,
  ProxySettingsUpdateResult,
  ProxyStatus,
  RunCliSyncInput,
  SaveCliSyncFileInput,
  UpdateCloudflaredSettingsInput,
  UpdateProxySettingsInput,
} from '@/features/router/types'

const fetchProxyModelCatalog = async (baseUrl: string, apiKey: string): Promise<LocalModelCatalogItem[]> => {
  const normalizedBaseUrl = baseUrl.trim().replace(/\/+$/, '')
  if (!normalizedBaseUrl) {
    return []
  }

  const headers: Record<string, string> = {}
  const normalizedApiKey = apiKey.trim()
  if (normalizedApiKey) {
    headers.Authorization = `Bearer ${normalizedApiKey}`
  }

  const response = await fetch(`${normalizedBaseUrl}/v1/models`, { headers })
  if (!response.ok) {
    throw new Error(`Proxy model catalog request failed with status ${response.status}.`)
  }

  const payload = (await response.json()) as { data?: Array<{ id?: string; ownedBy?: string; owned_by?: string }> }
  if (!Array.isArray(payload.data)) {
    return []
  }

  return payload.data
    .map((item) => ({
      id: typeof item.id === 'string' ? item.id : '',
      ownedBy: typeof item.ownedBy === 'string' ? item.ownedBy : typeof item.owned_by === 'string' ? item.owned_by : ''
    }))
    .filter((item) => item.id)
}

const getEffectiveModelCatalog = async (baseUrl: string, apiKey: string): Promise<LocalModelCatalogItem[]> => {
  try {
    const localModels = (await wailsClient.router.getLocalModelCatalog()).map(toLocalModelCatalogItem)
    if (localModels.length > 0) {
      return localModels
    }
  } catch {
    // Fall through to proxy model catalog lookup.
  }

  return fetchProxyModelCatalog(baseUrl, apiKey)
}

const executeEndpointTest = async ({ baseUrl, apiKey, endpointId, body = '' }: EndpointTestRequest): Promise<EndpointTestResult> => {
  const endpoint = getEndpointPreset(endpointId)
  const target = buildEndpointTarget(baseUrl, endpoint.path)
  const headers: Record<string, string> = {}
  const normalizedApiKey = apiKey.trim()
  if (normalizedApiKey) {
    headers.Authorization = `Bearer ${normalizedApiKey}`
    headers['X-API-Key'] = normalizedApiKey
  }

  const options: RequestInit = {
    method: endpoint.method,
    headers
  }

  if (endpoint.method === 'POST') {
    headers['Content-Type'] = 'application/json'
    options.body = body
  }

  try {
    const response = await fetch(target, options)
    const contentType = response.headers.get('content-type') || ''

    if (contentType.includes('text/event-stream')) {
      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('Response body is not readable')
      }

      const decoder = new TextDecoder()
      let responseText = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        responseText += decoder.decode(value, { stream: true })
      }

      return {
        status: `${response.status} ${response.statusText}`,
        responseText,
      }
    }

    const responseText = contentType.includes('application/json')
      ? JSON.stringify(await response.json(), null, 2)
      : await response.text()

    return {
      status: `${response.status} ${response.statusText}`,
      responseText,
    }
  } catch (error) {
    throw new Error(getErrorMessage(error, 'Request failed'))
  }
}

export const routerApi = {
  getProxyStatus: async (): Promise<ProxyStatus> => toProxyStatus(await wailsClient.router.getProxyStatus()),
  refreshCloudflaredStatus: async (): Promise<ProxyStatus> => toProxyStatus(await wailsClient.router.runCloudflaredAction('refresh-status')),
  getEffectiveModelCatalog,
  getCliSyncStatuses: async (): Promise<CliSyncStatus[]> =>
    (await wailsClient.router.getCliSyncStatuses())
      .map(toCliSyncStatus)
      .filter((status): status is CliSyncStatus => status !== null),
  getCliSyncFile: (input: CliSyncFileInput): Promise<string> =>
    wailsClient.router.getCliSyncFile({ target: input.target, path: input.path }),
  saveCliSyncFile: (input: SaveCliSyncFileInput): Promise<void> =>
    wailsClient.router.saveCliSyncFile({ target: input.target, path: input.path, content: input.content }),
  runCliSync: async (input: RunCliSyncInput): Promise<CliSyncResult> =>
    toCliSyncResult(await wailsClient.router.runCliSync({ target: input.target, model: input.model || '' })),
  updateProxySettings: async (input: UpdateProxySettingsInput): Promise<ProxySettingsUpdateResult> =>
    wailsClient.router.updateProxySettings({
      port: input.port,
      allowLan: input.allowLan,
      autoStartProxy: input.autoStartProxy,
      proxyApiKey: input.proxyApiKey,
      regenerateApiKey: input.regenerateApiKey ?? false,
      authorizationMode: input.authorizationMode,
      schedulingMode: input.schedulingMode,
    }),
  updateCloudflaredSettings: (input: UpdateCloudflaredSettingsInput): Promise<void> =>
    wailsClient.router.updateCloudflaredSettings({
      mode: input.mode,
      token: input.token,
      useHttp2: input.useHttp2,
    }),
  runProxyAction: (action: ProxyRuntimeAction): Promise<void> => wailsClient.router.runProxyAction(action),
  runCloudflaredAction: async (action: CloudflaredAction): Promise<ProxyStatus> =>
    toProxyStatus(await wailsClient.router.runCloudflaredAction(action)),
  executeEndpointTest,
  getModelAliases: (): Promise<Record<string, string>> => wailsClient.router.getModelAliases(),
  setModelAliases: (aliases: Record<string, string>): Promise<void> => wailsClient.router.setModelAliases(aliases),
}
