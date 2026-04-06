import type { auth, config, logger, main } from '../../../wailsjs/go/models'

export type WailsAppState = main.State
export type WailsLogEntry = logger.Entry
export type WailsAccount = config.Account
export type WailsProxyStatus = main.ProxyStatus
export type WailsModelCatalogItem = main.ModelCatalogItem
export type WailsCliSyncStatus = main.CLISyncStatus
export type WailsCliSyncResult = main.CLISyncResult

export type WailsAuthStart = auth.AuthStart
export type WailsAuthSessionView = auth.AuthSessionView
export type WailsAuthSyncResult = auth.AuthSyncResult
