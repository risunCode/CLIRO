<script lang="ts">
  import { Database, FolderOpen } from 'lucide-svelte'
  import Button from '@/components/common/Button.svelte'
  import SurfaceCard from '@/components/common/SurfaceCard.svelte'

  export let busy = false
  export let dataDirPath = '~/.cliro-go'
  export let onOpenDataDir: () => Promise<void>
</script>

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
