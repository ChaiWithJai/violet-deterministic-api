<script lang="ts">
  import type { DeployIntentResponse } from '../../lib/api/types';
  import PillButton from '../shared/PillButton.svelte';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    appId: string;
    verified: boolean;
    ondeploy: (type: 'self-host' | 'managed') => void;
    result?: DeployIntentResponse | null;
    loading?: boolean;
  }

  let { appId, verified, ondeploy, result = null, loading = false }: Props = $props();
</script>

<div class="deploy-intent">
  <h3 class="deploy-title">Deploy Intent</h3>

  {#if !verified}
    <div class="deploy-gate">
      <span class="gate-icon">🔒</span>
      <p class="gate-text">Verification must pass before deploying</p>
    </div>
  {:else}
    <div class="deploy-options">
      <button class="deploy-card" onclick={() => ondeploy('self-host')} disabled={loading}>
        <span class="card-icon">🖥️</span>
        <span class="card-label">Self-Host</span>
        <span class="card-desc">Deploy to your infrastructure</span>
      </button>

      <button class="deploy-card" onclick={() => ondeploy('managed')} disabled={loading}>
        <span class="card-icon">☁️</span>
        <span class="card-label">Managed</span>
        <span class="card-desc">Hosted by Violet</span>
      </button>
    </div>
  {/if}

  {#if result}
    <div class="deploy-result">
      <div class="result-row">
        <span>Intent ID:</span>
        <code>{result.intent_id}</code>
      </div>
      <div class="result-row">
        <span>Status:</span>
        <Badge variant={result.status === 'pending_approval' ? 'degraded' : 'pass'}>
          {result.status}
        </Badge>
      </div>
      <div class="result-row">
        <span>Approval Required:</span>
        <Badge variant={result.approval_required ? 'degraded' : 'default'}>
          {result.approval_required ? 'Yes' : 'No'}
        </Badge>
      </div>
      <div class="result-row">
        <span>Target:</span>
        <code>{result.target}</code>
      </div>
    </div>
  {/if}
</div>

<style>
  .deploy-intent {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .deploy-title {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 600;
  }

  .deploy-gate {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    padding: var(--space-lg);
    background: var(--degraded-subtle);
    border: 1px solid rgba(245, 158, 11, 0.2);
    border-radius: var(--radius-md);
  }

  .gate-icon { font-size: 1.5rem; }
  .gate-text {
    font-size: 0.875rem;
    color: var(--degraded);
  }

  .deploy-options {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-md);
  }

  .deploy-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-xl);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    text-align: center;
    transition: all var(--duration-fast) var(--ease-out);
  }

  .deploy-card:hover:not(:disabled) {
    border-color: var(--accent);
    background: var(--accent-subtle);
  }

  .deploy-card:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .card-icon { font-size: 2rem; }
  .card-label {
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--text-primary);
  }
  .card-desc {
    font-size: 0.75rem;
    color: var(--text-tertiary);
  }

  .deploy-result {
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .result-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .result-row code {
    font-family: var(--font-code);
    color: var(--text-primary);
  }
</style>
