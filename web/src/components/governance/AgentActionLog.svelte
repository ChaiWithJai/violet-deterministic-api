<script lang="ts">
  import Badge from '../shared/Badge.svelte';

  interface ActionEntry {
    id: string;
    action: string;
    actor: 'human' | 'agent';
    description: string;
    timestamp: string;
    status: string;
  }

  interface Props {
    actions: ActionEntry[];
  }

  let { actions }: Props = $props();

  const phaseOrder = ['plan', 'clarify', 'act', 'verify', 'deploy'];
</script>

<div class="action-log">
  {#each actions as action}
    <div class="action-entry">
      <div class="action-timeline-dot" class:agent={action.actor === 'agent'}></div>
      <div class="action-content">
        <div class="action-header">
          <span class="action-name">{action.action}</span>
          <div class="action-badges">
            <Badge variant={action.actor === 'human' ? 'info' : 'accent'}>
              {action.actor}
            </Badge>
            <Badge variant={action.status === 'completed' ? 'pass' : action.status === 'pending' ? 'degraded' : 'default'}>
              {action.status}
            </Badge>
          </div>
        </div>
        <p class="action-desc">{action.description}</p>
        <span class="action-time">{action.timestamp}</span>
      </div>
    </div>
  {/each}

  {#if actions.length === 0}
    <div class="empty-log">No agent actions recorded yet</div>
  {/if}
</div>

<style>
  .action-log {
    display: flex;
    flex-direction: column;
    gap: 0;
    position: relative;
    padding-left: var(--space-lg);
  }

  .action-log::before {
    content: '';
    position: absolute;
    left: 11px;
    top: 16px;
    bottom: 16px;
    width: 2px;
    background: var(--border-subtle);
  }

  .action-entry {
    display: flex;
    gap: var(--space-md);
    padding: var(--space-md) 0;
    position: relative;
  }

  .action-timeline-dot {
    flex-shrink: 0;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--info);
    border: 2px solid var(--bg-primary);
    margin-top: 4px;
    z-index: 1;
  }

  .action-timeline-dot.agent {
    background: var(--accent);
  }

  .action-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-md);
  }

  .action-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-sm);
  }

  .action-name {
    font-weight: 600;
    font-size: 0.875rem;
    text-transform: capitalize;
    color: var(--text-primary);
  }

  .action-badges {
    display: flex;
    gap: var(--space-xs);
  }

  .action-desc {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .action-time {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
  }

  .empty-log {
    text-align: center;
    padding: var(--space-2xl);
    color: var(--text-tertiary);
    font-style: italic;
  }
</style>
