<script lang="ts">
  import DecisionCard from '../components/decisions/DecisionCard.svelte';
  import ReplayPanel from '../components/decisions/ReplayPanel.svelte';
  import StatCounter from '../components/shared/StatCounter.svelte';
  import PillButton from '../components/shared/PillButton.svelte';
  import { decisionStore } from '../lib/stores/decision.svelte';
  import { createDecision, replayDecision } from '../lib/api/endpoints';
  import type { DecisionRequest, Candidate } from '../lib/api/types';

  let userId = $state('user-1');
  let contextKeys = $state('page=home');
  let candidatesText = $state('item-a,item-b,item-c');
  let error = $state<string | null>(null);

  function parseCandidates(text: string): Candidate[] {
    return text
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .map((id) => ({ item_id: id }));
  }

  async function handleCreate() {
    decisionStore.setLoading(true);
    error = null;

    const ctx: Record<string, string> = {};
    for (const pair of contextKeys.split(',')) {
      const [k, v] = pair.split('=').map((s) => s.trim());
      if (k && v) ctx[k] = v;
    }

    const req: DecisionRequest = {
      user_id: userId,
      context_keys: ctx,
      candidates: parseCandidates(candidatesText),
    };

    const res = await createDecision(req);
    decisionStore.setLoading(false);

    if (!res.ok) {
      error = (res.data as any).error || 'Decision failed';
      return;
    }

    decisionStore.addDecision(res.data);
  }

  async function handleReplay(decisionId: string) {
    decisionStore.setLoading(true);
    error = null;

    const res = await replayDecision({ decision_id: decisionId });
    decisionStore.setLoading(false);

    if (!res.ok) {
      error = (res.data as any).error || 'Replay failed';
      return;
    }

    decisionStore.setReplay(res.data);
  }
</script>

<div class="decisions-page">
  <div class="page-header">
    <h1 class="page-title">Decisions</h1>
    <p class="page-desc">Create deterministic decisions, inspect the pipeline, and verify replay consistency.</p>
  </div>

  <!-- Stats -->
  <div class="stats-row">
    <StatCounter label="Decisions" value={decisionStore.decisions.length} />
    <StatCounter label="Policy Ver" value="v1" />
    <StatCounter label="Data Ver" value="v1" />
  </div>

  <!-- Create Form -->
  <div class="create-form">
    <h3 class="form-title">Create Decision</h3>
    <div class="form-fields">
      <label class="field">
        <span class="field-label">User ID</span>
        <input type="text" class="field-input" bind:value={userId} />
      </label>
      <label class="field">
        <span class="field-label">Context Keys (key=value, ...)</span>
        <input type="text" class="field-input" bind:value={contextKeys} />
      </label>
      <label class="field">
        <span class="field-label">Candidates (comma-separated item IDs)</span>
        <input type="text" class="field-input" bind:value={candidatesText} />
      </label>
    </div>
    <PillButton onclick={handleCreate} disabled={decisionStore.loading}>
      {decisionStore.loading ? 'Creating...' : 'Create Decision'}
    </PillButton>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <!-- Replay Panel -->
  {#if decisionStore.activeReplay}
    <ReplayPanel replay={decisionStore.activeReplay} />
  {/if}

  <!-- Decision Cards -->
  <div class="decisions-list">
    {#each decisionStore.decisions as decision}
      <DecisionCard {decision} onreplay={handleReplay} />
    {/each}
  </div>
</div>

<style>
  .decisions-page {
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

  .page-desc {
    font-size: 0.875rem;
    color: var(--text-secondary);
    max-width: 600px;
  }

  .stats-row {
    display: flex;
    gap: var(--space-xl);
    padding: var(--space-md);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    width: fit-content;
  }

  .create-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    padding: var(--space-lg);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    max-width: 640px;
  }

  .form-title {
    font-family: var(--font-display);
    font-size: 1rem;
    font-weight: 600;
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
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-family: var(--font-code);
    font-size: 0.8125rem;
    color: var(--text-primary);
  }
  .field-input:focus {
    border-color: var(--accent);
  }

  .error-banner {
    padding: var(--space-md);
    background: var(--fail-subtle);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: var(--fail);
    font-size: 0.875rem;
  }

  .decisions-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }
</style>
