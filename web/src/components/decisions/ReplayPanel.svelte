<script lang="ts">
  import type { ReplayResponse } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    replay: ReplayResponse;
  }

  let { replay }: Props = $props();
</script>

<div class="replay-panel">
  <div class="replay-header">
    <h3 class="replay-title">Replay Comparison</h3>
    <Badge variant={replay.hashes_match ? 'pass' : 'fail'}>
      {replay.hashes_match ? 'Hashes Match' : 'Hash Mismatch'}
    </Badge>
  </div>

  <div class="replay-comparison">
    <div class="replay-column">
      <span class="column-label">Original</span>
      <code class="hash" class:match={replay.hashes_match}>{replay.original.decision_hash}</code>
      <div class="column-items">
        {#each replay.original.ranked_items as item}
          <div class="item-row">
            <span class="rank">#{item.rank}</span>
            <span class="id">{item.item_id}</span>
            <span class="score">{item.score.toFixed(4)}</span>
          </div>
        {/each}
      </div>
    </div>

    <div class="replay-divider">
      <span class="divider-icon">{replay.hashes_match ? '=' : '≠'}</span>
    </div>

    <div class="replay-column">
      <span class="column-label">Replayed</span>
      <code class="hash" class:match={replay.hashes_match}>{replay.replayed.decision_hash}</code>
      <div class="column-items">
        {#each replay.replayed.ranked_items as item}
          <div class="item-row">
            <span class="rank">#{item.rank}</span>
            <span class="id">{item.item_id}</span>
            <span class="score">{item.score.toFixed(4)}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>

<style>
  .replay-panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .replay-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .replay-title {
    font-family: var(--font-display);
    font-size: 1.125rem;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .replay-comparison {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    gap: var(--space-md);
  }

  .replay-column {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .column-label {
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-tertiary);
  }

  .hash {
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--text-secondary);
    word-break: break-all;
    padding: var(--space-sm);
    background: rgba(0, 0, 0, 0.2);
    border-radius: var(--radius-sm);
  }

  .hash.match {
    color: var(--pass);
    background: var(--pass-subtle);
  }

  .replay-divider {
    display: flex;
    align-items: center;
    padding-top: var(--space-xl);
  }

  .divider-icon {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-tertiary);
  }

  .column-items {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .item-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-family: var(--font-code);
    font-size: 0.75rem;
    padding: 2px 0;
  }

  .rank {
    color: var(--accent);
    font-weight: 600;
    width: 24px;
  }

  .id {
    color: var(--text-primary);
    flex: 1;
  }

  .score {
    color: var(--text-tertiary);
  }

  @media (max-width: 640px) {
    .replay-comparison {
      grid-template-columns: 1fr;
    }
    .replay-divider {
      padding: 0;
      justify-content: center;
    }
  }
</style>
