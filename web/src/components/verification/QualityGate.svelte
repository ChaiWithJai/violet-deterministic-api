<script lang="ts">
  import Badge from '../shared/Badge.svelte';
  import PillButton from '../shared/PillButton.svelte';

  interface Props {
    onrun: (target: string) => void;
    results?: { target: string; status: string; checks: { name: string; status: string; message?: string }[] }[];
    loading?: boolean;
  }

  let { onrun, results = [], loading = false }: Props = $props();

  const targets = ['web', 'mobile', 'api', 'verify', 'all'];
</script>

<div class="quality-gate">
  <h3 class="gate-title">Quality Gate</h3>

  <div class="gate-targets">
    {#each targets as target}
      <PillButton
        variant="secondary"
        size="sm"
        onclick={() => onrun(target)}
        disabled={loading}
      >
        Run {target}
      </PillButton>
    {/each}
  </div>

  {#if results.length > 0}
    <div class="gate-results">
      {#each results as result}
        <div class="gate-result">
          <div class="result-header">
            <span class="result-target">{result.target}</span>
            <Badge variant={result.status === 'pass' ? 'pass' : 'fail'}>
              {result.status}
            </Badge>
          </div>
          {#each result.checks as check}
            <div class="result-check">
              <span class="check-icon">{check.status === 'pass' ? '✓' : '✗'}</span>
              <span class="check-name">{check.name}</span>
              {#if check.message}
                <span class="check-msg">{check.message}</span>
              {/if}
            </div>
          {/each}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .quality-gate {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .gate-title {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .gate-targets {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }

  .gate-results {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .gate-result {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .result-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .result-target {
    font-weight: 600;
    font-size: 0.875rem;
    text-transform: capitalize;
  }

  .result-check {
    display: flex;
    align-items: baseline;
    gap: var(--space-sm);
    font-size: 0.8125rem;
  }

  .check-icon {
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .check-name {
    color: var(--text-secondary);
  }

  .check-msg {
    color: var(--text-tertiary);
    font-family: var(--font-code);
    font-size: 0.75rem;
  }
</style>
