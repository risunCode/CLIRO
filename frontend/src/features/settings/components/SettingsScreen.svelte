<script lang="ts">
  import type { SettingsActions } from '@/app/services/app-controller'
  import BackupToolsCard from '@/features/settings/components/BackupToolsCard.svelte'
  import DataFolderCard from '@/features/settings/components/DataFolderCard.svelte'
  import { validateBackupPayload } from '@/features/settings/utils/backup'
  import { deriveSettingsViewState } from '@/features/settings/store/task-state'
  import { createAsyncTaskState, runAsyncTask, type AsyncTaskState } from '@/shared/utils/async'

  export let settingsActions: SettingsActions

  let task: AsyncTaskState = createAsyncTaskState()
  let statusMessage = ''

  $: viewState = deriveSettingsViewState(task, statusMessage)

  const setBusy = async (action: () => Promise<void>, successMessage = 'Settings saved successfully.'): Promise<void> => {
    statusMessage = ''

    try {
      await runAsyncTask((nextState) => {
        task = nextState
      }, action)
      statusMessage = successMessage
    } catch {
      statusMessage = ''
    }
  }

  const handleRestoreFromFile = async (file: File): Promise<void> => {
    await setBusy(
      async () => {
        const text = await file.text()
        const parsedPayload = JSON.parse(text) as unknown
        const payload = validateBackupPayload(parsedPayload)

        await settingsActions.restoreBackup(payload, (progress) => {
          statusMessage = `Restoring step ${progress.index}/${progress.total}: ${progress.step}`
        })
      },
      'Backup restored successfully.'
    )
  }

  const handleExportBackup = async (): Promise<void> => {
    await setBusy(async () => {
      await settingsActions.exportBackup()
    }, 'Backup exported successfully.')
  }
</script>

<div class="settings-tab space-y-2.5">
  <div class="grid gap-2.5 lg:grid-cols-2">
    <DataFolderCard busy={viewState.busy} onOpenDataDir={settingsActions.openDataDir} />

    <BackupToolsCard
      busy={viewState.busy}
      statusMessage={viewState.statusMessage}
      errorMessage={viewState.errorMessage}
      onExportBackup={handleExportBackup}
      onRestoreBackupFile={handleRestoreFromFile}
    />
  </div>
</div>
