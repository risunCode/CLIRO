import type { WailsAccount, WailsAuthSessionView } from '@/backend/models/wails'

export type Account = WailsAccount

export type AuthSession = WailsAuthSessionView

export type KiroAuthSession = WailsAuthSessionView

export type SyncTargetID = 'kilo-cli' | 'opencode-cli' | 'codex-cli'

export type AccountAction =
  | 'refresh'
  | 'refresh-with-quota'
  | 'enable'
  | 'disable'
  | 'delete'
  | 'clear-cooldown'

export type QuotaAction = 'refresh-one' | 'refresh-all' | 'force-refresh-all'

export interface SyncResultBase {
  targetPath: string
  fileExisted: boolean
  updatedFields: string[]
  accountID: string
  provider: string
}

export interface KiloAuthSyncResult extends SyncResultBase {
  target: 'kilo-cli'
  openAICreated: boolean
  syncedExpires: number
  syncedExpiresAt?: string
}

export interface OpencodeAuthSyncResult extends SyncResultBase {
  target: 'opencode-cli'
  openAICreated: boolean
  syncedExpires: number
  syncedExpiresAt?: string
}

export interface CodexAuthSyncResult extends SyncResultBase {
  target: 'codex-cli'
  backupPath?: string
  backupCreated: boolean
  syncedAt?: string
}

export type AccountSyncResult = KiloAuthSyncResult | OpencodeAuthSyncResult | CodexAuthSyncResult
