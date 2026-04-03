<script lang="ts">
  import { Database, Upload } from 'lucide-svelte'
  import Button from '@/shared/components/Button.svelte'
  import SurfaceCard from '@/shared/components/SurfaceCard.svelte'

  export let busy = false
  export let statusMessage = ''
  export let errorMessage = ''
  export let onExportBackup: () => Promise<void>
  export let onRestoreBackupFile: (file: File) => Promise<void>

  let backupFileInput: HTMLInputElement | null = null

  const triggerRestore = (): void => {
    backupFileInput?.click()
  }

  const handleFileChange = async (event: Event): Promise<void> => {
    const target = event.currentTarget as HTMLInputElement
    const file = target.files?.[0]
    if (!file) {
      return
    }

    await onRestoreBackupFile(file)
    target.value = ''
  }
</script>

<SurfaceCard className="p-3.5">
  <div class="mb-3 flex items-center gap-2">
    <Upload size={15} class="text-text-secondary" />
    <p class="text-sm font-semibold text-text-primary">Backup Tools</p>
  </div>

  <div class="flex flex-wrap gap-2">
    <Button variant="secondary" size="sm" on:click={() => void onExportBackup()} disabled={busy}>
      <Database size={13} class="mr-1" />
      Export Backup
    </Button>
    <Button variant="secondary" size="sm" on:click={triggerRestore} disabled={busy}>
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

  <input bind:this={backupFileInput} type="file" accept=".json,application/json" class="hidden" on:change={handleFileChange} />
</SurfaceCard>
