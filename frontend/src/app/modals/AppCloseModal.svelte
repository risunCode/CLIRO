<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import BaseModal from '@/shared/components/BaseModal.svelte'
  import Button from '@/shared/components/Button.svelte'
  import ModalWindowHeader from '@/shared/components/ModalWindowHeader.svelte'

  export let open = false
  export let trayAvailable = false
  export let armed = false
  export let countdownSeconds = 0

  $: countdownLabel = countdownSeconds > 0 ? `${countdownSeconds} detik` : 'sebentar'

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
    <ModalWindowHeader
      title="Close CLIro-Go"
      description={armed ? `App akan tertutup otomatis dalam ${countdownLabel}. Tekan X sekali lagi untuk langsung keluar.` : 'Choose whether to fully close the app or keep it running in the tray.'}
    />
  </svelte:fragment>

  {#if armed}
    <p class="text-xs text-text-secondary">Auto-close sudah aktif. Klik tombol X satu kali lagi untuk langsung menutup aplikasi sekarang.</p>
  {:else}
    <p class="text-xs text-text-secondary">Closing the app stops local services. Minimizing keeps the proxy running in the background.</p>
  {/if}
  {#if !trayAvailable}
    <p class="text-xs text-text-secondary">System tray is unavailable on this device.</p>
  {/if}

  <svelte:fragment slot="footer">
    <Button variant="secondary" size="sm" disabled={!trayAvailable} on:click={() => dispatch('hideToTray')}>Minimize to Tray</Button>
    <Button variant="danger" size="sm" on:click={() => dispatch('confirmQuit')}>{armed ? 'Close Now' : 'Close App'}</Button>
  </svelte:fragment>
</BaseModal>
