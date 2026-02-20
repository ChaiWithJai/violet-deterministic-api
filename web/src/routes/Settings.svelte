<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../lib/stores/auth.svelte';
  import { getHealth, getLLMProviders } from '../lib/api/endpoints';
  import Badge from '../components/shared/Badge.svelte';
  import PillButton from '../components/shared/PillButton.svelte';
  import type { HealthResponse, LLMProvider } from '../lib/api/types';

  let health = $state<HealthResponse | null>(null);
  let providers = $state<LLMProvider[]>([]);
  let healthError = $state<string | null>(null);
  let token = $state(authStore.token);
  let baseUrl = $state(authStore.baseUrl);

  onMount(() => {
    checkHealth();
    loadProviders();
  });

  async function checkHealth() {
    healthError = null;
    const res = await getHealth();
    if (res.ok) health = res.data;
    else healthError = (res.data as any).error || 'Health check failed';
  }

  async function loadProviders() {
    const res = await getLLMProviders();
    if (res.ok) providers = (res.data as any).providers || [];
  }

  function saveAuth() {
    authStore.token = token;
    authStore.baseUrl = baseUrl;
    checkHealth();
    loadProviders();
  }

  function disconnect() {
    authStore.clear();
  }
</script>

<div class="settings-page">
  <div class="page-header">
    <h1 class="page-title">Settings</h1>
  </div>

  <!-- Connection -->
  <section class="section">
    <h2 class="section-title">Connection</h2>
    <div class="form-card">
      <label class="field">
        <span class="field-label">API Base URL</span>
        <input type="url" class="field-input" bind:value={baseUrl} />
      </label>
      <label class="field">
        <span class="field-label">Bearer Token</span>
        <input type="password" class="field-input" bind:value={token} />
      </label>
      <div class="btn-row">
        <PillButton onclick={saveAuth}>Save & Reconnect</PillButton>
        <PillButton variant="ghost" onclick={disconnect}>Disconnect</PillButton>
      </div>
    </div>
  </section>

  <!-- Health -->
  <section class="section">
    <h2 class="section-title">Health Check</h2>
    {#if health}
      <div class="health-card">
        <div class="health-row">
          <span>Status</span>
          <Badge variant="pass">{health.status}</Badge>
        </div>
        <div class="health-row">
          <span>Service</span>
          <code>{health.service}</code>
        </div>
        <div class="health-row">
          <span>Policy Version</span>
          <code>{health.policy_version}</code>
        </div>
        <div class="health-row">
          <span>Data Version</span>
          <code>{health.data_version}</code>
        </div>
        <div class="health-row">
          <span>Idempotency Cleanup</span>
          <code>{health.idempotency_cleanup_deleted_total}</code>
        </div>
      </div>
    {:else if healthError}
      <div class="error-banner">{healthError}</div>
    {:else}
      <PillButton variant="secondary" onclick={checkHealth}>Check Health</PillButton>
    {/if}
  </section>

  <!-- LLM Providers -->
  <section class="section">
    <h2 class="section-title">LLM Providers</h2>
    <div class="providers-grid">
      {#each providers as provider}
        <div class="provider-card">
          <div class="provider-header">
            <span class="provider-name">{provider.name}</span>
            <Badge variant={provider.status === 'available' ? 'pass' : 'default'}>
              {provider.status}
            </Badge>
          </div>
          <span class="provider-type">{provider.type}</span>
          <div class="provider-models">
            {#each provider.models as model}
              <Badge variant={model === provider.default_model ? 'accent' : 'default'}>
                {model}
              </Badge>
            {/each}
          </div>
        </div>
      {:else}
        <p class="empty-text">No providers configured</p>
      {/each}
    </div>
  </section>
</div>

<style>
  .settings-page {
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
    max-width: 640px;
  }

  .page-title {
    font-family: var(--font-display);
    font-size: 2rem;
    font-weight: 700;
    letter-spacing: -0.03em;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .section-title {
    font-family: var(--font-display);
    font-size: 1.125rem;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .form-card {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    padding: var(--space-lg);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
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
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--text-primary);
  }
  .field-input:focus { border-color: var(--accent); }

  .btn-row {
    display: flex;
    gap: var(--space-sm);
  }

  .health-card {
    padding: var(--space-lg);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .health-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .health-row code {
    font-family: var(--font-code);
    color: var(--text-primary);
  }

  .error-banner {
    padding: var(--space-md);
    background: var(--fail-subtle);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: var(--fail);
    font-size: 0.875rem;
  }

  .providers-grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .provider-card {
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .provider-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .provider-name {
    font-weight: 600;
    font-size: 0.875rem;
  }

  .provider-type {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    font-family: var(--font-code);
  }

  .provider-models {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-xs);
  }

  .empty-text {
    color: var(--text-tertiary);
    font-style: italic;
  }
</style>
