import type { RouterActions } from '@/app/services/app-controller'
import { routerApi } from '@/backend/gateways/router-gateway'
import type { EndpointTestRequest, EndpointTestResult, LocalModelCatalogItem } from '@/features/router/types'

export interface RouterPanelActions extends RouterActions {
  getEffectiveModelCatalog: (baseUrl: string, apiKey: string) => Promise<LocalModelCatalogItem[]>
  executeEndpointTest: (request: EndpointTestRequest) => Promise<EndpointTestResult>
}

export const createRouterPanelActions = (routerActions: RouterActions): RouterPanelActions => {
  return {
    ...routerActions,
    getEffectiveModelCatalog: (baseUrl: string, apiKey: string) => routerApi.getEffectiveModelCatalog(baseUrl, apiKey),
    executeEndpointTest: (request: EndpointTestRequest) => routerApi.executeEndpointTest(request)
  }
}
