<script lang="ts">
  import { onMount } from 'svelte';
  import MutationPanel from '../components/blueprint/MutationPanel.svelte';
  import DeployIntent from '../components/blueprint/DeployIntent.svelte';
  import CodeBlock from '../components/shared/CodeBlock.svelte';
  import Badge from '../components/shared/Badge.svelte';
  import { appStore } from '../lib/stores/app.svelte';
  import {
    getApp,
    createMutation,
    verifyApp,
    deploySelfHost,
    deployManaged,
  } from '../lib/api/endpoints';
  import type { MutationResponse, VerifyResponse, DeployIntentResponse } from '../lib/api/types';

  interface Props {
    appId: string;
  }

  let { appId }: Props = $props();

  let mutations = $state<MutationResponse[]>([]);
  let verifyResult = $state<VerifyResponse | null>(null);
  let deployResult = $state<DeployIntentResponse | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);

  let verified = $derived(verifyResult?.status === 'pass');

  onMount(() => {
    loadApp();
  });

  async function loadApp() {
    appStore.setLoading(true);
    const res = await getApp(appId);
    appStore.setLoading(false);
    if (res.ok) {
      appStore.setActive(res.data);
    } else {
      error = (res.data as any).error || 'Failed to load app';
    }
  }

  async function handleMutate(cls: string, path: string, value: string) {
    loading = true;
    error = null;
    let parsedValue: unknown = value;
    try { parsedValue = JSON.parse(value); } catch { /* keep as string */ }

    const res = await createMutation(appId, {
      mutation_class: cls,
      path,
      value: parsedValue,
    });
    loading = false;

    if (res.ok) {
      mutations = [...mutations, res.data];
      loadApp();
    } else {
      error = (res.data as any).error || 'Mutation failed';
    }
  }

  async function handleVerify() {
    loading = true;
    error = null;
    const res = await verifyApp(appId);
    loading = false;
    if (res.ok) verifyResult = res.data;
    else error = (res.data as any).error || 'Verify failed';
  }

  async function handleDeploy(type: 'self-host' | 'managed') {
    loading = true;
    error = null;
    const fn = type === 'self-host' ? deploySelfHost : deployManaged;
    const res = await fn(appId);
    loading = false;
    if (res.ok) deployResult = res.data;
    else error = (res.data as any).error || 'Deploy failed';
  }
</script>

<div class="blueprint-page">
  <div class="page-header">
    <h1 class="page-title">Blueprint</h1>
    {#if appStore.active}
      <div class="app-meta">
        <Badge variant="accent">{appStore.active.name}</Badge>
        <Badge variant="default">v{appStore.active.version}</Badge>
        <Badge variant="default">{appStore.active.region}</Badge>
      </div>
    {/if}
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if appStore.active}
    <!-- Blueprint JSON -->
    <CodeBlock
      code={JSON.stringify(appStore.active.blueprint, null, 2)}
      lang="json"
      maxHeight="300px"
    />

    <!-- Mutations -->
    <MutationPanel {appId} onmutate={handleMutate} results={mutations} {loading} />

    <!-- Verify Gate -->
    <div class="verify-section">
      <button class="verify-btn" onclick={handleVerify} disabled={loading}>
        {loading ? 'Verifying...' : 'Run Verification'}
      </button>
      {#if verifyResult}
        <Badge variant={verified ? 'pass' : 'fail'}>{verifyResult.status}</Badge>
        <div class="verify-checks">
          {#each verifyResult.checks as check}
            <span class="check-item" class:pass={check.status === 'pass'}>
              {check.status === 'pass' ? '✓' : '✗'} {check.name}
            </span>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Deploy -->
    <DeployIntent
      {appId}
      {verified}
      ondeploy={handleDeploy}
      result={deployResult}
      {loading}
    />
  {:else if appStore.loading}
    <div class="loading-state">Loading app...</div>
  {:else}
    <div class="empty-state">App not found</div>
  {/if}
</div>

<style>
  .blueprint-page {
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
  }

  .page-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .page-title {
    font-family: var(--font-display);
    font-size: 2rem;
    font-weight: 700;
    letter-spacing: -0.03em;
  }

  .app-meta {
    display: flex;
    gap: var(--space-sm);
  }

  .error-banner {
    padding: var(--space-md);
    background: var(--fail-subtle);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: var(--fail);
    font-size: 0.875rem;
  }

  .verify-section {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-md);
  }

  .verify-btn {
    padding: 10px 24px;
    border-radius: var(--radius-pill);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    transition: all var(--duration-fast) var(--ease-out);
  }
  .verify-btn:hover:not(:disabled) {
    border-color: var(--accent);
    color: var(--accent);
  }
  .verify-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .verify-checks {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-sm);
  }

  .check-item {
    font-size: 0.8125rem;
    color: var(--fail);
  }
  .check-item.pass {
    color: var(--pass);
  }

  .loading-state, .empty-state {
    text-align: center;
    padding: var(--space-3xl);
    color: var(--text-tertiary);
    font-style: italic;
  }
</style>
