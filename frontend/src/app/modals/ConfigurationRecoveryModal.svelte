<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import BaseModal from '@/shared/components/BaseModal.svelte'
  import Button from '@/shared/components/Button.svelte'
  import ModalWindowHeader from '@/shared/components/ModalWindowHeader.svelte'

  export let open = false
  export let warnings: Array<{
    code: string
    filePath: string
    backupPath?: string
    message: string
  }> = []

  const dispatch = createEventDispatcher<{ dismiss: void; openDataDir: void }>()
</script>

<BaseModal
  {open}
  overlayClass="items-center justify-center p-4"
  cardClass="w-full max-w-2xl overflow-hidden"
  headerClass="border-b border-border px-5 py-4"
  bodyClass="space-y-3 px-5 py-4 text-sm text-text-secondary"
  footerClass="flex items-center justify-end gap-2 border-t border-border px-5 py-4"
  on:close={() => dispatch('dismiss')}
>
  <svelte:fragment slot="header">
    <ModalWindowHeader
      title="Configuration Recovered"
      description="CLIro-Go detected invalid local configuration data, recovered safe defaults, and kept the app running."
    />
  </svelte:fragment>

  <div class="space-y-3">
    {#each warnings as warning, index (warning.filePath + warning.code + index)}
      <div class="rounded-sm border border-border bg-app p-3 text-xs">
        <p class="font-semibold text-text-primary">{warning.filePath}</p>
        {#if warning.backupPath}
          <p class="mt-1 break-all text-text-secondary">Backup: {warning.backupPath}</p>
        {/if}
        <p class="mt-2 leading-relaxed text-text-secondary">{warning.message}</p>
      </div>
    {/each}
  </div>

  <p class="text-xs text-text-secondary">Review the recovered files in your local data folder if you want to restore or inspect the corrupted originals.</p>

  <svelte:fragment slot="footer">
    <Button variant="secondary" size="sm" on:click={() => dispatch('openDataDir')}>Open Data Folder</Button>
    <Button variant="primary" size="sm" on:click={() => dispatch('dismiss')}>Understood</Button>
  </svelte:fragment>
</BaseModal>
