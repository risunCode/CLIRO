export interface ProxyStatus {
  running: boolean
  port: number
  url: string
  bindAddress: string
  allowLan: boolean
  autoStartProxy: boolean
  proxyApiKey: string
  authorizationMode: boolean
  schedulingMode: string
  cloudflared: CloudflaredState
}

export interface CloudflaredState {
  enabled: boolean
  mode: 'quick' | 'auth'
  token: string
  useHttp2: boolean
  installed: boolean
  version: string
  running: boolean
  url: string
  error: string
}

export type CliSyncAppID = 'claude-code' | 'opencode-cli' | 'kilo-cli' | 'codex-ai'

export interface CliSyncFile {
  name: string
  path: string
}

export interface CliSyncStatus {
  id: CliSyncAppID
  label: string
  installed: boolean
  installPath?: string
  version?: string
  synced: boolean
  currentBaseUrl?: string
  currentModel?: string
  files: CliSyncFile[]
}

export interface CliSyncResult {
  id: CliSyncAppID
  label: string
  model?: string
  currentBaseUrl?: string
  files: CliSyncFile[]
}

export interface RunCliSyncInput {
  target: CliSyncAppID
  model?: string
}

export interface CliSyncFileInput {
  target: CliSyncAppID
  path: string
}

export interface SaveCliSyncFileInput {
  target: CliSyncAppID
  path: string
  content: string
}

export interface UpdateProxySettingsInput {
  port?: number
  allowLan?: boolean
  autoStartProxy?: boolean
  proxyApiKey?: string
  regenerateApiKey?: boolean
  authorizationMode?: boolean
  schedulingMode?: string
}

export interface ProxySettingsUpdateResult {
  restartedProxy: boolean
  restartedCloudflared: boolean
  generatedApiKey?: string
}

export interface UpdateCloudflaredSettingsInput {
  mode?: 'quick' | 'auth'
  token?: string
  useHttp2?: boolean
}

export type ProxyRuntimeAction = 'start' | 'stop' | 'toggle'

export type CloudflaredAction = 'install' | 'start' | 'stop' | 'refresh-status'

export interface LocalModelCatalogItem {
  id: string
  ownedBy: string
}

export interface EndpointTestRequest {
  baseUrl: string
  apiKey: string
  endpointId: string
  body?: string
}

export interface EndpointTestResult {
  status: string
  responseText: string
}
