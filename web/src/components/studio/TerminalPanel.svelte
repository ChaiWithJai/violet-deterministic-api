<script lang="ts">
  import { sendTerminalCommand } from '../../lib/api/endpoints';

  interface Props {
    terminalLogs: string[];
    consoleLogs: string[];
    jobId: string;
  }

  let { terminalLogs, consoleLogs, jobId }: Props = $props();

  let activeTab = $state<'terminal' | 'console'>('terminal');
  let terminalInput = $state('');
  let localOutput = $state<string[]>([]);

  async function handleCommand() {
    const cmd = terminalInput.trim();
    if (!cmd) return;
    terminalInput = '';
    localOutput = [...localOutput, `$ ${cmd}`];

    const res = await sendTerminalCommand(jobId, cmd);
    if (res.ok) {
      localOutput = [...localOutput, ...res.data.output];
    } else {
      localOutput = [...localOutput, `Error: command failed`];
    }
  }

  let allTerminalLines = $derived([...terminalLogs, ...localOutput]);
</script>

<div class="terminal-panel">
  <div class="terminal-tabs">
    <button class="terminal-tab" class:active={activeTab === 'terminal'} onclick={() => activeTab = 'terminal'}>Terminal</button>
    <button class="terminal-tab" class:active={activeTab === 'console'} onclick={() => activeTab = 'console'}>Console</button>
  </div>

  <div class="terminal-output">
    {#if activeTab === 'terminal'}
      {#each allTerminalLines as line}
        <div class="terminal-line">
          <span class="terminal-text">{line}</span>
        </div>
      {/each}
      {#if allTerminalLines.length === 0}
        <div class="terminal-empty">Waiting for output...</div>
      {/if}
    {:else}
      {#each consoleLogs as line}
        <div class="terminal-line">
          <span class="terminal-text">{line}</span>
        </div>
      {/each}
      {#if consoleLogs.length === 0}
        <div class="terminal-empty">No console logs</div>
      {/if}
    {/if}
  </div>

  {#if activeTab === 'terminal'}
    <div class="terminal-input-bar">
      <span class="terminal-prompt">$</span>
      <input
        type="text"
        class="terminal-input"
        placeholder="Enter command..."
        bind:value={terminalInput}
        onkeydown={(e) => e.key === 'Enter' && handleCommand()}
      />
    </div>
  {/if}
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
    gap: 2px;
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
