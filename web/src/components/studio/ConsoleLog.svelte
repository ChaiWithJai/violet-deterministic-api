<script lang="ts">
  import type { StudioEvent } from '../../lib/api/types';

  interface Props {
    events: StudioEvent[];
  }

  let { events }: Props = $props();
</script>

<div class="console-log">
  {#each events as event, i}
    <div class="log-entry">
      <span class="log-idx">{i + 1}</span>
      <span class="log-type">{event.type}</span>
      {#if event.phase}<span class="log-phase">{event.phase}</span>{/if}
      {#if event.message}<span class="log-msg">{event.message}</span>{/if}
    </div>
  {/each}
  {#if events.length === 0}
    <div class="log-empty">No events yet</div>
  {/if}
</div>

<style>
  .console-log {
    font-family: var(--font-code);
    font-size: 0.75rem;
    padding: var(--space-sm);
    overflow-y: auto;
  }

  .log-entry {
    display: flex;
    align-items: baseline;
    gap: var(--space-sm);
    padding: 2px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.02);
  }

  .log-idx {
    color: var(--text-tertiary);
    width: 24px;
    text-align: right;
    flex-shrink: 0;
  }

  .log-type {
    color: var(--accent);
    font-weight: 500;
    flex-shrink: 0;
  }

  .log-phase {
    color: var(--info);
    flex-shrink: 0;
  }

  .log-msg {
    color: var(--text-secondary);
  }

  .log-empty {
    color: var(--text-tertiary);
    font-style: italic;
    padding: var(--space-md);
  }
</style>
