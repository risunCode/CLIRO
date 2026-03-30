<script lang="ts">
  import { Database, FolderOpen, Upload } from 'lucide-svelte'
  import Button from '@/components/common/Button.svelte'
  import SurfaceCard from '@/components/common/SurfaceCard.svelte'
  import type { AppState } from '@/app/types'
  import type { Account } from '@/features/accounts/types'

  interface BackupPayload {
    version: number
    exportedAt: string
    state: AppState | null
    accounts: Account[]
  }

  interface RestoreProgress {
    step: string
    index: number
    total: number
  }

  export let onOpenDataDir: () => Promise<void>
  export let onExportBackup: () => Promise<void>
  export let onRestoreBackup: (payload: BackupPayload, onProgress?: (progress: RestoreProgress) => void) => Promise<void>

  let backupFileInput: HTMLInputElement | null = null
  let busy = false
  let statusMessage = ''
  let errorMessage = ''

  const setBusy = async (action: () => Promise<void>, successMessage = 'Settings saved successfully.'): Promise<void> => {
    busy = true
    errorMessage = ''
    statusMessage = ''
    try {
      await action()
      statusMessage = successMessage
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : 'Operation failed.'
    } finally {
      busy = false
    }
  }

  const isRecord = (value: unknown): value is Record<string, unknown> => {
    return typeof value === 'object' && value !== null
  }

  const validateBackupPayload = (value: unknown): BackupPayload => {
    if (!isRecord(value)) throw new Error('Backup payload must be a JSON object.')

    const version = Number(value.version)
    if (!Number.isFinite(version) || version <= 0) throw new Error('Backup payload version is invalid.')

    const rawState = value.state
    const state = rawState === null || rawState === undefined ? null : (isRecord(rawState) ? (rawState as unknown as AppState) : null)
    if (rawState !== null && rawState !== undefined && !isRecord(rawState)) {
      throw new Error('Backup payload state must be an object or null.')
    }

    if (!Array.isArray(value.accounts)) throw new Error('Backup payload accounts must be an array.')

    const accounts = value.accounts.filter((entry) => isRecord(entry)) as unknown as Account[]
    const exportedAt = typeof value.exportedAt === 'string' ? value.exportedAt : new Date().toISOString()

    return { version, exportedAt, state, accounts }
  }

  const handleRestoreFromFile = async (event: Event): Promise<void> => {
    const target = event.currentTarget as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    await setBusy(
      async () => {
        const text = await file.text()
        const parsedPayload = JSON.parse(text) as unknown
        const payload = validateBackupPayload(parsedPayload)

        await onRestoreBackup(payload, (progress) => {
          statusMessage = `Restoring step ${progress.index}/${progress.total}: ${progress.step}`
        })
      },
      'Backup restored successfully.'
    )

    target.value = ''
  }

  const handleExportBackup = async (): Promise<void> => {
    await setBusy(async () => {
      await onExportBackup()
    }, 'Backup exported successfully.')
  }

  const dataDirPath = '~/.cliro-go'
</script>

<div class="settings-tab space-y-2.5">
  <!-- Data Folder + Backup Tools -->
  <div class="grid gap-2.5 lg:grid-cols-2">
    <SurfaceCard className="p-3.5">
      <div class="mb-3 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Database size={15} class="text-text-secondary" />
          <p class="text-sm font-semibold text-text-primary">Data Folder</p>
        </div>
        <Button variant="secondary" size="sm" on:click={() => void onOpenDataDir()} disabled={busy}>
          <FolderOpen size={13} class="mr-1" />
          Open
        </Button>
      </div>

      <div class="rounded border border-border bg-app p-2 font-mono text-xs text-text-secondary">{dataDirPath}</div>
    </SurfaceCard>

    <SurfaceCard className="p-3.5">
      <div class="mb-3 flex items-center gap-2">
        <Upload size={15} class="text-text-secondary" />
        <p class="text-sm font-semibold text-text-primary">Backup Tools</p>
      </div>

      <div class="flex flex-wrap gap-2">
        <Button variant="secondary" size="sm" on:click={() => void handleExportBackup()} disabled={busy}>
          <Database size={13} class="mr-1" />
          Export Backup
        </Button>
        <Button
          variant="secondary"
          size="sm"
          on:click={() => {
            backupFileInput?.click()
          }}
          disabled={busy}
        >
          <Upload size={13} class="mr-1" />
          Restore Backup
        </Button>
      </div>

      {#if statusMessage}
        <p class="mt-2 text-xs text-success">{statusMessage}</p>
      {/if}
      {#if errorMessage}
        <p class="mt-2 text-xs text-error">{errorMessage}</p>
      {/if}

      <input bind:this={backupFileInput} type="file" accept=".json,application/json" class="hidden" on:change={handleRestoreFromFile} />
    </SurfaceCard>
  </div>
</div>
