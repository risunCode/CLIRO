<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import BaseModal from '@/components/common/BaseModal.svelte'
  import Button from '@/components/common/Button.svelte'
  import ModalWindowHeader from '@/components/common/ModalWindowHeader.svelte'

  export let open = false
  export let trayAvailable = false

  const dispatch = createEventDispatcher<{ dismiss: void; confirmQuit: void; hideToTray: void }>()
</script>

<BaseModal
  {open}
  overlayClass="items-center justify-center p-4"
  cardClass="w-full max-w-lg overflow-hidden"
  headerClass="border-b border-border px-5 py-4"
  bodyClass="space-y-3 px-5 py-4 text-sm text-text-secondary"
  footerClass="flex items-center justify-end gap-2 border-t border-border px-5 py-4"
  on:close={() => dispatch('dismiss')}
>
  <svelte:fragment slot="header">
    <ModalWindowHeader title="Close CLIro-Go" description="Choose whether to fully close the app or keep it running in the tray." />
  </svelte:fragment>

  <p class="text-xs text-text-secondary">Closing the app stops local services. Minimizing keeps the proxy running in the background.</p>
  {#if !trayAvailable}
    <p class="text-xs text-text-secondary">System tray is unavailable on this device.</p>
  {/if}

  <svelte:fragment slot="footer">
    <Button variant="secondary" size="sm" disabled={!trayAvailable} on:click={() => dispatch('hideToTray')}>Minimize to Tray</Button>
    <Button variant="danger" size="sm" on:click={() => dispatch('confirmQuit')}>Close App</Button>
  </svelte:fragment>
</BaseModal>
