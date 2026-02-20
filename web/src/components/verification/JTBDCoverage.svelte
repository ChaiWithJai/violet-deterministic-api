<script lang="ts">
  import type { JTBDCoverage as JTBDType } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    items: JTBDType[];
  }

  let { items }: Props = $props();
</script>

<div class="jtbd-coverage">
  <h3 class="jtbd-title">JTBD Coverage</h3>
  <div class="jtbd-grid">
    {#each items as item}
      <div class="jtbd-card" class:covered={item.covered}>
        <Badge variant={item.covered ? 'pass' : 'default'}>
          {item.covered ? 'Covered' : 'Pending'}
        </Badge>
        <span class="jtbd-name">{item.jtbd}</span>
        {#if item.evidence}
          <p class="jtbd-evidence">{item.evidence}</p>
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .jtbd-coverage {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .jtbd-title {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 600;
  }

  .jtbd-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: var(--space-sm);
  }

  .jtbd-card {
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .jtbd-card.covered {
    border-color: rgba(34, 197, 94, 0.2);
  }

  .jtbd-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .jtbd-evidence {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    line-height: 1.4;
  }
</style>
