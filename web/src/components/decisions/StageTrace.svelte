<script lang="ts">
  import type { StageResult } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    stages: StageResult[];
  }

  let { stages }: Props = $props();
</script>

<div class="stage-trace">
  {#each stages as stage, i}
    <div class="stage-block">
      <div class="stage-header">
        <span class="stage-name">{stage.stage}</span>
        <span class="stage-time">{stage.latency_ms}ms</span>
      </div>
      <div class="stage-body">
        {#if stage.error}
          <Badge variant="fail">{stage.error}</Badge>
        {:else}
          <span class="stage-count">{stage.items.length} items</span>
        {/if}
      </div>
    </div>
    {#if i < stages.length - 1}
      <div class="stage-connector">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
          <path d="M5 12h14M14 7l5 5-5 5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
    {/if}
  {/each}
</div>

<style>
  .stage-trace {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    flex-wrap: wrap;
  }

  .stage-block {
    padding: var(--space-sm) var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stage-header {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .stage-name {
    font-family: var(--font-code);
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--accent);
  }

  .stage-time {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
  }

  .stage-body {
    font-size: 0.75rem;
  }

  .stage-count {
    color: var(--text-secondary);
    font-family: var(--font-code);
  }

  .stage-connector {
    color: var(--text-tertiary);
    display: flex;
    align-items: center;
  }
</style>
