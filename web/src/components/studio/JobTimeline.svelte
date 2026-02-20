<script lang="ts">
  import type { WorkloadItem } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    workload: WorkloadItem[];
  }

  let { workload }: Props = $props();

  function statusVariant(status: string): 'pass' | 'fail' | 'degraded' | 'default' {
    if (status === 'done' || status === 'complete') return 'pass';
    if (status === 'in_progress' || status === 'active') return 'degraded';
    if (status === 'failed') return 'fail';
    return 'default';
  }
</script>

<div class="timeline">
  {#each workload as item}
    <div class="timeline-phase" class:done={item.status === 'done' || item.status === 'complete'} class:active={item.status === 'in_progress' || item.status === 'active'}>
      <div class="phase-dot">
        {#if item.status === 'done' || item.status === 'complete'}&#10003;{:else if item.status === 'in_progress' || item.status === 'active'}&#9679;{:else}&#9675;{/if}
      </div>
      <div class="phase-content">
        <div class="phase-header">
          <span class="phase-name">{item.phase}</span>
          <Badge variant={statusVariant(item.status)}>{item.status}</Badge>
        </div>
        <p class="phase-task">{item.task}</p>
        <div class="phase-meta">
          <span class="meta-owner">{item.owner}</span>
          <span class="meta-estimate">{item.estimate_hours}h est.</span>
        </div>
      </div>
    </div>
  {/each}
  {#if workload.length === 0}
    <div class="empty-state">No workload items yet</div>
  {/if}
</div>

<style>
  .timeline {
    display: flex;
    flex-direction: column;
    gap: 0;
    position: relative;
    padding-left: var(--space-md);
  }

  .timeline::before {
    content: '';
    position: absolute;
    left: 11px;
    top: 12px;
    bottom: 12px;
    width: 2px;
    background: var(--border-subtle);
  }

  .timeline-phase {
    display: flex;
    align-items: flex-start;
    gap: var(--space-md);
    padding: var(--space-sm) 0;
    position: relative;
  }

  .phase-dot {
    flex-shrink: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    font-size: 0.6875rem;
    font-weight: 700;
    background: var(--bg-elevated);
    border: 2px solid var(--border-subtle);
    color: var(--text-tertiary);
    z-index: 1;
  }

  .done .phase-dot {
    background: var(--pass);
    border-color: var(--pass);
    color: #fff;
  }

  .active .phase-dot {
    background: var(--accent);
    border-color: var(--accent);
    color: #fff;
    box-shadow: 0 0 8px var(--accent-glow);
  }

  .phase-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 2px;
    flex: 1;
  }

  .phase-header {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .phase-name {
    font-size: 0.8125rem;
    font-weight: 600;
    text-transform: capitalize;
    color: var(--text-primary);
  }

  .phase-task {
    font-size: 0.75rem;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .phase-meta {
    display: flex;
    gap: var(--space-md);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
    font-family: var(--font-code);
  }

  .empty-state {
    color: var(--text-tertiary);
    font-style: italic;
    padding: var(--space-md);
  }
</style>
