<script lang="ts">
  import type { DecisionResponse } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    decision: DecisionResponse;
    onreplay?: (id: string) => void;
  }

  let { decision, onreplay }: Props = $props();
</script>

<div class="decision-card">
  <div class="card-header">
    <code class="decision-hash">{decision.decision_hash}</code>
    <Badge variant={decision.dependency_status === 'healthy' ? 'pass' : 'degraded'}>
      {decision.dependency_status}
    </Badge>
  </div>

  <div class="card-meta">
    <span class="meta-item">ID: <code>{decision.decision_id}</code></span>
    <span class="meta-item">Latency: {decision.latency_ms}ms</span>
    <span class="meta-item">Items: {decision.ranked_items.length}</span>
  </div>

  <div class="stage-trace">
    {#each decision.stages as stage, i}
      <div class="stage">
        <span class="stage-name">{stage.stage}</span>
        <span class="stage-latency">{stage.latency_ms}ms</span>
        {#if stage.error}
          <Badge variant="fail">error</Badge>
        {:else}
          <Badge variant="pass">{stage.items.length} items</Badge>
        {/if}
      </div>
      {#if i < decision.stages.length - 1}
        <span class="stage-arrow">→</span>
      {/if}
    {/each}
  </div>

  <div class="card-items">
    {#each decision.ranked_items.slice(0, 5) as item}
      <div class="ranked-item">
        <span class="item-rank">#{item.rank}</span>
        <span class="item-id">{item.item_id}</span>
        <span class="item-score">{item.score.toFixed(4)}</span>
        <Badge variant="default">{item.source}</Badge>
      </div>
    {/each}
    {#if decision.ranked_items.length > 5}
      <span class="more-items">+{decision.ranked_items.length - 5} more</span>
    {/if}
  </div>

  {#if onreplay}
    <button class="replay-btn" onclick={() => onreplay(decision.decision_id)}>
      Replay
    </button>
  {/if}
</div>

<style>
  .decision-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-md);
  }

  .decision-hash {
    font-family: var(--font-code);
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--accent);
    letter-spacing: 0.01em;
  }

  .card-meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-md);
    font-size: 0.75rem;
    color: var(--text-tertiary);
  }

  .card-meta code {
    font-family: var(--font-code);
    color: var(--text-secondary);
  }

  .stage-trace {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: var(--space-sm);
    padding: var(--space-md);
    background: rgba(0, 0, 0, 0.2);
    border-radius: var(--radius-sm);
  }

  .stage {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }

  .stage-name {
    font-family: var(--font-code);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-primary);
    text-transform: uppercase;
  }

  .stage-latency {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
  }

  .stage-arrow {
    color: var(--text-tertiary);
    font-size: 0.75rem;
  }

  .card-items {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .ranked-item {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-size: 0.8125rem;
    padding: 4px 0;
  }

  .item-rank {
    font-family: var(--font-code);
    font-weight: 700;
    color: var(--accent);
    width: 28px;
  }

  .item-id {
    font-family: var(--font-code);
    color: var(--text-primary);
    flex: 1;
  }

  .item-score {
    font-family: var(--font-code);
    font-size: 0.75rem;
    color: var(--text-secondary);
  }

  .more-items {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    padding-left: 28px;
  }

  .replay-btn {
    align-self: flex-start;
    padding: 6px 16px;
    border-radius: var(--radius-pill);
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-secondary);
    transition: all var(--duration-fast) var(--ease-out);
  }
  .replay-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
  }
</style>
