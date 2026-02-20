<script lang="ts">
  import type { StudioEvent } from '../../lib/api/types';

  interface Props {
    events: StudioEvent[];
    jobId: string;
  }

  let { events, jobId }: Props = $props();

  let consoleEvents = $derived(
    events.filter((e) => e.type === 'console' || e.type === 'log' || e.message)
  );

  let terminalInput = $state('');

  async function sendCommand() {
    if (!terminalInput.trim()) return;
    // Terminal command would POST to /v1/studio/jobs/{id}/terminal
    terminalInput = '';
  }
</script>

<div class="terminal-panel">
  <div class="terminal-tabs">
    <span class="terminal-tab active">Console</span>
  </div>

  <div class="terminal-output">
    {#each consoleEvents as event}
      <div class="terminal-line">
        <span class="terminal-prefix">&gt;</span>
        <span class="terminal-text">{event.message || JSON.stringify(event.data)}</span>
      </div>
    {/each}
    {#if consoleEvents.length === 0}
      <div class="terminal-empty">Waiting for output...</div>
    {/if}
  </div>

  <div class="terminal-input-bar">
    <span class="terminal-prompt">$</span>
    <input
      type="text"
      class="terminal-input"
      placeholder="Enter command..."
      bind:value={terminalInput}
      onkeydown={(e) => e.key === 'Enter' && sendCommand()}
    />
  </div>
</div>

<style>
  .terminal-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .terminal-tabs {
    display: flex;
    padding: var(--space-xs);
    border-bottom: 1px solid var(--border-subtle);
  }

  .terminal-tab {
    padding: 4px 12px;
    font-size: 0.6875rem;
    font-family: var(--font-code);
    color: var(--text-tertiary);
    border-radius: var(--radius-sm);
  }
  .terminal-tab.active {
    background: var(--bg-surface);
    color: var(--text-secondary);
  }

  .terminal-output {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-sm);
    font-family: var(--font-code);
    font-size: 0.75rem;
    line-height: 1.7;
  }

  .terminal-line {
    display: flex;
    gap: var(--space-sm);
  }

  .terminal-prefix {
    color: var(--accent);
    flex-shrink: 0;
  }

  .terminal-text {
    color: var(--text-secondary);
    word-break: break-all;
  }

  .terminal-empty {
    color: var(--text-tertiary);
    font-style: italic;
  }

  .terminal-input-bar {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    border-top: 1px solid var(--border-subtle);
    background: rgba(0, 0, 0, 0.3);
  }

  .terminal-prompt {
    font-family: var(--font-code);
    font-size: 0.75rem;
    color: var(--pass);
  }

  .terminal-input {
    flex: 1;
    font-family: var(--font-code);
    font-size: 0.75rem;
    color: var(--text-primary);
  }
  .terminal-input::placeholder {
    color: var(--text-tertiary);
  }
</style>
