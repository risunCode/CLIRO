import type {
  WailsCliSyncResult,
  WailsCliSyncStatus,
  WailsModelCatalogItem,
  WailsProxyStatus,
} from '@/backend/models/wails'
import type { CliSyncAppID, CliSyncResult, CliSyncStatus, LocalModelCatalogItem, ProxyStatus } from '@/features/router/types'

const toCliSyncAppID = (value: string): CliSyncAppID | null => {
  if (value === 'claude-code' || value === 'opencode-cli' || value === 'kilo-cli' || value === 'codex-ai') {
    return value
  }
  return null
}

export const toProxyStatus = (payload: WailsProxyStatus): ProxyStatus => {
  const cloudflaredMode = payload.cloudflared.mode === 'auth' ? 'auth' : 'quick'

  return {
    running: payload.running,
    port: payload.port,
    url: payload.url,
    bindAddress: payload.bindAddress,
    allowLan: payload.allowLan,
    autoStartProxy: payload.autoStartProxy,
    proxyApiKey: payload.proxyApiKey,
    authorizationMode: payload.authorizationMode,
    schedulingMode: payload.schedulingMode || 'balance',
    cloudflared: {
      enabled: payload.cloudflared.enabled,
      mode: cloudflaredMode,
      token: payload.cloudflared.token,
      useHttp2: payload.cloudflared.useHttp2,
      installed: payload.cloudflared.installed,
      version: payload.cloudflared.version,
      running: payload.cloudflared.running,
      url: payload.cloudflared.url,
      error: payload.cloudflared.error,
    },
  }
}

export const toCliSyncStatus = (payload: WailsCliSyncStatus): CliSyncStatus | null => {
  const id = toCliSyncAppID(payload.id)
  if (!id) {
    return null
  }

  return {
    id,
    label: payload.label,
    installed: payload.installed,
    installPath: payload.installPath || undefined,
    version: payload.version || undefined,
    synced: payload.synced,
    currentBaseUrl: payload.currentBaseUrl || undefined,
    currentModel: payload.currentModel || undefined,
    files: payload.files.map((file) => ({
      name: file.name,
      path: file.path,
    })),
  }
}

export const toCliSyncResult = (payload: WailsCliSyncResult): CliSyncResult => {
  const id = toCliSyncAppID(payload.id)
  if (!id) {
    throw new Error('Unsupported CLI sync result id')
  }

  return {
    id,
    label: payload.label,
    model: payload.model || undefined,
    currentBaseUrl: payload.currentBaseUrl || undefined,
    files: payload.files.map((file) => ({
      name: file.name,
      path: file.path,
    })),
  }
}

export const toLocalModelCatalogItem = (payload: WailsModelCatalogItem): LocalModelCatalogItem => {
  return {
    id: payload.id,
    ownedBy: payload.ownedBy,
  }
}
