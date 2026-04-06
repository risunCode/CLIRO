import * as appBindings from '../../../wailsjs/go/main/App'

export const wailsClient = {
  system: {
    getState: appBindings.GetState,
    getHostName: appBindings.GetHostName,
    runAction: appBindings.RunSystemAction,
    openExternalURL: appBindings.OpenExternalURL,
  },
  logs: {
    getLogs: appBindings.GetLogs,
  },
  auth: {
    startAuth: appBindings.StartAuth,
    startSocialAuth: appBindings.StartSocialAuth,
    getAuthSession: appBindings.GetAuthSession,
    cancelAuth: appBindings.CancelAuth,
    submitAuthCode: appBindings.SubmitAuthCode,
  },
  accounts: {
    getAccounts: appBindings.GetAccounts,
    importAccounts: appBindings.ImportAccounts,
    runAction: appBindings.RunAccountAction,
    runQuotaAction: appBindings.RunQuotaAction,
    syncAccountAuth: appBindings.SyncAccountAuth,
  },
  router: {
    getProxyStatus: appBindings.GetProxyStatus,
    getLocalModelCatalog: appBindings.GetLocalModelCatalog,
    getCliSyncStatuses: appBindings.GetCLISyncStatuses,
    runCliSync: appBindings.RunCLISync,
    getCliSyncFile: appBindings.GetCLISyncFile,
    saveCliSyncFile: appBindings.SaveCLISyncFile,
    updateProxySettings: appBindings.UpdateProxySettings,
    updateCloudflaredSettings: appBindings.UpdateCloudflaredSettings,
    runProxyAction: appBindings.RunProxyAction,
    runCloudflaredAction: appBindings.RunCloudflaredAction,
    getModelAliases: appBindings.GetModelAliases,
    setModelAliases: appBindings.SetModelAliases,
  },
}
