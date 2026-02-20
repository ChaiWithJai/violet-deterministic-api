<script lang="ts">
  import type { VerificationResult } from '../../lib/api/types';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    results: VerificationResult[];
  }

  let { results }: Props = $props();
</script>

<div class="verification-report">
  {#each results as result}
    <div class="report-item">
      <div class="report-header">
        <span class="report-target">{result.target}</span>
        <Badge variant={result.status === 'pass' ? 'pass' : result.status === 'fail' ? 'fail' : 'degraded'}>
          {result.status}
        </Badge>
      </div>
      <div class="report-checks">
        {#each result.checks as check}
          <div class="check-row">
            <span class="check-status" class:pass={check.status === 'pass'} class:fail={check.status !== 'pass'}>
              {check.status === 'pass' ? '✓' : '✗'}
            </span>
            <span class="check-name">{check.name}</span>
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .verification-report {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .report-item {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: var(--space-md);
  }

  .report-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-sm);
  }

  .report-target {
    font-weight: 600;
    font-size: 0.875rem;
    text-transform: capitalize;
  }

  .report-checks {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .check-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-size: 0.8125rem;
  }

  .check-status {
    font-size: 0.75rem;
    font-weight: 700;
  }
  .check-status.pass { color: var(--pass); }
  .check-status.fail { color: var(--fail); }

  .check-name {
    color: var(--text-secondary);
  }
</style>
