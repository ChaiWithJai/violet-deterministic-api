<script lang="ts">
  import type { AgentPlanResponse, StudioJobRequest } from '../../lib/api/types';
  import PillButton from '../shared/PillButton.svelte';

  interface Props {
    plan: AgentPlanResponse;
    ongenerate: (req: StudioJobRequest) => void;
    loading?: boolean;
  }

  let { plan, ongenerate, loading = false }: Props = $props();

  let appName = $state('');
  let domain = $state('');
  let template = $state('violet-rails-extension');

  function handleGenerate() {
    ongenerate({
      prompt: plan.steps.map((s) => s.description).join('\n'),
      app_name: appName || undefined,
      domain: domain || undefined,
      template,
    });
  }
</script>

<div class="confirm-form">
  <h2 class="confirm-title">Scope Confirmation</h2>

  <div class="steps-list">
    {#each plan.steps as step}
      <div class="step-card">
        <span class="step-num">{step.step}</span>
        <div class="step-content">
          <span class="step-action">{step.action}</span>
          <p class="step-desc">{step.description}</p>
        </div>
      </div>
    {/each}
  </div>

  <div class="form-fields">
    <label class="field">
      <span class="field-label">App Name (optional)</span>
      <input type="text" class="field-input" placeholder="my-saas-app" bind:value={appName} />
    </label>

    <label class="field">
      <span class="field-label">Domain (optional)</span>
      <input type="text" class="field-input" placeholder="app.example.com" bind:value={domain} />
    </label>

    <label class="field">
      <span class="field-label">Template</span>
      <select class="field-input" bind:value={template}>
        <option value="violet-rails-extension">Violet Rails Extension</option>
        <option value="standalone">Standalone</option>
        <option value="api-only">API Only</option>
      </select>
    </label>
  </div>

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

  .steps-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .step-card {
    display: flex;
    align-items: flex-start;
    gap: var(--space-md);
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
  }

  .step-num {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--accent-subtle);
    color: var(--accent);
    border-radius: 50%;
    font-family: var(--font-code);
    font-size: 0.75rem;
    font-weight: 700;
  }

  .step-content {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .step-action {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--accent);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .step-desc {
    font-size: 0.875rem;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .form-fields {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .field-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .field-input {
    width: 100%;
    padding: 10px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-size: 0.8125rem;
    color: var(--text-primary);
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  .field-input:focus {
    border-color: var(--accent);
  }

  select.field-input {
    cursor: pointer;
    appearance: auto;
  }
</style>
