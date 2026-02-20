<script lang="ts">
  interface Props {
    onsubmit: (prompt: string) => void;
    loading?: boolean;
  }

  let { onsubmit, loading = false }: Props = $props();

  let prompt = $state('');

  const signals = [
    { keyword: 'crm', label: 'CRM', icon: '👥' },
    { keyword: 'mobile', label: 'Mobile App', icon: '📱' },
    { keyword: 'agent', label: 'AI Agent', icon: '🤖' },
    { keyword: 'dashboard', label: 'Dashboard', icon: '📊' },
    { keyword: 'ecommerce', label: 'E-commerce', icon: '🛒' },
    { keyword: 'api', label: 'API Service', icon: '⚡' },
  ];

  let detectedSignals = $derived(
    signals.filter((s) => prompt.toLowerCase().includes(s.keyword))
  );

  function handleSubmit() {
    if (prompt.trim() && !loading) {
      onsubmit(prompt.trim());
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && e.metaKey) {
      handleSubmit();
    }
  }
</script>

<div class="prompt-area">
  <div class="prompt-container">
    <textarea
      class="prompt-input"
      placeholder="Describe the app you want to build..."
      bind:value={prompt}
      onkeydown={handleKeydown}
      rows="4"
    ></textarea>

    {#if detectedSignals.length > 0}
      <div class="signal-cards">
        {#each detectedSignals as signal}
          <span class="signal-card">
            <span class="signal-icon">{signal.icon}</span>
            {signal.label}
          </span>
        {/each}
      </div>
    {/if}

    <div class="prompt-footer">
      <span class="prompt-hint">⌘ Enter to submit</span>
      <button
        class="submit-btn"
        onclick={handleSubmit}
        disabled={!prompt.trim() || loading}
      >
        {loading ? 'Planning...' : 'Draft Scope'}
      </button>
    </div>
  </div>
</div>

<style>
  .prompt-area {
    position: relative;
    padding: var(--space-3xl) 0;
  }

  .prompt-area::before {
    content: '';
    position: absolute;
    inset: 0;
    background:
      radial-gradient(ellipse 80% 60% at 50% 40%, var(--accent-glow), transparent),
      radial-gradient(ellipse 60% 40% at 30% 60%, rgba(59, 130, 246, 0.08), transparent);
    pointer-events: none;
    z-index: 0;
  }

  .prompt-container {
    position: relative;
    z-index: 1;
    max-width: 720px;
    margin: 0 auto;
    background: var(--bg-glass);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .prompt-input {
    width: 100%;
    resize: none;
    font-family: var(--font-body);
    font-size: 1rem;
    line-height: 1.6;
    color: var(--text-primary);
    background: transparent;
  }
  .prompt-input::placeholder {
    color: var(--text-tertiary);
  }

  .signal-cards {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }

  .signal-card {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 12px;
    background: var(--accent-subtle);
    border: 1px solid var(--border-accent);
    border-radius: var(--radius-pill);
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--accent);
  }

  .signal-icon {
    font-size: 0.875rem;
  }

  .prompt-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .prompt-hint {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    font-family: var(--font-code);
  }

  .submit-btn {
    padding: 10px 24px;
    border-radius: var(--radius-pill);
    background: var(--accent);
    color: var(--text-on-accent);
    font-family: var(--font-body);
    font-size: 0.875rem;
    font-weight: 600;
    transition: all var(--duration-fast) var(--ease-out);
  }
  .submit-btn:hover:not(:disabled) {
    background: var(--accent-hover);
    box-shadow: var(--shadow-glow);
  }
  .submit-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
