<script lang="ts">
  import type { Confirmation, StudioJobRequest } from '../../lib/api/types';
  import PillButton from '../shared/PillButton.svelte';

  interface Props {
    confirmation: Partial<Confirmation>;
    checks: string[];
    ongenerate: (req: StudioJobRequest) => void;
    loading?: boolean;
  }

  let { confirmation, checks, ongenerate, loading = false }: Props = $props();

  function handleGenerate() {
    ongenerate({
      prompt: confirmation.prompt ?? '',
      app_name: confirmation.app_name,
      domain: confirmation.domain,
      template: confirmation.template,
      source_system: confirmation.source_system,
      primary_users: confirmation.primary_users,
      core_workflows: confirmation.core_workflows,
      data_entities: confirmation.data_entities,
      deployment_target: confirmation.deployment_target,
      region: confirmation.region,
      generation_depth: confirmation.generation_depth,
    });
  }

  const fields: { key: keyof Confirmation; label: string }[] = [
    { key: 'app_name', label: 'App Name' },
    { key: 'domain', label: 'Domain' },
    { key: 'template', label: 'Template' },
    { key: 'source_system', label: 'Source System' },
    { key: 'deployment_target', label: 'Deployment Target' },
    { key: 'region', label: 'Region' },
    { key: 'plan', label: 'Plan' },
    { key: 'generation_depth', label: 'Generation Depth' },
  ];

  const arrayFields: { key: keyof Confirmation; label: string }[] = [
    { key: 'primary_users', label: 'Primary Users' },
    { key: 'core_workflows', label: 'Core Workflows' },
    { key: 'data_entities', label: 'Data Entities' },
    { key: 'integrations', label: 'Integrations' },
    { key: 'constraints', label: 'Constraints' },
  ];
</script>

<div class="confirm-form">
  <h2 class="confirm-title">Scope Confirmation</h2>

  {#if checks.length > 0}
    <div class="checks-list">
      {#each checks as check}
        <div class="check-item">
          <span class="check-icon">&#10003;</span>
          <span class="check-text">{check}</span>
        </div>
      {/each}
    </div>
  {/if}

  <div class="confirmation-grid">
    {#each fields as { key, label }}
      {@const val = confirmation[key]}
      {#if val && typeof val === 'string'}
        <div class="conf-field">
          <span class="conf-label">{label}</span>
          <span class="conf-value">{val}</span>
        </div>
      {/if}
    {/each}

    {#each arrayFields as { key, label }}
      {@const arr = confirmation[key]}
      {#if Array.isArray(arr) && arr.length > 0}
        <div class="conf-field">
          <span class="conf-label">{label}</span>
          <div class="conf-pills">
            {#each arr as item}
              <span class="conf-pill">{item}</span>
            {/each}
          </div>
        </div>
      {/if}
    {/each}
  </div>

  {#if confirmation.prompt}
    <div class="prompt-preview">
      <span class="conf-label">Original Prompt</span>
      <p class="prompt-text">{confirmation.prompt}</p>
    </div>
  {/if}

  <PillButton size="lg" onclick={handleGenerate} disabled={loading}>
    {loading ? 'Creating Job...' : 'Generate'}
  </PillButton>
</div>

<style>
  .confirm-form {
    max-width: 720px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .confirm-title {
    font-family: var(--font-display);
    font-size: 1.25rem;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--text-primary);
  }

  .checks-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: var(--space-md);
    background: var(--pass-subtle);
    border: 1px solid rgba(34, 197, 94, 0.15);
    border-radius: var(--radius-md);
  }

  .check-item {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-size: 0.8125rem;
  }

  .check-icon {
    color: var(--pass);
    font-weight: 700;
    font-size: 0.75rem;
  }

  .check-text {
    color: var(--text-secondary);
  }

  .confirmation-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-md);
  }

  @media (max-width: 640px) {
    .confirmation-grid {
      grid-template-columns: 1fr;
    }
  }

  .conf-field {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
  }

  .conf-label {
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-tertiary);
  }

  .conf-value {
    font-size: 0.875rem;
    color: var(--text-primary);
  }

  .conf-pills {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-xs);
  }

  .conf-pill {
    padding: 2px 10px;
    border-radius: var(--radius-pill);
    background: var(--accent-subtle);
    font-size: 0.75rem;
    color: var(--accent);
    font-weight: 500;
  }

  .prompt-preview {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
  }

  .prompt-text {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    line-height: 1.6;
    font-style: italic;
  }
</style>
