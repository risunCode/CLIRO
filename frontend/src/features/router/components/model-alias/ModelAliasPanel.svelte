<script lang="ts">
  import { onMount } from 'svelte'
  import { ArrowRightLeft, Plus, Trash2 } from 'lucide-svelte'
  import Button from '@/shared/components/Button.svelte'
  import CollapsibleSurfaceSection from '@/shared/components/CollapsibleSurfaceSection.svelte'
  import { aliasRowsFromRecord, cloneAliasRows, createEmptyAliasRow, serializeAliasRows, type AliasFormRow } from '@/features/router/lib/alias-form'

  export let busy = false
  export let onGetModelAliases: () => Promise<Record<string, string>>
  export let onSetModelAliases: (aliases: Record<string, string>) => Promise<void>

  let expanded = false
  let aliases: AliasFormRow[] = []
  let savedAliases: AliasFormRow[] = []
  let loading = false
  let error = ''
  let dirty = false
  let loadRequestId = 0

  onMount(() => {
    void loadAliases().catch(() => {})
  })

  const loadAliases = async (): Promise<void> => {
    const requestId = ++loadRequestId
    loading = true
    error = ''

    try {
      const rows = aliasRowsFromRecord(await onGetModelAliases())
      if (requestId !== loadRequestId) {
        return
      }
      savedAliases = cloneAliasRows(rows)
      if (!dirty) {
        aliases = cloneAliasRows(rows)
      }
    } catch (loadError) {
      if (requestId !== loadRequestId) {
        return
      }
      error = loadError instanceof Error ? loadError.message : 'Failed to load model aliases'
    } finally {
      if (requestId === loadRequestId) {
        loading = false
      }
    }
  }

  const addAlias = (): void => {
    aliases = [...aliases, createEmptyAliasRow()]
    dirty = true
  }

  const removeAlias = (index: number): void => {
    aliases = aliases.filter((_, currentIndex) => currentIndex !== index)
    dirty = true
  }

  const updateAlias = (index: number, field: 'from' | 'to', value: string): void => {
    aliases = aliases.map((alias, currentIndex) => {
      if (currentIndex !== index) {
        return alias
      }
      return {
        ...alias,
        [field]: value
      }
    })
    dirty = true
    error = ''
  }

  const saveAliases = async (): Promise<void> => {
    error = ''
    loading = true

    try {
      const aliasMap = serializeAliasRows(aliases)
      await onSetModelAliases(aliasMap)
      savedAliases = aliasRowsFromRecord(aliasMap)
      aliases = cloneAliasRows(savedAliases)
      dirty = false
    } catch (saveError) {
      error = saveError instanceof Error ? saveError.message : 'Failed to save model aliases'
    } finally {
      loading = false
    }
  }

  const resetAliases = (): void => {
    aliases = cloneAliasRows(savedAliases)
    dirty = false
    error = ''
  }
</script>

<CollapsibleSurfaceSection
  bind:open={expanded}
  icon={ArrowRightLeft}
  title="Model Mapping Alias"
  subtitle="Map model names to different providers (e.g., gpt-4 -> claude-sonnet-4)"
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
              on:input={(event) => updateAlias(index, 'from', event.currentTarget.value)}
              disabled={loading || busy}
            />
            <span class="arrow">-></span>
            <input
              type="text"
              class="alias-input"
              placeholder="Target model (e.g., claude-sonnet-4)"
              value={alias.to}
              on:input={(event) => updateAlias(index, 'to', event.currentTarget.value)}
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
        <Button variant="secondary" size="sm" on:click={addAlias} disabled={loading || busy}>
          <Plus size={16} />
          Add Alias
        </Button>

        {#if dirty}
          <div class="save-actions">
            <Button variant="secondary" size="sm" on:click={resetAliases} disabled={loading || busy}>
              Cancel
            </Button>
            <Button variant="primary" size="sm" on:click={saveAliases} disabled={loading || busy}>
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
    color: var(--color-text-secondary);
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
    background: color-mix(in srgb, var(--color-surface) 86%, var(--color-app));
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text-primary);
    font-size: 0.875rem;
    font-family: 'JetBrains Mono', monospace;
    transition: border-color 0.2s;
  }

  .alias-input:focus {
    outline: none;
    border-color: var(--color-accent-primary);
  }

  .alias-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .alias-input::placeholder {
    color: var(--color-text-muted);
  }

  .arrow {
    color: var(--color-text-secondary);
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .remove-btn {
    padding: 0.5rem;
    background: transparent;
    border: 1px solid var(--color-border);
    border-radius: 6px;
    color: var(--color-text-secondary);
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .remove-btn:hover:not(:disabled) {
    background: color-mix(in srgb, var(--color-surface-hover) 82%, var(--color-app));
    border-color: var(--color-error);
    color: var(--color-error);
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
    background: color-mix(in srgb, var(--color-error) 10%, var(--color-app));
    border: 1px solid color-mix(in srgb, var(--color-error) 45%, var(--color-border));
    border-radius: 6px;
    color: var(--color-error);
    font-size: 0.875rem;
  }
</style>
