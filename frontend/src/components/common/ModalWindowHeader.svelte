<script lang="ts">
  export let title = ''
  export let description = ''
  export let titleClassName = 'text-base font-semibold text-text-primary'
  export let descriptionClassName = 'mt-1 text-sm text-text-secondary'
</script>

<div class="flex items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <div class="mb-2 flex items-center gap-2">
      <span class="h-3 w-3 rounded-full bg-[#ef4444]" aria-hidden="true" />
      <span class="h-3 w-3 rounded-full bg-[#f59e0b]" aria-hidden="true" />
      <span class="h-3 w-3 rounded-full bg-[#22c55e]" aria-hidden="true" />
    </div>

    <h2 class={`${titleClassName} min-w-0 break-words`}>{title}</h2>

    {#if description}
      <p class={descriptionClassName}>{description}</p>
    {/if}
  </div>

  {#if $$slots.actions}
    <div class="flex shrink-0 items-center gap-2">
      <slot name="actions" />
    </div>
  {/if}
</div>
