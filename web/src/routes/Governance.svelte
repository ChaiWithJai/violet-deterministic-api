<script lang="ts">
  import { onMount } from 'svelte';
  import AgentActionLog from '../components/governance/AgentActionLog.svelte';
  import Accordion from '../components/shared/Accordion.svelte';
  import Badge from '../components/shared/Badge.svelte';
  import { getTools } from '../lib/api/endpoints';
  import type { Tool } from '../lib/api/types';

  // Mock action log — in production, this would come from an API endpoint
  let actions = $state<{
    id: string;
    action: string;
    actor: 'human' | 'agent';
    description: string;
    timestamp: string;
    status: string;
  }[]>([]);

  let tools = $state<Tool[]>([]);
  let error = $state<string | null>(null);

  onMount(() => {
    loadTools();
  });

  async function loadTools() {
    const res = await getTools();
    if (res.ok) {
      tools = (res.data as any).tools || [];
    }
  }
</script>

<div class="governance-page">
  <div class="page-header">
    <h1 class="page-title">Governance</h1>
    <p class="page-desc">Monitor AI agent actions, enforce human approval gates, and audit the plan → act → verify → deploy pipeline.</p>
  </div>

  <!-- Agent Action Timeline -->
  <section class="section">
    <h2 class="section-title">Agent Action Log</h2>
    <AgentActionLog {actions} />
  </section>

  <!-- Tools Catalog -->
  <section class="section">
    <h2 class="section-title">Tools Catalog</h2>
    <div class="tools-grid">
      {#each tools as tool}
        <Accordion title={tool.name}>
          <div class="tool-detail">
            <p class="tool-desc">{tool.description}</p>
            <Badge variant="default">{tool.category}</Badge>
          </div>
        </Accordion>
      {:else}
        <p class="empty-text">No tools available</p>
      {/each}
    </div>
  </section>
</div>

<style>
  .governance-page {
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

  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .section-title {
    font-family: var(--font-display);
    font-size: 1.25rem;
    font-weight: 600;
    letter-spacing: -0.02em;
  }

  .tools-grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    max-width: 640px;
  }

  .tool-detail {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .tool-desc {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .empty-text {
    color: var(--text-tertiary);
    font-style: italic;
  }
</style>
