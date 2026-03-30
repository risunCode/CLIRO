<script lang="ts">
  import { ArrowRightLeft, Plus, Trash2 } from 'lucide-svelte'
  import Button from '@/components/common/Button.svelte'
  import CollapsibleSurfaceSection from '@/components/common/CollapsibleSurfaceSection.svelte'
  import { onMount } from 'svelte'

  export let busy = false
  export let onGetModelAliases: () => Promise<Record<string, string>>
  export let onSetModelAliases: (aliases: Record<string, string>) => Promise<void>

  let expanded = false
  let aliases: Array<{ from: string; to: string }> = []
  let loading = false
  let error = ''
  let dirty = false

  onMount(async () => {
    await loadAliases()
  })

  const loadAliases = async (): Promise<void> => {
    loading = true
    error = ''
    try {
      const data = await onGetModelAliases()
      aliases = Object.entries(data).map(([from, to]) => ({ from, to }))
  dirty = false
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load model aliases'
    } finally {
      loading = false
    }
  }

  const addAlias = (): void => {
    aliases = [...aliases, { from: '', to: '' }]
    dirty = true
  }

  const removeAlias = (index: number): void => {
    aliases = aliases.filter((_, i) => i !== index)
    dirty = true
  }

  const updateAlias = (index: number, field: 'from' | 'to', value: string): void => {
    const next = [...aliases]
    next[index][field] = value
    aliases = next
    dirty = true
  }

  const saveAliases = async (): Promise<void> => {
  error = ''

 // Validate
    const trimmed = aliases.map(a => ({ from: a.from.trim(), to: a.to.trim() }))
    const invalid = trimmed.some(a => a.from === '' || a.to === '')
    if (invalid) {
      error = 'All alias fields must be filled'
      return
    }

    // Check duplicates
    const fromSet = new Set(trimmed.map(a => a.from))
    if (fromSet.size !== trimmed.length) {
      error = 'Duplicate source model names found'
      return
    }

    loading = true
    try {
      const aliasMap: Record<string, string> = {}
      for (const { from, to } of trimmed) {
        aliasMap[from] = to
    }
      await onSetModelAliases(aliasMap)
      dirty = false
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save model aliases'
 } finally {
      loading = false
    }
  }

  const resetAliases = async (): Promise<void> => {
    await loadAliases()
  }
</script>

<CollapsibleSurfaceSection
  bind:open={expanded}
  icon={ArrowRightLeft}
  title="Model Mapping Alias"
  subtitle="Map model names to different providers (e.g., gpt-4 → claude-sonnet-4)"
  pill="Cross-Provider"
  ariaLabel="Toggle model alias mapping"
>
  <div class="model-alias-container">
    {#if loading && aliases.length === 0}
    <div class="loading-state">Loading aliases...</div>
    {:else}
      <div class="alias-list">
        {#each aliases as alias, index}
          <div class="alias-row">
          <input
              type="text"
 class="alias-input"
         placeholder="Source model (e.g., gpt-4)"
     value={alias.from}
 on:input={(e) => updateAlias(index, 'from', e.currentTarget.value)}
   disabled={loading || busy}
            />
   <span class="arrow">→</span>
     <input
          type="text"
              class="alias-input"
    placeholder="Target model (e.g., claude-sonnet-4)"
   value={alias.to}
           on:input={(e) => updateAlias(index, 'to', e.currentTarget.value)}
      disabled={loading || busy}
 />
            <button
         class="remove-btn"
      on:click={() => removeAlias(index)}
              disabled={loading || busy}
           aria-label="Remove alias"
            >
      <Trash2 size={16} />
        </button>
          </div>
        {/each}
      </div>

      <div class="actions">
     <Button
          variant="secondary"
   size="sm"
     on:click={addAlias}
       disabled={loading || busy}
        >
    <Plus size={16} />
          Add Alias
        </Button>

        {#if dirty}
          <div class="save-actions">
            <Button
     variant="secondary"
     size="sm"
     on:click={resetAliases}
    disabled={loading || busy}
            >
         Cancel
            </Button>
            <Button
        variant="primary"
     size="sm"
            on:click={saveAliases}
              disabled={loading || busy}
  >
      Save Changes
       </Button>
 </div>
        {/if}
      </div>

      {#if error}
        <div class="error-message">{error}</div>
      {/if}
    {/if}
  </div>
</CollapsibleSurfaceSection>

<style>
  .model-alias-container {
    display: flex;
flex-direction: column;
    gap: 1rem;
  }

  .loading-state {
    padding: 1rem;
    text-align: center;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }

  .alias-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .alias-row {
    display: flex;
align-items: center;
    gap: 0.75rem;
  }

  .alias-input {
    flex: 1;
    padding: 0.5rem 0.75rem;
  background: var(--surface-secondary);
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 0.875rem;
 font-family: 'JetBrains Mono', monospace;
    transition: border-color 0.2s;
}

  .alias-input:focus {
    outline: none;
    border-color: var(--accent-primary);
  }

  .alias-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .alias-input::placeholder {
  color: var(--text-tertiary);
  }

  .arrow {
    color: var(--text-secondary);
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .remove-btn {
    padding: 0.5rem;
background: transparent;
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    color: var(--text-secondary);
  cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .remove-btn:hover:not(:disabled) {
    background: var(--surface-hover);
    border-color: var(--error);
 color: var(--error);
  }

  .remove-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .actions {
    display: flex;
  align-items: center;
    gap: 0.75rem;
    padding-top: 0.5rem;
  }

  .save-actions {
    display: flex;
    gap: 0.5rem;
    margin-left: auto;
  }

  .error-message {
    padding: 0.75rem;
    background: var(--error-bg);
 border: 1px solid var(--error);
    border-radius: 6px;
    color: var(--error);
    font-size: 0.875rem;
  }
</style>
