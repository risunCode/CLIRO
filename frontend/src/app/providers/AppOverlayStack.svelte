<script lang="ts">
  import ToastViewport from '@/components/common/ToastViewport.svelte'
  import ConfigurationRecoveryModal from '@/app/modals/ConfigurationRecoveryModal.svelte'
  import UpdateRequiredModal from '@/app/modals/UpdateRequiredModal.svelte'
  import type { StartupWarningEntry } from '@/app/services/startup-warnings'

  export let showConfigurationErrorModal = false
  export let startupWarnings: StartupWarningEntry[] = []
  export let onDismissConfigurationErrorModal: () => void
  export let onOpenDataDir: () => Promise<void>
  export let showUpdatePrompt = false
  export let currentVersion = ''
  export let latestVersion = ''
  export let releaseName = ''
  export let releaseUrl = ''
  export let onDismissUpdatePrompt: () => void
  export let onOpenUpdateReleasePage: () => Promise<void>
</script>

<ToastViewport />

<ConfigurationRecoveryModal
  open={showConfigurationErrorModal}
  warnings={startupWarnings}
  on:dismiss={onDismissConfigurationErrorModal}
  on:openDataDir={onOpenDataDir}
/>

<UpdateRequiredModal
  open={showUpdatePrompt}
  {currentVersion}
  {latestVersion}
  {releaseName}
  {releaseUrl}
  on:dismiss={onDismissUpdatePrompt}
  on:openRelease={onOpenUpdateReleasePage}
/>
