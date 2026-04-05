import type { auth, config, logger, main } from '../../../wailsjs/go/models'

export type WailsAppState = main.State
export type WailsLogEntry = logger.Entry
export type WailsAccount = config.Account
export type WailsProxyStatus = main.ProxyStatus
export type WailsModelCatalogItem = main.ModelCatalogItem
export type WailsCliSyncStatus = main.CLISyncStatus
export type WailsCliSyncResult = main.CLISyncResult
export type WailsRunCliSyncInput = main.RunCLISyncInput
export type WailsCliSyncFileInput = main.CLISyncFileInput
export type WailsSaveCliSyncFileInput = main.SaveCLISyncFileInput
export type WailsUpdateProxySettingsInput = main.UpdateProxySettingsInput
export type WailsProxySettingsUpdateResult = main.ProxySettingsUpdateResult
export type WailsUpdateCloudflaredSettingsInput = main.UpdateCloudflaredSettingsInput
export type WailsRunAccountActionInput = main.RunAccountActionInput
export type WailsRunQuotaActionInput = main.RunQuotaActionInput
export type WailsRunSystemActionInput = main.RunSystemActionInput

export type WailsAuthStart = auth.AuthStart
export type WailsAuthSessionView = auth.AuthSessionView
export type WailsAuthSyncResult = auth.AuthSyncResult

// Backward-compatible aliases for feature types.
export type WailsCodexAuthStart = WailsAuthStart
export type WailsCodexAuthSessionView = WailsAuthSessionView
export type WailsKiroAuthStart = WailsAuthStart
