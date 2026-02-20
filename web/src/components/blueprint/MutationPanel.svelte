<script lang="ts">
  import type { MutationResponse } from '../../lib/api/types';
  import PillButton from '../shared/PillButton.svelte';
  import Badge from '../shared/Badge.svelte';

  interface Props {
    appId: string;
    onmutate: (cls: string, path: string, value: string) => void;
    results?: MutationResponse[];
    loading?: boolean;
  }

  let { appId, onmutate, results = [], loading = false }: Props = $props();

  let mutationClass = $state('set_name');
  let path = $state('');
  let value = $state('');

  const classes = ['set_name', 'set_plan', 'set_region', 'set_feature_flag'];

  function submit() {
    if (path && value) {
      onmutate(mutationClass, path, value);
    }
  }
</script>

<div class="mutation-panel">
  <h3 class="panel-title">Mutations</h3>

  <div class="mutation-form">
    <select class="field-input" bind:value={mutationClass}>
      {#each classes as cls}
        <option value={cls}>{cls}</option>
      {/each}
    </select>

    <input
      type="text"
      class="field-input"
      placeholder="Path (e.g. /name)"
      bind:value={path}
    />

    <input
      type="text"
      class="field-input"
      placeholder="Value"
      bind:value={value}
    />

    <PillButton variant="primary" size="sm" onclick={submit} disabled={loading || !path || !value}>
      {loading ? 'Applying...' : 'Apply Mutation'}
    </PillButton>
  </div>

  {#if results.length > 0}
    <div class="mutation-results">
      {#each results as result}
        <div class="mutation-result">
          <div class="result-header">
            <code>{result.mutation_class}</code>
            <Badge variant={result.policy_check === 'pass' ? 'pass' : 'fail'}>
              {result.policy_check}
            </Badge>
          </div>
          <div class="result-detail">
            <span class="detail-path">{result.path}</span>
            <span class="detail-arrow">→</span>
            <span class="detail-value">{JSON.stringify(result.value)}</span>
          </div>
          <span class="result-version">v{result.version}</span>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .mutation-panel {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .panel-title {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 600;
  }

  .mutation-form {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
    align-items: center;
  }

  .field-input {
    padding: 8px 12px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    font-size: 0.8125rem;
    color: var(--text-primary);
    font-family: var(--font-code);
  }
  .field-input:focus {
    border-color: var(--accent);
  }

  select.field-input {
    cursor: pointer;
    appearance: auto;
  }

  .mutation-results {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .mutation-result {
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

  .result-header code {
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--accent);
  }

  .result-detail {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-family: var(--font-code);
    font-size: 0.8125rem;
  }

  .detail-path { color: var(--text-secondary); }
  .detail-arrow { color: var(--text-tertiary); }
  .detail-value { color: var(--text-primary); }

  .result-version {
    font-family: var(--font-code);
    font-size: 0.6875rem;
    color: var(--text-tertiary);
  }
</style>
