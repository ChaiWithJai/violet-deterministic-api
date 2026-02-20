<script lang="ts">
  interface Props {
    logs: string[];
  }

  let { logs }: Props = $props();
</script>

<div class="console-log">
  {#each logs as line, i}
    <div class="log-entry">
      <span class="log-idx">{i + 1}</span>
      <span class="log-msg">{line}</span>
    </div>
  {/each}
  {#if logs.length === 0}
    <div class="log-empty">No logs yet</div>
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

  .log-msg {
    color: var(--text-secondary);
  }

  .log-empty {
    color: var(--text-tertiary);
    font-style: italic;
    padding: var(--space-md);
  }
</style>
