<script lang="ts">
  import type { AppActions, AppOverlayState, SettingsActions } from '@/app/services/app-controller'
  import AppCloseModal from '@/app/modals/AppCloseModal.svelte'
  import ConfigurationRecoveryModal from '@/app/modals/ConfigurationRecoveryModal.svelte'
  import UpdateRequiredModal from '@/app/modals/UpdateRequiredModal.svelte'
  import ToastViewport from '@/components/common/ToastViewport.svelte'

  export let overlays: AppOverlayState
  export let appActions: AppActions
  export let settingsActions: SettingsActions
</script>

<ToastViewport />

<AppCloseModal
  open={overlays.showClosePrompt}
  trayAvailable={overlays.trayAvailable}
  armed={overlays.closePromptArmed}
  countdownSeconds={overlays.closePromptCountdown}
  on:dismiss={appActions.dismissClosePrompt}
  on:confirmQuit={appActions.confirmQuit}
  on:hideToTray={appActions.hideToTray}
/>

<ConfigurationRecoveryModal
  open={overlays.showConfigurationErrorModal}
  warnings={overlays.startupWarnings}
  on:dismiss={appActions.dismissConfigurationErrorModal}
  on:openDataDir={settingsActions.openDataDir}
/>

<UpdateRequiredModal
  open={overlays.showUpdatePrompt}
  currentVersion={overlays.updateInfo?.currentVersion || ''}
  latestVersion={overlays.updateInfo?.latestVersion || ''}
  releaseName={overlays.updateInfo?.releaseName || ''}
  releaseUrl={overlays.updateInfo?.releaseUrl || ''}
  on:dismiss={appActions.dismissUpdatePrompt}
  on:openRelease={appActions.openUpdateReleasePage}
/>
