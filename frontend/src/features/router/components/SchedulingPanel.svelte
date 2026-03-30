<script lang="ts">
  import { RefreshCw } from 'lucide-svelte'
  import Button from '@/components/common/Button.svelte'
  import CollapsibleSurfaceSection from '@/components/common/CollapsibleSurfaceSection.svelte'
  import ToggleSwitch from '@/components/common/ToggleSwitch.svelte'
  import type { ProxyStatus } from '@/features/router/types'
  import { SCHEDULING_MODE_CARDS, normalizeCircuitSteps, toSchedulingMode, type SchedulingMode } from '@/features/router/lib/scheduling'

  export let proxyStatus: ProxyStatus | null = null
  export let busy = false
  export let onSetSchedulingMode: (mode: string) => Promise<void>
  export let onSetCircuitBreaker: (enabled: boolean) => Promise<void>
  export let onSetCircuitSteps: (steps: number[]) => Promise<void>

  let expanded = false
  let schedulingModeInput: SchedulingMode = 'balance'
  let circuitBreakerInput = false
  let circuitStepInputs = ['10', '30', '60']
  let circuitStepsDirty = false
  let schedulingError = ''

  $: if (proxyStatus && !busy) {
    schedulingModeInput = toSchedulingMode(proxyStatus.schedulingMode)
    circuitBreakerInput = proxyStatus.circuitBreaker
    const normalizedSteps = normalizeCircuitSteps(proxyStatus.circuitSteps)
    if (!circuitStepsDirty) {
      circuitStepInputs = normalizedSteps.map((value) => String(value))
    }
  }

  const applySchedulingMode = async (mode: SchedulingMode): Promise<void> => {
    if (mode === schedulingModeInput) {
      return
    }
    schedulingModeInput = mode
    await onSetSchedulingMode(mode)
  }

  const updateCircuitBreaker = async (): Promise<void> => {
    schedulingError = ''
    await onSetCircuitBreaker(circuitBreakerInput)
  }

  const updateCircuitStepInput = (index: number, value: string): void => {
    const next = [...circuitStepInputs]
    next[index] = value
    circuitStepInputs = next
    circuitStepsDirty = true
  }

  const handleCircuitStepInput = (index: number, event: Event): void => {
    const target = event.currentTarget as HTMLInputElement
    updateCircuitStepInput(index, target.value)
  }

  const applyCircuitSteps = async (): Promise<void> => {
    const parsed = circuitStepInputs.map((value) => Number.parseInt(value.trim(), 10))
    const invalid = parsed.some((value) => !Number.isFinite(value) || value <= 0 || value > 3600)
    if (invalid) {
      schedulingError = 'Each circuit breaker step must be 1-3600 seconds.'
      return
    }

    schedulingError = ''
    const normalized = parsed.map((value) => Math.round(value))
    await onSetCircuitSteps(normalized)
    circuitStepInputs = normalized.map((value) => String(value))
    circuitStepsDirty = false
  }
</script>

<CollapsibleSurfaceSection
  bind:open={expanded}
  icon={RefreshCw}
  title="Account Scheduling & Rotation"
  subtitle="Control account routing mode and staged failure backoff behavior."
  pill="Routing Policy"
  ariaLabel="Toggle account scheduling and rotation"
>
      <div class="api-rotation-grid">
        <div class="api-rotation-modes">
          <p class="api-endpoint-label">Scheduling Mode</p>
          <div class="api-rotation-mode-list">
            {#each SCHEDULING_MODE_CARDS as card}
              <button
                type="button"
                class={`api-rotation-mode-card ${schedulingModeInput === card.id ? 'is-active' : ''}`}
                on:click={() => void applySchedulingMode(card.id)}
                disabled={busy}
              >
                <span class="api-rotation-mode-title">{card.label}</span>
                <span class="api-rotation-mode-desc">{card.description}</span>
              </button>
            {/each}
          </div>
        </div>

        <div class="api-rotation-side">
          <div class="api-rotation-side-card api-rotation-info ui-panel-soft">
            <p>
              {#if schedulingModeInput === 'cache_first'}
                Prioritizes bound sessions to maximize cache hit continuity.
              {:else if schedulingModeInput === 'balance'}
                Favors accounts with lower request/error load for balanced utilization.
              {:else}
                Uses round-robin order for low-latency throughput at high concurrency.
              {/if}
            </p>
          </div>

          <div class="api-rotation-side-card api-rotation-breaker ui-panel-soft">
            <ToggleSwitch
              label="Circuit Breaker (staged cooldown after repeated failures)"
              bind:checked={circuitBreakerInput}
              on:change={updateCircuitBreaker}
              disabled={busy}
            />
            <p class="api-rotation-footnote">Steps apply in order: #1, #2, #3, then remain on step #3.</p>
            <p class="api-rotation-footnote">Exhausted/usage-limit errors skip to the next account and do not consume circuit steps.</p>
          </div>

          <div class="api-rotation-side-card api-rotation-steps ui-panel-soft">
            <p class="api-endpoint-label">Circuit Steps (seconds)</p>
            <div class="api-rotation-steps-grid">
              <label class="api-rotation-step-item">
                <span>Step 1</span>
                <input
                  type="number"
                  min="1"
                  max="3600"
                  class="ui-control-input ui-control-select-sm"
                  value={circuitStepInputs[0]}
                  on:input={(event) => handleCircuitStepInput(0, event)}
                  disabled={busy}
                />
              </label>

              <label class="api-rotation-step-item">
                <span>Step 2</span>
                <input
                  type="number"
                  min="1"
                  max="3600"
                  class="ui-control-input ui-control-select-sm"
                  value={circuitStepInputs[1]}
                  on:input={(event) => handleCircuitStepInput(1, event)}
                  disabled={busy}
                />
              </label>

              <label class="api-rotation-step-item">
                <span>Step 3</span>
                <input
                  type="number"
                  min="1"
                  max="3600"
                  class="ui-control-input ui-control-select-sm"
                  value={circuitStepInputs[2]}
                  on:input={(event) => handleCircuitStepInput(2, event)}
                  disabled={busy}
                />
              </label>
            </div>

            {#if schedulingError}
              <p class="api-rotation-error">{schedulingError}</p>
            {/if}

            <Button variant="secondary" size="sm" on:click={applyCircuitSteps} disabled={busy || !circuitStepsDirty}>
              Apply Steps
            </Button>
          </div>
        </div>
        </div>
</CollapsibleSurfaceSection>
