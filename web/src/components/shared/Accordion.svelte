<script lang="ts">
  interface Props {
    title: string;
    open?: boolean;
    children: import('svelte').Snippet;
  }

  let { title, open = false, children }: Props = $props();

  let expanded = $state(open);
</script>

<div class="accordion" class:expanded>
  <button class="accordion-header" onclick={() => expanded = !expanded}>
    <span class="accordion-title">{title}</span>
    <span class="accordion-chevron">{expanded ? '−' : '+'}</span>
  </button>
  {#if expanded}
    <div class="accordion-body">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .accordion {
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .accordion-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: var(--space-md);
    background: var(--bg-surface);
    text-align: left;
    transition: background var(--duration-fast) var(--ease-out);
  }
  .accordion-header:hover {
    background: var(--bg-surface-hover);
  }

  .accordion-title {
    font-family: var(--font-body);
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .accordion-chevron {
    font-size: 1rem;
    color: var(--text-tertiary);
    width: 20px;
    text-align: center;
  }

  .accordion-body {
    padding: var(--space-md);
    border-top: 1px solid var(--border-subtle);
  }
</style>
